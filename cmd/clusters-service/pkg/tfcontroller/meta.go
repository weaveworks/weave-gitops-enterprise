package tfcontroller

import (
	"fmt"

	tapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/tfcontroller/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
)

// ParseTemplateMeta parses a byte slice into a TemplateMeta struct which
// contains the objects that are in the template, along with the parameters used
// by each of the objects.

const (
	// DisplayNameAnnotation is the annotation used for labeling template resources used by tf-controller
	DisplayNameAnnotation = "tfcontroller.weave.works/display-name"
)

func ParseTemplateMeta(s *tapiv1.TFTemplate) (*TemplateMeta, error) {
	proc := processor.NewSimpleProcessor()
	variables := map[string]bool{}
	var objects []Object
	for _, v := range s.Spec.ResourceTemplates {
		tv, err := proc.GetVariables(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to get parameters processing template: %w", err)
		}
		for _, n := range tv {
			variables[n] = true
		}
		var uv unstructured.Unstructured
		if err := uv.UnmarshalJSON(v.RawExtension.Raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resourceTemplate: %w", err)
		}
		objects = append(objects, Object{
			Kind:       uv.GetKind(),
			APIVersion: uv.GetAPIVersion(),
			Params:     tv, Name: uv.GetName(),
			DisplayName: uv.GetAnnotations()[DisplayNameAnnotation],
		})
	}

	enriched, err := templates.ParamsFromSpec(s.Spec)
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
// the parametesr.
type Object struct {
	Kind        string   `json:"kind"`
	APIVersion  string   `json:"apiVersion"`
	Name        string   `json:"name"`
	Params      []string `json:"params"`
	DisplayName string   `json:"displayName"`
}

// TemplateMeta contains all the objects extracted from a TFTemplate along
// with the parameters.
type TemplateMeta struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Params      []templates.Param `json:"params,omitempty"`
	Objects     []Object          `json:"objects,omitempty"`
}
