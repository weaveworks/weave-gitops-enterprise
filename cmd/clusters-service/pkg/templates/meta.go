package templates

import (
	"fmt"

	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// DisplayNameAnnotation is the annotation used for labeling template resources
	GitOpsTemplateNameAnnotation = "clustertemplates.weave.works/display-name"
	CAPIDisplayNameAnnotation    = "capi.weave.works/display-name"
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
	for _, v := range s.Spec.ResourceTemplates {
		params, err := processor.ParamNames(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse params in template: %w", err)
		}
		var uv unstructured.Unstructured
		if err := uv.UnmarshalJSON(v.RawExtension.Raw); err != nil {
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

	enriched, err := processor.Params()
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters from the spec: %w", err)
	}
	return &TemplateMeta{
		Description: s.Spec.Description,
		Name:        s.ObjectMeta.Name,
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
