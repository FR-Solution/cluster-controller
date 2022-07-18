package kubernetes

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), staticpodCRD, v1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create cdr in kubernetes: %w", err)
	}

	return nil
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
