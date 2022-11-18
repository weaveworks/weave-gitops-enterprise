package templates

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Template implementations provide access to key fields for different
// templates.
type Template interface {
	GetName() string
	GetNamespace() string
	GetSpec() TemplateSpec
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetObjectKind() schema.ObjectKind
}

// These are the options for rendering templates using different "languages"
// e.g. envsubst or Go templating.
const (
	// RenderTypeEnvsubst uses https://github.com/a8m/envsubst for rendering
	RenderTypeEnvsubst = "envsubst"
	// RenderTypeTemplating use text/templating for rendering
	RenderTypeTemplating = "templating"
)

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

	// TestField is purely here to ensure that our conversions are working.
	TestField string `json:"testField,omitempty"`
}

// TemplateParam is a parameter that can be templated into a struct.
// +kubebuilder:object:generate=true
type TemplateParam struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Required indicates whether the param must contain a non-empty value
	// +kubebuilder:default:=true
	// +optional
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"`

	// Default specifies the default value for the parameter
	// +optional
	Default string `json:"default,omitempty"`
}

// ResourceTemplate describes a resource to create.
// +kubebuilder:skipversion
// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:object:generate=true
type ResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
}
