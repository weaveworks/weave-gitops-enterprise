package templates

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

func ParseFile(fname string) (*templates.Template, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	return ParseBytes(b, fname)
}

func ParseFileFromFS(fsys fs.FS, fname string) (*templates.Template, error) {
	b, err := fs.ReadFile(fsys, fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	return ParseBytes(b, fname)
}

func ParseBytes(b []byte, key string) (*templates.Template, error) {
	var t templates.Template
	err := yaml.Unmarshal(b, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}
	return &t, nil
}

// ParseConfigMap parses a ConfigMap and returns a map of CAPITemplates indexed by their name.
// The name of the template is set to the key of the ConfigMap.Data map.
func ParseConfigMap(cm corev1.ConfigMap) (map[string]*templates.Template, error) {
	tm := map[string]*templates.Template{}

	for k, v := range cm.Data {
		t, err := ParseBytes([]byte(v), k)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal template %s from configmap %s, err: %w", k, cm.ObjectMeta.Name, err)
		}
		tm[t.Name] = t
	}
	return tm, nil
}

// Params extracts the named parameters from resource templates in a spec.
func Params(s templates.TemplateSpec) ([]string, error) {
	variables := sets.NewString()
	for _, v := range s.ResourceTemplates {
		params, err := paramsFromResourceTemplate(s, v)
		if err != nil {
			return nil, err
		}
		variables.Insert(params...)
	}
	names := variables.List()
	sort.Strings(names)
	return names, nil
}

// paramsFromResourceTemplate extracts the named parameters from a specific
// resource template.
func paramsFromResourceTemplate(s templates.TemplateSpec, rt templates.ResourceTemplate) ([]string, error) {
	var (
		names []string
		err   error
	)

	switch s.RenderType {
	case templates.RenderTypeTemplating:
		names, err = goTemplateParams(s, rt)
	default:
		names, err = envsubstParams(s, rt)
	}
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func envsubstParams(s templates.TemplateSpec, rt templates.ResourceTemplate) ([]string, error) {
	proc := processor.NewSimpleProcessor()
	variables := sets.NewString()
	tv, err := proc.GetVariables(rt.RawExtension.Raw)
	if err != nil {
		return nil, fmt.Errorf("processing template: %w", err)
	}
	variables.Insert(tv...)
	return variables.List(), nil
}

var paramsRE = regexp.MustCompile(`{{.*\.params\.([A-Za-z0-9_]+).*}}`)

func goTemplateParams(s templates.TemplateSpec, rt templates.ResourceTemplate) ([]string, error) {
	variables := sets.NewString()
	b, err := yaml.JSONToYAML(rt.RawExtension.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
	}
	result := paramsRE.FindAllSubmatch(b, -1)
	for _, r := range result {
		variables.Insert(string(r[1]))
	}
	return variables.List(), nil
}
