package capi

import (
	"fmt"
	"io/fs"
	"os"
	"sort"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"
)

func ParseFile(fname string) (*capiv1.CAPITemplate, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	return ParseBytes(b, fname)
}

func ParseFileFromFS(fsys fs.FS, fname string) (*capiv1.CAPITemplate, error) {
	b, err := fs.ReadFile(fsys, fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	return ParseBytes(b, fname)
}

func ParseBytes(b []byte, key string) (*capiv1.CAPITemplate, error) {
	var t capiv1.CAPITemplate
	err := yaml.Unmarshal(b, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}
	return &t, nil
}

// ParseConfigMap parses a ConfigMap and returns a map of CAPITemplates indexed by their name.
// The name of the template is set to the key of the ConfigMap.Data map.
func ParseConfigMap(cm corev1.ConfigMap) (map[string]*capiv1.CAPITemplate, error) {
	tm := map[string]*capiv1.CAPITemplate{}

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
func Params(s capiv1.CAPITemplateSpec) ([]string, error) {
	proc := processor.NewSimpleProcessor()
	variables := map[string]bool{}
	for _, v := range s.ResourceTemplates {
		tv, err := proc.GetVariables(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("processing template: %w", err)
		}
		for _, n := range tv {
			variables[n] = true
		}
	}
	var names []string
	for k := range variables {
		names = append(names, k)
	}
	sort.Strings(names)
	return names, nil
}
