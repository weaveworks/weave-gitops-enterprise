package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// These are the options for rendering templates using different "languages"
// e.g. envsubst or Go templating.
const (
	// RenderTypeEnvsubst uses https://github.com/a8m/envsubst for rendering
	RenderTypeEnvsubst = "envsubst"
	// RenderTypeTemplating use text/templating for rendering
	RenderTypeTemplating = "templating"
)

// Template defines a template which can be either a CAPI or a Terraform Template.
// +kubebuilder:object:generate=true
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TemplateSpec `json:"spec,omitempty"`
}

// TemplateSpec defines the base template spec needs for CAPI or Terraform Templates.
// +kubebuilder:object:generate=true
type TemplateSpec struct {
	// Description is used to describe the purpose of this template for user
	// information.
	Description string `json:"description,omitempty"`

	// RenderType specifies which templating language to use to render
	// templates.
	// Defaults to 'envsubst', valid values are ('envsubst', 'templating').
	// +kubebuilder:validation:Enum=envsubst;templating
	// +kubebuilder:default:=envsubst
	// +optional
	RenderType string `json:"renderType,omitempty"`

	// Params is the set of parameters that are used in this template with
	// descriptions.
	Params []TemplateParam `json:"params,omitempty"` // Described above

	// ResourceTemplates are a set of templates for resources that are generated
	// from this Template.
	ResourceTemplates []ResourceTemplate `json:"resourcetemplates,omitempty"`
}

// TemplateParam is a parameter that can be templated into a struct.
// +kubebuilder:object:generate=true
type TemplateParam struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Options     []string `json:"options,omitempty"`
}

// ResourceTemplate describes a resource to create.
// +kubebuilder:skipversion
// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:object:generate=true
type ResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
}
