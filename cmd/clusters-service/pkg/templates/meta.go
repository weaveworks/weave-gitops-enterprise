package templates

import (
	"fmt"

	apitemplates "github.com/weaveworks/templates-controller/apis/core"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// DisplayNameAnnotation is the annotation used for labeling template resources
	GitOpsTemplateNameAnnotation = "templates.weave.works/display-name"
	CAPIDisplayNameAnnotation    = "capi.weave.works/display-name"
	// CostEstimationAnnotation is to signal we should try and estimate the cost of a template when rendering it
	CostEstimationAnnotation        = "templates.weave.works/cost-estimation-enabled"
	AddCommonBasesAnnotation        = "templates.weave.works/add-common-bases"
	InjectPruneAnnotationAnnotation = "templates.weave.works/inject-prune-annotation"
	ProfilesAnnotation              = "capi.weave.works/profile-"
	SopsKustomizationAnnotation     = "templates.weave.works/sops-enabled"
)

// ParseTemplateMeta parses a byte slice into a TemplateMeta struct which
// contains the objects that are in the template, along with the parameters used
// by each of the objects.
func ParseTemplateMeta(s apitemplates.Template, annotation string) (*TemplateMeta, error) {
	processor, err := NewProcessorForTemplate(s)
	if err != nil {
		return nil, err
	}

	var objects []Object
	for _, resourcetemplateDefinition := range s.GetSpec().ResourceTemplates {
		for _, v := range resourcetemplateDefinition.Content {
			params, err := processor.ParamNames(v.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse params in template: %w", err)
			}
			var uv unstructured.Unstructured
			if err := uv.UnmarshalJSON(v.Raw); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resourceTemplate: %w", err)
			}
			objects = append(objects, Object{
				Kind:        uv.GetKind(),
				APIVersion:  uv.GetAPIVersion(),
				Params:      params,
				Name:        uv.GetName(),
				DisplayName: uv.GetAnnotations()[annotation],
			})
		}
	}

	enriched, err := processor.Params()
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters from the spec: %w", err)
	}
	return &TemplateMeta{
		Description: s.GetSpec().Description,
		Name:        s.GetName(),
		Objects:     objects,
		Params:      enriched,
	}, nil
}

// Object contains the details of the object rendered from a template along with
// the parameters.
type Object struct {
	Kind        string   `json:"kind"`
	APIVersion  string   `json:"apiVersion"`
	Name        string   `json:"name"`
	Params      []string `json:"params"`
	DisplayName string   `json:"displayName"`
}

// TemplateMeta contains all the objects
// with the parameters.
type TemplateMeta struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Params      []Param  `json:"params,omitempty"`
	Objects     []Object `json:"objects,omitempty"`
}
