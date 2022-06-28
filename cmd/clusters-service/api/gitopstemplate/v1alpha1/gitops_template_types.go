package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

const Kind = "GitOpsTemplate"

// GitOpsTemplate is the Schema for the GitOpsTemplate API
// +kubebuilder:object:root=true
type GitOpsTemplate struct {
	templates.Template `json:",inline"`
}

// GitOpsTemplateList contains a list of GitOpsTemplate
// +kubebuilder:object:root=true
type GitOpsTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOpsTemplate{}, &GitOpsTemplateList{})
}
