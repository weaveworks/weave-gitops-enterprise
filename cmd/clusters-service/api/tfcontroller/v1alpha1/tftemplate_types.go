package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

const Kind = "TFTemplate"

//+kubebuilder:object:root=true

// TFTemplate is the Schema for the TFTemplates API
type TFTemplate struct {
	templates.Template
}

//+kubebuilder:object:root=true

// TFTemplateList contains a list of TFTemplate
type TFTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TFTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TFTemplate{}, &TFTemplateList{})
}
