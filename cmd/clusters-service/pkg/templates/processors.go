package templates

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig/v3"
)

// TemplateDelimiterAnnotation can be added to a Template to change the Go
// template delimiter.
//
// It's assumed to be a string with "left,right"
// By default the delimiters are the standard Go templating delimiters:
// {{ and }}.
const TemplateDelimiterAnnotation string = "templates.weave.works/delimiters"

var templateFuncs template.FuncMap = makeTemplateFunctions()

type RenderedTemplate struct {
	Data [][]byte
	Path string
}

// Processor is a generic template parser/renderer.
type Processor interface {
	// Render implementations accept a template and a key/value map of params to
	// render and render the template.
	Render([]byte, map[string]string) ([]byte, error)
	// ParamNames implementations parse a resource template and extract the
	// parameters.
	ParamNames([]byte) ([]string, error)
}

// RenderOptFunc is a functional option for Rendering templates.
type RenderOptFunc func(uns *unstructured.Unstructured) error

// NewProcessorForTemplate creates and returns an appropriate processor for a
// template based on its declared type.
func NewProcessorForTemplate(t templatesv1.Template) (*TemplateProcessor, error) {
	switch t.GetSpec().RenderType {
	case "", templatesv1.RenderTypeEnvsubst:
		return &TemplateProcessor{Processor: NewEnvsubstTemplateProcessor(), Template: t}, nil
	case templatesv1.RenderTypeTemplating:
		return &TemplateProcessor{Processor: NewTextTemplateProcessor(t), Template: t}, nil

	}

	return nil, fmt.Errorf("unknown template renderType: %s", t.GetSpec().RenderType)
}

// TemplateProcessor does the work of rendering a template.
type TemplateProcessor struct {
	templatesv1.Template
	Processor
}

// Params returns the set of parameters discovered in the resource templates.
//
// These are discovered in the templates, and enriched from the parameters
// declared on the template.
//
// The returned slice is sorted by Name.
func (p TemplateProcessor) Params() ([]Param, error) {
	paramNames := sets.NewString()
	for _, resourcetemplateDefinition := range p.GetSpec().ResourceTemplates {
		if resourcetemplateDefinition.Content != nil && resourcetemplateDefinition.Raw != "" {
			return nil, fmt.Errorf("cannot specify both raw and content in the same resource template: %s/%s",
				p.GetName(), p.GetNamespace())
		}

		names, err := p.Processor.ParamNames([]byte(resourcetemplateDefinition.Path))
		if err != nil {
			return nil, fmt.Errorf("failed to get params from template path: %w", err)
		}
		paramNames.Insert(names...)

		for _, v := range resourcetemplateDefinition.Content {
			names, err := p.Processor.ParamNames(v.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to get params from template: %w", err)
			}
			paramNames.Insert(names...)
		}

		if resourcetemplateDefinition.Raw != "" {
			names, err := p.Processor.ParamNames([]byte(resourcetemplateDefinition.Raw))
			if err != nil {
				return nil, fmt.Errorf("failed to get params from raw template: %w", err)
			}
			paramNames.Insert(names...)
		}
	}

	for _, v := range ProfileAnnotations(p) {
		names, err := p.Processor.ParamNames([]byte(v))
		if err != nil {
			return nil, fmt.Errorf("failed to get params from annotation: %w", err)
		}
		paramNames.Insert(names...)
	}

	for _, profile := range p.GetSpec().Charts.Items {
		if profile.HelmReleaseTemplate.Content != nil {
			names, err := p.Processor.ParamNames(profile.HelmReleaseTemplate.Content.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to get params from profile.spec of %s: %w", profile.Chart, err)
			}
			paramNames.Insert(names...)
		}
		if profile.HelmReleaseTemplate.Path != "" {
			names, err := p.Processor.ParamNames([]byte(profile.HelmReleaseTemplate.Path))
			if err != nil {
				return nil, fmt.Errorf("failed to get params from profile.spec of %s: %w", profile.Chart, err)
			}
			paramNames.Insert(names...)
		}
		if profile.Values != nil {
			names, err := p.Processor.ParamNames(profile.Values.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to get params from profile.values of %s: %w", profile.Chart, err)
			}
			paramNames.Insert(names...)
		}
	}

	if p.GetSpec().Charts.HelmRepositoryTemplate.Path != "" {
		helmRepoTemplatePath := p.GetSpec().Charts.HelmRepositoryTemplate.Path
		names, err := p.Processor.ParamNames([]byte(helmRepoTemplatePath))
		if err != nil {
			return nil, fmt.Errorf("failed to get params from chart.helmRepositoryTemplate.path of %s: %w", helmRepoTemplatePath, err)
		}
		paramNames.Insert(names...)
	}

	paramsMeta := map[string]Param{}
	for _, v := range paramNames.List() {
		paramsMeta[v] = Param{Name: v}
	}

	declaredParams := p.GetSpec().Params
	for _, v := range declaredParams {
		if m, ok := paramsMeta[v.Name]; ok {
			m.Description = v.Description
			m.Options = v.Options
			m.Required = v.Required
			m.Default = v.Default
			paramsMeta[v.Name] = m
		}
	}

	var params []Param
	for _, v := range paramsMeta {
		params = append(params, v)
	}

	paramOrders := map[string]int{}
	for i, v := range declaredParams {
		paramOrders[v.Name] = i
	}
	sort.Slice(params, func(i, j int) bool {
		iOrder, iExists := paramOrders[params[i].Name]
		jOrder, jExists := paramOrders[params[j].Name]
		switch {
		case iExists && jExists:
			return iOrder < jOrder
		case !iExists && !jExists:
			return params[i].Name < params[j].Name
		case iExists:
			return true
		default:
			return false
		}
	})

	return params, nil
}

