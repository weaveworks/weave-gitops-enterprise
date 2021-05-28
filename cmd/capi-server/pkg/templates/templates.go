package templates

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
)

// TemplateParams is a map of parameter to value for rendering templates.
type TemplateParams map[string]string

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name string) ([]byte, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
}

// RenderTemplate renders the named template loading it from the library, and combining it with the provided parameters.
func RenderTemplate(ctx context.Context, getter TemplateGetter, name string, params TemplateParams) ([]byte, error) {
	template, err := getter.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not find template %q: %w", name, err)
	}
	return renderTemplate(template, params)
}

// renderTemplate renders a template given a body and parameters to fill in the template fields.
func renderTemplate(template []byte, params TemplateParams) ([]byte, error) {
	proc := yamlprocessor.NewSimpleProcessor()
	processedYAML, err := proc.Process(template, func(key string) (string, error) {
		if value, ok := params[key]; ok {
			return value, nil
		}
		return "", fmt.Errorf("failed to find template parameter %q", key)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return processedYAML, nil
}
