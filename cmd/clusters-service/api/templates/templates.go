//+kubebuilder:object:generate=true
package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
}

//+kubebuilder:subresource:status
//+kubebuilder:object:root=true

// Template defines a template which can be either a CAPI or a Terraform Template.
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// TemplateSpec defines the base template spec needs for CAPI or Terraform Templates.
type TemplateSpec struct {
	Description       string             `json:"description,omitempty"`
	Params            []TemplateParam    `json:"params,omitempty"` // Described above
	ResourceTemplates []ResourceTemplate `json:"resourcetemplates,omitempty"`
}

// TemplateParam is a parameter that can be templated into a struct.
type TemplateParam struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Options     []string `json:"options,omitempty"`
}

//+kubebuilder:pruning:PreserveUnknownFields

// ResourceTemplate describes a resource to create.
type ResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
}
