package capi

import (
	"fmt"

	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"
)

// Render takes a te
func Render(spec capiv1.CAPITemplateSpec, vars map[string]string) ([][]byte, error) {
	proc := processor.NewSimpleProcessor()
	var processed [][]byte
	for _, v := range spec.ResourceTemplates {
		b, err := proc.Process(v.RawExtension.Raw, func(n string) (string, error) {
			if s, ok := vars[n]; ok {
				return s, nil
			}
			return "", fmt.Errorf("variable %s not found", n)
		})
		if err != nil {
			return nil, fmt.Errorf("processing template: %w", err)
		}
		data, err := yaml.JSONToYAML(b)
		if err != nil {
			return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
		}
		processed = append(processed, data)
	}
	return processed, nil
}
