package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	apiextensions_v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

	staticpod sync.Map
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

	crd := &apiextensions_v1.CustomResourceDefinition{}

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
	podManifest := new(core_v1.Pod)
	if err := json.Unmarshal(data, podManifest); err != nil {
		return fmt.Errorf("unmarchal pod manifest: %w", err)
	}

	spod := staticpodTemplate
	spod.Name = podManifest.GetName()
	spod.Spec.Template = core_v1.PodTemplateSpec{
		ObjectMeta: podManifest.ObjectMeta,
		Spec:       podManifest.Spec,
	}

	resourceClient, err := newResourceClient(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	if _, err = resourceClient.Create(spod); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create staticpod %s : %w", podManifest.GetName(), err)
	}

	s.staticpod.Store(spod.Name, spod)
	return nil
}

func (s *client) CreateInformer() error {
	client, err := dynamic.NewForConfig(s.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Minute, core_v1.NamespaceAll, nil)
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
			typedObj, ok := newObj.(*unstructured.Unstructured)
			if !ok {
				zap.L().Error("convert object to unstructured.Unstructured")
				return
			}

			oldPod, isExist := s.staticpod.Load(typedObj.GetName())
			if !isExist {
				zap.L().Warn("not found staticpod", zap.String("name", typedObj.GetName()))
				return
			}

			specData, err := json.Marshal(typedObj.Object["spec"])
			if err != nil {
				zap.L().Error("marshal spec to json")
				return
			}

			var newSpec apps_v1.DeploymentSpec
			if err = json.Unmarshal(specData, &newSpec); err != nil {
				zap.L().Error("marshal json to spec")
				return
			}

			spod := oldPod.(*staticpod)
			spod.Spec = newSpec

			s.staticpod.Store(typedObj.GetName(), spod)
		},
	})
	go informer.Run(s.shutdownChan)

	return nil
}
