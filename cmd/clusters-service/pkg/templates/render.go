package templates

import (
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/yaml"
)

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
// into the correct namespace. It uses the RESTMapper to determine if the object
// is namespaced or not.
func InNamespace(ns string, rm meta.RESTMapper) RenderOptFunc {
	return func(uns *unstructured.Unstructured) error {
		gvk := uns.GroupVersionKind()
		mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			// Check if the error is a NoKindMatchError
			if _, isNoKindMatchError := err.(*meta.NoKindMatchError); isNoKindMatchError {
				log.Printf("kind '%s' not matched due to missing CRD. adding to namespace: %s", gvk.Kind, ns)

				uns.SetNamespace(ns)

				return nil
			}

			return err
		}

		// For root-scoped resource, don't set the namespace.
		if mapping.Scope.Name() == meta.RESTScopeNameRoot {
			return nil
		}
		// For namespaced resources, set the namespace if not specified.
		if uns.GetNamespace() == "" {
			uns.SetNamespace(ns)
		}
		return nil
	}
}

// InjectLabels is a render option that updates the object metadata with the desired labels
func InjectLabels(labels map[string]string) RenderOptFunc {
	return func(uns *unstructured.Unstructured) error {
		existing := uns.GetLabels()
		if existing == nil {
			existing = map[string]string{}
		}
		for k, v := range labels {
			existing[k] = v
		}
		uns.SetLabels(existing)
		return nil
	}
}

// ConvertToUnstructured converts slices of bytes to slices of unstructured
// using a decoding serializer.
func ConvertToUnstructured(values [][]byte) ([]*unstructured.Unstructured, error) {
	converted := make([]*unstructured.Unstructured, len(values))
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	for i := range values {
		uns := &unstructured.Unstructured{}
		_, _, err := dec.Decode(values[i], nil, uns)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the YAML: %w", err)
		}
		converted[i] = uns
	}
	return converted, nil
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
