package kubernetes

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type staticpod struct {
	v1.Deployment `json:",inline" yaml:",inline"`
}

var (
	group       = "fraima.io"
	kind        = "staticpod"
	listKind    = "StaticPodList"
	plural      = "staticpod"
	singular    = "staticpod"
	versionName = "v1beta1"
	namespace   = "kube-system"

	staticpodCRD = &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", plural, group),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     kind,
				ListKind: listKind,
				Plural:   plural,
				Singular: singular,
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: versionName,
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
	staticpodTemplate = &staticpod{
		Deployment: v1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: fmt.Sprintf("%s/%s", group, versionName),
				Kind:       kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
		},
	}
)
