package templates

// TemplateSpec defines the base template spec needs for CAPI or Terraform Templates.
// +kubebuilder:object:generate=true
type TemplateSpecV1 struct {
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
