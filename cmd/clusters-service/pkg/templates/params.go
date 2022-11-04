package templates

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// ParamsFromTemplate extracts the named parameters from a CAPITemplate, finding all
// the named parameters in each of the resource templates, and enriching that
// with data from the params field in the spec (if found).
//
// Any fields in the templates, but not in the params will not be enriched, and
// only the name will be returned.
func ParamsFromTemplate(t templates.Template) ([]templates.TemplateParam, error) {
	proc, err := NewProcessorForTemplate(t)
	if err != nil {
		return nil, err
	}
	return proc.Params()
}

func ParamValuesWithDefaults(specParams []templates.TemplateParam, vars map[string]string) (map[string]string, error) {
	if vars == nil {
		vars = map[string]string{}
	}

	for _, param := range specParams {
		val, ok := vars[param.Name]
		if !ok || val == "" {
			if param.Default != "" {
				vars[param.Name] = param.Default
			} else {
				if param.Required {
					return nil, fmt.Errorf("missing required parameter: %s", param.Name)
				}
				vars[param.Name] = ""
			}
		}
	}

	return vars, nil
}
