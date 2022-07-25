package kubernetes

import (
	"context"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   group,
	Version: versionName,
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&staticpod{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

type resourceClient struct {
	cli *rest.RESTClient
}

func newResourceClient(cfg *rest.Config) (*resourceClient, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &resourceClient{cli: client}, nil
}

func (c *resourceClient) Create(obj *staticpod) (*staticpod, error) {
	result := &staticpod{}
	err := c.cli.Post().
		Namespace(namespace).Resource(plural).
		Body(obj).
		Do(context.Background()).
		Into(result)
	return result, err
}

func (c *resourceClient) Update(obj *staticpod) (*staticpod, error) {
	result := &staticpod{}
	err := c.cli.Put().
		Namespace(namespace).Resource(plural).
		Body(obj).
		Do(context.Background()).
		Into(result)
	return result, err
}

func (c *resourceClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.cli.Delete().
		Namespace(namespace).Resource(plural).
		Name(name).
		Body(options).
		Do(context.Background()).
		Error()
}

func (c *resourceClient) Get(name string) (*staticpod, error) {
	result := &staticpod{}
	err := c.cli.Get().
		Namespace(namespace).Resource(plural).
		Name(name).
		Do(context.Background()).
		Into(result)
	return result, err
}
