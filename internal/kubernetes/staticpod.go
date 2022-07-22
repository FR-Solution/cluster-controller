package kubernetes

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StaticpodResource struct {
	v1.Deployment
}

var (
	Group       = "fraima.io"
	Kind        = "staticpod"
	ListKind    = "StaticPodList"
	Plural      = "staticpod"
	Singular    = "staticpod"
	VersionName = "v1beta1"

	staticpodCRD = &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", Plural, Group),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: Group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     Kind,
				ListKind: ListKind,
				Plural:   Plural,
				Singular: Singular,
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: VersionName,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type:     "object",
							Required: []string{"metadata", "spec"},
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"metadata": {
									Type: "object",
								},
								"spec": {
									Type: "object",
								},
							},
						},
					},
					Served:  true,
					Storage: true,
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
				},
			},
		},
	}
	staticpodTemplate = &StaticpodResource{
		Deployment: v1.Deployment{
			TypeMeta: meta_v1.TypeMeta{
				APIVersion: fmt.Sprintf("%s/%s", Group, VersionName),
				Kind:       Kind,
			},
		},
	}
)
