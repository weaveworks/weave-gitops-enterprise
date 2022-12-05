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

	Charts ChartsSpec `json:"charts,omitempty"`

	// ResourceTemplates are a set of templates for resources that are generated
	// from this Template.
	ResourceTemplates []ResourceTemplate `json:"resourcetemplates,omitempty"`
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

// ChartsSpec defines the spec for a set of Helm charts.
// +kubebuilder:object:generate=true

type ChartsSpec struct {
	Items []Chart `json:"items,omitempty"`
}

// Chart is the set of values that control the default and required values
// of a chart/profile in a template.
// +kubebuilder:object:generate=true
type Chart struct {
	// Name of the chart/profile in the Helm repository.
	// Shortcut to template.content.spec.chart.spec.chart
	Chart string `json:"chart"`
	// Default version to select.
	// Shortcut to template.content.spec.chart.spec.version
	Version string `json:"version,omitempty"`
	// Shortcut to template.content.spec.targetNamespace
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// Layer, overrides the default layer provided in the Helm Repository
	Layer string `json:"layer,omitempty"`
	// If true the chart/profile will always be installed
	Required bool `json:"required,omitempty"`
	// If true you can change the values and version of the chart/profile
	Editable bool `json:"editable,omitempty"`
	// Shortcut to template.content.spec.values
	Values *HelmReleaseValues `json:"values,omitempty"`
	// Template for the HelmRelease, merged with the default template
	HelmReleaseTemplate HelmReleaseTemplateSpec `json:"template,omitempty"`
}

// HelmReleaseTemplateSpec is a future proof way to define a template with a path
// path is not yet used, but will be used in the near future
// +kubebuilder:object:generate=true

type HelmReleaseTemplateSpec struct {
	// Content of the template
	Content *HelmReleaseTemplate `json:"content,omitempty"`
}

// HelmReleaseTemplate is the HelmRelease.spec that can be overridden
// +kubebuilder:skipversion
// +kubebuilder:object:generate=true
// +kubebuilder:pruning:PreserveUnknownFields
type HelmReleaseTemplate struct {
	runtime.RawExtension `json:",inline"`
}

// HelmReleaseValues describes the values for a profile.
// +kubebuilder:skipversion
// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:object:generate=true
type HelmReleaseValues struct {
	runtime.RawExtension `json:",inline"`
}

// ResourceTemplate describes a resource to create.
// +kubebuilder:skipversion
// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:object:generate=true
type ResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
}
