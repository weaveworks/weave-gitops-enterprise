package templates

import (
	"fmt"
	"regexp"

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

type templateRenderFunc func(tmpl []byte, values map[string]string) ([]byte, error)

var templateRenderers = map[string]templateRenderFunc{
	templates.RenderTypeEnvsubst:   ProcessTemplate,
	templates.RenderTypeTemplating: processTemplatingTemplate,
}

// NewTextTemplateProcessor creates and returns a new TextTemplateProcessor.
func NewTextTemplateProcessor() *TextTemplateProcessor {
	return &TextTemplateProcessor{}
}

// TextProcessor is an implementation of the Processor interface that uses Go's
// text/template to render templates.
type TextTemplateProcessor struct {
}

func (p *TextTemplateProcessor) Render([]byte, map[string]string) ([]byte, error) {
	return nil, nil
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

// NewEnvsubstTemplateProcessor creates and returns a new
// EnvsubstTemplateProcessor.
func NewEnvsubstTemplateProcessor() *EnvsubstTemplateProcessor {
	return &EnvsubstTemplateProcessor{}
}

// EnvsubstTemplateProcessor is an implementation of the Processor interface
// that uses envsubst to render templates.
type EnvsubstTemplateProcessor struct {
}

func (p *EnvsubstTemplateProcessor) Render([]byte, map[string]string) ([]byte, error) {
	return nil, nil
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