// RenderTemplates renders all the resourceTemplates in the template.
func (p TemplateProcessor) RenderTemplates(vars map[string]string, opts ...RenderOptFunc) ([]RenderedTemplate, error) {
	params, err := p.Params()
	if err != nil {
		return nil, err
	}

	if vars == nil {
		vars = map[string]string{}
	}

	for _, param := range params {
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

	var renderedTemplates []RenderedTemplate
	for _, resourcetemplateDefinition := range p.GetSpec().ResourceTemplates {
		if resourcetemplateDefinition.Content != nil && resourcetemplateDefinition.Raw != "" {
			return nil, fmt.Errorf("cannot specify both raw and content in the same resource template: %s/%s",
				p.GetName(), p.GetNamespace())
		}

		var renderedPath string
		if resourcetemplateDefinition.Path != "" {
			p, err := p.Processor.Render([]byte(resourcetemplateDefinition.Path), vars)
			if err != nil {
				return nil, fmt.Errorf("failed to render resource template definition path: %w", err)
			}
			renderedPath = string(p)
		}
		var processed [][]byte
		if resourcetemplateDefinition.Raw != "" {
			data, err := p.Processor.Render([]byte(resourcetemplateDefinition.Raw), vars)
			if err != nil {
				return nil, fmt.Errorf("processing template: %w", err)
			}

			processed = append(processed, data)
		} else {
			for _, v := range resourcetemplateDefinition.Content {
				b, err := yaml.JSONToYAML(v.Raw)
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
		}
		renderedTemplates = append(renderedTemplates, RenderedTemplate{
			Data: processed,
			Path: renderedPath,
		})
	}

	return renderedTemplates, nil
}

// NewTextTemplateProcessor creates and returns a new TextTemplateProcessor.
func NewTextTemplateProcessor(t templatesv1.Template) *TextTemplateProcessor {
	return &TextTemplateProcessor{template: t}
}

// TextProcessor is an implementation of the Processor interface that uses Go's
// text/template to render .
type TextTemplateProcessor struct {
	template templatesv1.Template
}

func (p *TextTemplateProcessor) Render(tmpl []byte, values map[string]string) ([]byte, error) {
	parsed, err := p.parseToTemplate(tmpl)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := parsed.Execute(&out, map[string]interface{}{
		"params":   values,
		"template": templateMetadata(p.template),
	}); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return out.Bytes(), nil
}

func (p *TextTemplateProcessor) parseToTemplate(tmpl []byte) (*template.Template, error) {
	templateName := p.template.GetName()
	left, right := p.templateDelims()
	parsed, err := template.New(templateName).Funcs(templateFuncs).Delims(left, right).Parse(string(tmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return parsed, nil
}

func (p *TextTemplateProcessor) templateDelims() (string, string) {
	ann, ok := p.template.GetAnnotations()[TemplateDelimiterAnnotation]
	if ok {
		if elems := strings.Split(ann, ","); len(elems) == 2 {
			return elems[0], elems[1]
		}
	}
	return "{{", "}}"
}

func (p *TextTemplateProcessor) ParamNames(raw []byte) ([]string, error) {
	b, err := yaml.JSONToYAML(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
	}

	parsed, err := p.parseToTemplate(b)
	if err != nil {
		return nil, err
	}

	// The *parse.Tree field is exported only for use by html/template
	// and should be treated as unexported by all other clients.
	//
	// This is a bit naughty but we would need to recreate the template builtins
	// to pass to the parser implementation ("eq" etc) as they are internal to
	// the package.
	return parseTextParamNames(parsed.Tree), nil
}

// NewEnvsubstTemplateProcessor creates and returns a new
// EnvsubstTemplateProcessor.
func NewEnvsubstTemplateProcessor() *EnvsubstTemplateProcessor {
	return &EnvsubstTemplateProcessor{}
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

func (p *EnvsubstTemplateProcessor) ParamNames(raw []byte) ([]string, error) {
	proc := processor.NewSimpleProcessor()
	variables := sets.NewString()
	tv, err := proc.GetVariables(raw)
	if err != nil {
		return nil, fmt.Errorf("processing template: %w", err)
	}
	variables.Insert(tv...)

	return variables.List(), nil
}

func makeTemplateFunctions() template.FuncMap {
	f := sprig.TxtFuncMap()
	unwanted := []string{
		"env", "expandenv", "getHostByName", "genPrivateKey", "derivePassword", "sha256sum",
		"base", "dir", "ext", "clean", "isAbs", "osBase", "osDir", "osExt", "osClean", "osIsAbs"}
	for _, v := range unwanted {
		delete(f, v)
	}
	return f
}

// this could add additional fields from the data.
func templateMetadata(t templatesv1.Template) map[string]any {
	return map[string]any{
		"meta": map[string]any{
			"name":      t.GetName(),
			"namespace": t.GetNamespace(),
		},
	}
}
