package v1alpha1

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const Kind = "CAPITemplate"

// CAPITemplate is the Schema for the capitemplates API
// +kubebuilder:object:root=true
type CAPITemplate struct {
	templates.Template `json:",inline"`
}

// CAPITemplateList contains a list of CAPITemplate
// +kubebuilder:object:root=true
type CAPITemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPITemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPITemplate{}, &CAPITemplateList{})
}
