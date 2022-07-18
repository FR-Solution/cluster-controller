package kubernetes

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectConfig struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata"`
	Spec          ObjectConfigSpec   `json:"spec"`
	Status        ObjectConfigStatus `json:"status,omitempty"`
}
type ObjectConfigSpec struct {
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	Domain string `json:"domain"`
}

type ObjectConfigStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type ObjectConfigList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata"`
	Items       []ObjectConfig `json:"items"`
}
