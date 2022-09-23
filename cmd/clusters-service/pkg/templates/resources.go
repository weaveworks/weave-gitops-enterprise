package templates

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InjectJSONAnnotation marshals a value as JSON and adds an annotation to the
// first resource in a slice of bytes, or to any resource that already has the
// annotation.
func InjectJSONAnnotation(resources [][]byte, annotation string, value interface{}) ([][]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data when annotating resource: %w", err)
	}
	updated := make([][]byte, len(resources))
	for i := range resources {
		annotated, err := processUnstructured(resources[i], func(uns *unstructured.Unstructured) error {
			// This doesn't use GetAnnotations because we're processing templates
			// that may not be valid, and GetAnnotations drops invalid data
			// silently.
			ann, _, err := unstructured.NestedStringMap(uns.Object, "metadata", "annotations")
			if err != nil {
				return fmt.Errorf("error getting existing annotations: %w", err)
			}
			if ann == nil {
				ann = make(map[string]string)
			}
			_, ok := ann[annotation]
			if i != 0 && !ok {
				updated[i] = resources[i]
				return nil
			}
			ann[annotation] = string(b)
			uns.SetAnnotations(ann)
			return nil
		})
		if err != nil {
			return nil, err
		}
		updated[i] = annotated
	}

	return updated, nil
}
