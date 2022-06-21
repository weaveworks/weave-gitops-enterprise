package templates

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// RenderOptFunc is a functional option for Rendering templates.
type RenderOptFunc func(uns *unstructured.Unstructured) error

// InjectPruneAnnotation injects an annotation on everything
// but Cluster and GitopsCluster objects
// to instruct flux *not* to prune these objects.
func InjectPruneAnnotation(uns *unstructured.Unstructured) error {
	if uns.GetKind() != "Cluster" && uns.GetKind() != "GitopsCluster" {
		// NOTE: This is doing the same thing as uns.GetAnnotations() but with
		// error handling, GetAnnotations is unlikely to change behaviour.
		ann, _, err := unstructured.NestedStringMap(uns.Object, "metadata", "annotations")
		if err != nil {
			return fmt.Errorf("failed trying to inject prune annotation: %w", err)
		}
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
func Render(spec templates.TemplateSpec, vars map[string]string, opts ...RenderOptFunc) ([][]byte, error) {
	var processed [][]byte
	for _, v := range spec.ResourceTemplates {
		b, err := yaml.JSONToYAML(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert back to YAML: %w", err)
		}

		data, err := ProcessTemplate(b, vars)
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

// ProcessTemplate receives a template and a map of values, and substitutes
// the template variables with concrete values.
func ProcessTemplate(template []byte, values map[string]string) ([]byte, error) {
	proc := processor.NewSimpleProcessor()

	rendered, err := proc.Process([]byte(template), func(n string) (string, error) {
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
