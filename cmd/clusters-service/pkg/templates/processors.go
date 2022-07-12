package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// Processor is a generic template parser/renderer.
type Processor interface {
	// Render implementations accept a template and a key/value map of params to
	// render and render the template.
	Render([]byte, map[string]string) ([]byte, error)
	// ParamNames implementations parse a resource template and extract the
	// parameters.
	ParamNames(templates.ResourceTemplate) ([]string, error)
}

// RenderOptFunc is a functional option for Rendering templates.
type RenderOptFunc func(uns *unstructured.Unstructured) error

// NewProcessorForTemplate creates and returns an appropriate processor for a
// template based on its declared type.
func NewProcessorForTemplate(t templates.Template) (*TemplateProcessor, error) {
	switch t.Spec.RenderType {
	case "", templates.RenderTypeEnvsubst:
		return &TemplateProcessor{Processor: NewEnvsubstTemplateProcessor(), Template: t}, nil
	case templates.RenderTypeTemplating:
		return &TemplateProcessor{Processor: NewTextTemplateProcessor(), Template: t}, nil

	}
	return nil, fmt.Errorf("unknown template renderType: %s", t.Spec.RenderType)
}

// TemplateProcessor does the work of rendering a template.
type TemplateProcessor struct {
	templates.Template
	Processor
}

func (p TemplateProcessor) Params() ([]Param, error) {
	paramNames := sets.NewString()
	for _, v := range p.Template.Spec.ResourceTemplates {
		names, err := p.Processor.ParamNames(v)
		if err != nil {
			return nil, fmt.Errorf("failed to get params from template: %w", err)
		}
		paramNames.Insert(names...)
	}

	paramsMeta := map[string]Param{}
	for _, v := range paramNames.List() {
		paramsMeta[v] = Param{Name: v}
	}

	for _, v := range p.Template.Spec.Params {
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

// RenderTemplates renders all the resourceTemplates in the template.
func (p TemplateProcessor) RenderTemplates(vars map[string]string, opts ...RenderOptFunc) ([][]byte, error) {
	var processed [][]byte
	for _, v := range p.Template.Spec.ResourceTemplates {
		b, err := yaml.JSONToYAML(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
		}

		data, err := p.Processor.Render(b, vars)
		if err != nil {
			return nil, fmt.Errorf("processing template: %w", err)
		}

		data, err = processUnstructured(data, opts...)
		if err != nil {
			return nil, fmt.Errorf("modifying template: %w", err)
		}
		processed = append(processed, data)
	}

	return processed, nil
}

// NewTextTemplateProcessor creates and returns a new TextTemplateProcessor.
func NewTextTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{Processor: &TextTemplateProcessor{}}
}

// TextProcessor is an implementation of the Processor interface that uses Go's
// text/template to render templates.
type TextTemplateProcessor struct {
}

func (p *TextTemplateProcessor) Render(tmpl []byte, values map[string]string) ([]byte, error) {
	parsed, err := template.New("capi-template").Parse(string(tmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var out bytes.Buffer
	if err := parsed.Execute(&out, map[string]interface{}{"params": values}); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return out.Bytes(), nil
}

var paramsRE = regexp.MustCompile(`{{.*\.params\.([A-Za-z0-9_]+).*}}`)

func (p *TextTemplateProcessor) ParamNames(rt templates.ResourceTemplate) ([]string, error) {
	b, err := yaml.JSONToYAML(rt.RawExtension.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
	}
	result := paramsRE.FindAllSubmatch(b, -1)
	variables := sets.NewString()
	for _, r := range result {
		variables.Insert(string(r[1]))
	}

	return variables.List(), nil
}

func (p TemplateProcessor) AllParamNames() ([]string, error) {
	paramNames := sets.NewString()
	for _, v := range p.Template.Spec.ResourceTemplates {
		names, err := p.Processor.ParamNames(v)
		if err != nil {
			return nil, fmt.Errorf("failed to get params from template: %w", err)
		}
		paramNames.Insert(names...)
	}

	params := paramNames.List()
	sort.Strings(params)

	return params, nil
}

// NewEnvsubstTemplateProcessor creates and returns a new
// EnvsubstTemplateProcessor.
func NewEnvsubstTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{Processor: &EnvsubstTemplateProcessor{}}
}

// EnvsubstTemplateProcessor is an implementation of the Processor interface
// that uses envsubst to render templates.
type EnvsubstTemplateProcessor struct {
}

func (p *EnvsubstTemplateProcessor) Render(tmpl []byte, values map[string]string) ([]byte, error) {
	proc := processor.NewSimpleProcessor()

	rendered, err := proc.Process(tmpl, func(n string) (string, error) {
		if s, ok := values[n]; ok {
			return s, nil
		}
		return "", fmt.Errorf("variable %s not found", n)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process template values: %w", err)
	}

	return rendered, nil
}

func (p *EnvsubstTemplateProcessor) ParamNames(rt templates.ResourceTemplate) ([]string, error) {
	proc := processor.NewSimpleProcessor()
	variables := sets.NewString()
	tv, err := proc.GetVariables(rt.RawExtension.Raw)
	if err != nil {
		return nil, fmt.Errorf("processing template: %w", err)
	}
	variables.Insert(tv...)

	return variables.List(), nil
}
