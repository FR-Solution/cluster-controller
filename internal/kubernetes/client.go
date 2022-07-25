package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	restConfig   *rest.Config
	shutdownChan chan struct{}
}

func NewClient(kubeconfigPath string) (*client, error) {
	configBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %s : %w", kubeconfigPath, err)
	}

	cli := &client{
		shutdownChan: make(chan struct{}),
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

	data, _ := yaml.Marshal(staticpodCRD)
	fmt.Println(string(data))

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

	data, _ = yaml.Marshal(staticpod)
	fmt.Println(string(data))

	resourceClient, err := newResourceClient(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	if _, err = resourceClient.Create(staticpod); err != nil {
		return fmt.Errorf("create staticpod %s : %w", podManifest.GetName(), err)
	}
	return nil
}

func (s *client) CreateInformer() error {
	clusterClient, err := dynamic.NewForConfig(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new dinamic client: %w", err)
	}

	resource := schema.GroupVersionResource{
		Group:    group,
		Version:  versionName,
		Resource: plural,
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, namespace, nil)
	go factory.Start(s.shutdownChan)

	informer := factory.ForResource(resource).Informer()
	if !cache.WaitForCacheSync(s.shutdownChan, informer.HasSynced) {
		err = fmt.Errorf("Timed out waiting for caches to sync")
		runtime.HandleError(err)
		return err
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { fmt.Println("add") },
		UpdateFunc: func(interface{}, interface{}) { fmt.Println("update") },
		DeleteFunc: func(interface{}) { fmt.Println("delete") },
	})
	return nil
}
