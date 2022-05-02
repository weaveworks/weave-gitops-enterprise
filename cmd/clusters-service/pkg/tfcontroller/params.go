package tfcontroller

import (
	"fmt"
	"sort"

	apiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
)

// Param is a parameter that can be templated into a struct.
type Param struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Options     []string `json:"options"`
}

// ParamsFromSpec extracts the named parameters from a TFTemplate, finding all
// the named parameters in each of the resource templates, and enriching that
// with data from the params field in the spec (if found).
//
// Any fields in the templates, but not in the params will not be enriched, and
// only the name will be returned.
func ParamsFromSpec(s apiv1.TFTemplateSpec) ([]Param, error) {
	paramNames, err := Params(s)
	if err != nil {
		return nil, fmt.Errorf("failed to get params from template: %w", err)
	}
	paramsMeta := map[string]Param{}
	for _, v := range paramNames {
		paramsMeta[v] = Param{Name: v}
	}

	for _, v := range s.Params {
		if m, ok := paramsMeta[v.Name]; ok {
			m.Description = v.Description
			m.Options = v.Options
			m.Required = v.Required
			paramsMeta[v.Name] = m
		}
	}

	var params []Param
	for _, v := range paramsMeta {
		params = append(params, v)
	}
	sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })
	return params, nil
}
