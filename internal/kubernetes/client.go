package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	restConfig   *rest.Config
	shutdownChan chan struct{}

	staticpod map[string]*staticpod
}

func NewClient(kubeconfigPath string) (*client, error) {
	configBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %s : %w", kubeconfigPath, err)
	}

	cli := &client{
		shutdownChan: make(chan struct{}),
		staticpod:    make(map[string]*staticpod),
	}

	cli.restConfig, err = clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("new rest config: %w\n", err)
	}
	return cli, nil
}

func (s *client) Close() {
	close(s.shutdownChan)
	runtime.HandleCrash()
}

func (s *client) CreateCRD() error {
	kubeClient, err := clientset.NewForConfig(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), staticpodCRD, meta_v1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create cdr in kubernetes: %w", err)
	}

	return nil
}

func (s *client) CreateStaticPod(data []byte) error {
	podManifest := new(v1.Pod)
	if err := json.Unmarshal(data, podManifest); err != nil {
		return fmt.Errorf("unmarchal pod manifest: %w", err)
	}

	staticpod := staticpodTemplate
	staticpod.Name = podManifest.GetName()
	staticpod.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: podManifest.ObjectMeta,
		Spec:       podManifest.Spec,
	}

	resourceClient, err := newResourceClient(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	if _, err = resourceClient.Create(staticpod); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create staticpod %s : %w", podManifest.GetName(), err)
	}

	s.staticpod[staticpod.Name] = staticpod
	return nil
}

func (s *client) CreateInformer() error {
	clientset, err := kubernetes.NewForConfig(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	watchlist := cache.NewListWatchFromClient(clientset.AppsV1().RESTClient(), "v1beta1.io.fraima.staticpod", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&staticpod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("service added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("service deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("service changed \n")
			},
		},
	)
	go controller.Run(s.shutdownChan)
	return nil
}
