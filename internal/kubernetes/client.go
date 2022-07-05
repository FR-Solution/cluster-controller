package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	cli *clientset.Clientset
}

func NewClient(kubeconfigPath string) (*Client, error) {
	configBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %s : %w", kubeconfigPath, err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("new rest config: %w\n", err)
	}

	cli, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("new clientset: %w", err)
	}
	return &Client{
		cli: cli,
	}, err
}

func (c *Client) CreateCDR(crdData []byte) error {
	var tmp map[string]any
	if err := yaml.Unmarshal(crdData, &tmp); err != nil {
		return fmt.Errorf("parse manifest to cdr struct: %w", err)
	}
	data, err := json.Marshal(tmp)
	if err != nil {
		return fmt.Errorf("parse manifest to cdr struct: %w", err)
	}

	crd := new(v1beta1.CustomResourceDefinition)
	if err := json.Unmarshal(data, crd); err != nil {
		return fmt.Errorf("parse manifest to cdr struct: %w", err)
	}

	_, err = c.cli.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), crd, v1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create cdr in kubernetes: %w", err)
	}

	return nil
}
