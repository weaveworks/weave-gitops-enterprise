package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CAPITemplateSpec defines the desired state of CAPITemplate
type CAPITemplateSpec struct {
	Description       string                 `json:"description,omitempty"`
	Params            []TemplateParam        `json:"params,omitempty"` // Described above
	ResourceTemplates []CAPIResourceTemplate `json:"resourcetemplates,omitempty"`
}

// Param is a parameter that can be templated into a struct.
type TemplateParam struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Options     []string `json:"options,omitempty"`
}

//+kubebuilder:pruning:PreserveUnknownFields

// CAPIResourceTemplate describes a resource to create.
type CAPIResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
}

// CAPITemplateStatus defines the observed state of CAPITemplate
type CAPITemplateStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CAPITemplate is the Schema for the capitemplates API
type CAPITemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPITemplateSpec   `json:"spec,omitempty"`
	Status CAPITemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CAPITemplateList contains a list of CAPITemplate
type CAPITemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPITemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPITemplate{}, &CAPITemplateList{})
}
