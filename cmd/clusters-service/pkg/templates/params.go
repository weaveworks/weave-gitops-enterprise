package templates

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// ParamsFromTemplate extracts the named parameters from a CAPITemplate, finding all
// the named parameters in each of the resource templates, and enriching that
// with data from the params field in the spec (if found).
//
// Any fields in the templates, but not in the params will not be enriched, and
// only the name will be returned.
func ParamsFromTemplate(t templates.Template) ([]Param, error) {
	proc, err := NewProcessorForTemplate(t)
	if err != nil {
		return nil, err
	}
	return proc.Params()
}
