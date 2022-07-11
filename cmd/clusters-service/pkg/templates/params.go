package templates

import (
	"fmt"
	"sort"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// ParamsFromTemplate extracts the named parameters from a CAPITemplate, finding all
// the named parameters in each of the resource templates, and enriching that
// with data from the params field in the spec (if found).
//
// Any fields in the templates, but not in the params will not be enriched, and
// only the name will be returned.
func ParamsFromTemplate(t templates.Template) ([]Param, error) {
	paramNames, err := Params(t)
	if err != nil {
		return nil, fmt.Errorf("failed to get params from template: %w", err)
	}
	paramsMeta := map[string]Param{}
	for _, v := range paramNames {
		paramsMeta[v] = Param{Name: v}
	}

	for _, v := range t.Spec.Params {
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
