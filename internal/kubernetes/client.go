package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	restConfig *rest.Config
}

func NewClient(kubeconfigPath string) (*Client, error) {
	configBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %s : %w", kubeconfigPath, err)
	}

	cli := &Client{}

	cli.restConfig, err = clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("new rest config: %w\n", err)
	}
	return cli, nil
}

func (c *Client) CreateCRD() error {
	kubeClient, err := clientset.NewForConfig(c.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), staticpodCRD, meta_v1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create cdr in kubernetes: %w", err)
	}

	return nil
}

func (c *Client) CreateStaticPod(data []byte) error {
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

	resourceClient, err := newResourceClient(c.restConfig)
	if err != nil {
		return fmt.Errorf("create new clientset: %w", err)
	}

	_, err = resourceClient.Create(staticpod)
	return err
}

func (c *Client) CreateInformer(crd *v1beta1.CustomResourceDefinition) error {
	// clusterClient, err := dynamic.NewForConfig(c.restConfig)
	// if err != nil {
	// 	return fmt.Errorf("create new dinamic client: %w", err)
	// }

	// resource := schema.GroupVersionResource{
	// 	Group:    crd.Spec.Group,
	// 	Version:  crd.GetResourceVersion(),
	// 	Resource: crd.GetObjectMeta().GetName(),
	// }
	// factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, crd.Namespace, nil)
	// informer := factory.ForResource(resource).Informer()
	return nil
}
