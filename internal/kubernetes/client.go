package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
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

	crd := &apiextensionsv1.CustomResourceDefinition{}

	crdJSONData, _ := yaml.YAMLToJSON(staticpodCDR)

	if err := json.Unmarshal(crdJSONData, crd); err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd, meta_v1.CreateOptions{})
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
	client, err := dynamic.NewForConfig(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Minute, v1.NamespaceAll, nil)
	go factory.Start(s.shutdownChan)

	informer := factory.ForResource(
		schema.GroupVersionResource{
			Group:    group,
			Version:  versionName,
			Resource: plural,
		},
	).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			typedObj := oldObj.(*unstructured.Unstructured)
			bytes, _ := typedObj.MarshalJSON()
			fmt.Println(string(bytes))

			typedObj = newObj.(*unstructured.Unstructured)
			typedObj.GetAnnotations()
			bytes, _ = typedObj.MarshalJSON()
			fmt.Println(string(bytes))
		},
	})
	go informer.Run(s.shutdownChan)

	return nil
}
