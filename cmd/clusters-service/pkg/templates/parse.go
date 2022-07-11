package templates

import (
	"fmt"
	"io/fs"
	"os"

	corev1 "k8s.io/api/core/v1"
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
