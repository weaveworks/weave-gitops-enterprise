package templates

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// TemplateDelimiterAnnotation can be added to a Template to change the Go
// template delimiter.
//
// It's assumed to be a string with "left,right"
// By default the delimiters are the standard Go templating delimiters:
// {{ and }}.
const TemplateDelimiterAnnotation string = "templates.weave.works/delimiters"

var templateFuncs template.FuncMap = makeTemplateFunctions()

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
func NewProcessorForTemplate(t templates.Template) (*TemplateProcessor, error) {
	switch t.GetSpec().RenderType {
	case "", templates.RenderTypeEnvsubst:
		return &TemplateProcessor{Processor: NewEnvsubstTemplateProcessor(), Template: t}, nil
	case templates.RenderTypeTemplating:
		return &TemplateProcessor{Processor: NewTextTemplateProcessor(t), Template: t}, nil

	}

	return nil, fmt.Errorf("unknown template renderType: %s", t.GetSpec().RenderType)
}

// TemplateProcessor does the work of rendering a template.
type TemplateProcessor struct {
	templates.Template
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
	for _, v := range p.GetSpec().ResourceTemplates {
		names, err := p.Processor.ParamNames(v.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to get params from template: %w", err)
		}
		paramNames.Insert(names...)
	}

	for k, v := range p.GetAnnotations() {
		if strings.HasPrefix(k, "capi.weave.works/profile-") {
			names, err := p.Processor.ParamNames([]byte(v))
			if err != nil {
				return nil, fmt.Errorf("failed to get params from annotation: %w", err)
			}
			paramNames.Insert(names...)
		}
	}

	for _, profile := range p.GetSpec().Charts.Items {
		if profile.HelmReleaseTemplate.Content != nil {
			names, err := p.Processor.ParamNames(profile.HelmReleaseTemplate.Content.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to get params from profile.spec of %s: %w", profile.Name, err)
			}
			paramNames.Insert(names...)
		}
		if profile.Values != nil {
			names, err := p.Processor.ParamNames(profile.Values.Raw)
			if err != nil {
				return nil, fmt.Errorf("failed to get params from profile.values of %s: %w", profile.Name, err)
			}
			paramNames.Insert(names...)
		}
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
//
// TODO: This should return []*unstructured.Unstructured to avoid having to
// convert back and forth.
func (p TemplateProcessor) RenderTemplates(vars map[string]string, opts ...RenderOptFunc) ([][]byte, error) {
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

	var processed [][]byte
	for _, v := range p.GetSpec().ResourceTemplates {
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
func NewTextTemplateProcessor(t templates.Template) *TextTemplateProcessor {
	return &TextTemplateProcessor{template: t}
}

// TextProcessor is an implementation of the Processor interface that uses Go's
// text/template to render templates.
type TextTemplateProcessor struct {
	template templates.Template
}

func (p *TextTemplateProcessor) Render(tmpl []byte, values map[string]string) ([]byte, error) {
	left, right := p.templateDelims()
	parsed, err := template.New("capi-template").Funcs(templateFuncs).Delims(left, right).Parse(string(tmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var out bytes.Buffer
	if err := parsed.Execute(&out, map[string]interface{}{"params": values}); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return out.Bytes(), nil
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

	left, right := p.templateDelims()
	paramsString := regexp.QuoteMeta(left) + `.*\.params\.([A-Za-z0-9_]+).*` + regexp.QuoteMeta(right)

	paramsRE, err := regexp.Compile(paramsString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters using regexp %q: %w", paramsString, err)
	}
	result := paramsRE.FindAllSubmatch(b, -1)
	variables := sets.NewString()
	for _, r := range result {
		variables.Insert(string(r[1]))
	}

	return variables.List(), nil
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
