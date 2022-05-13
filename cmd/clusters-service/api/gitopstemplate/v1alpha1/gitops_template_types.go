package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

const Kind = "GitOpsTemplate"

//+kubebuilder:object:root=true

// GitOpsTemplate is the Schema for the GitOpsTemplate API
type GitOpsTemplate struct {
	templates.Template
}

//+kubebuilder:object:root=true

// GitOpsTemplateList contains a list of GitOpsTemplate
type GitOpsTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOpsTemplate{}, &GitOpsTemplateList{})
}
