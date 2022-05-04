package tfcontroller

import (
	"fmt"

	tapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/tfcontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"
)

// RenderOptFunc is a functional option for Rendering templates.
type RenderOptFunc func(uns *unstructured.Unstructured) error

// InjectPruneAnnotation injects an annotation on everything but Cluster objects
// to instruct flux *not* to prune these objects.
func InjectPruneAnnotation(uns *unstructured.Unstructured) error {
	if uns.GetKind() != "Cluster" {
		ann := uns.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		ann["kustomize.toolkit.fluxcd.io/prune"] = "disabled"
		uns.SetAnnotations(ann)
	}

	return nil
}

// InNamespace is a Render option that updates the object metadata to put it
// into the correct namespace.
func InNamespace(ns string) RenderOptFunc {
	return func(uns *unstructured.Unstructured) error {
		// If not specified set it.
		if uns.GetNamespace() == "" {
			uns.SetNamespace(ns)
		}
		return nil
	}
}

// Render takes template Spec and vars and returns a slice of byte-slices with
// the bodies of the rendered objects.
func Render(spec tapiv1.TFTemplateSpec, vars map[string]string, opts ...RenderOptFunc) ([][]byte, error) {
	proc := processor.NewSimpleProcessor()
	var processed [][]byte
	for _, v := range spec.ResourceTemplates {
		b, err := yaml.JSONToYAML(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
		}

		data, err := proc.Process(b, func(n string) (string, error) {
			if s, ok := vars[n]; ok {
				return s, nil
			}
			return "", fmt.Errorf("variable %s not found", n)
		})
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

func processUnstructured(b []byte, opts ...RenderOptFunc) ([]byte, error) {
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	uns := &unstructured.Unstructured{}
	_, _, err := dec.Decode(b, nil, uns)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the YAML: %w", err)
	}
	for _, o := range opts {
		err := o(uns)
		if err != nil {
			return nil, err
		}
	}
	updated, err := yaml.Marshal(uns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the updated object: %w", err)
	}
	return updated, nil
}
