package templates

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InjectJSONAnnotation marshals a value as JSON and adds an annotation to the
// first resource in a slice of bytes.
func InjectJSONAnnotation(resources [][]byte, annotation string, value interface{}) ([][]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data when annotating resource: %w", err)
	}
	updated := make([][]byte, len(resources))
	for i := range resources {
		if i != 0 {
			updated[i] = resources[i]
			continue
		}
		annotated, err := processUnstructured(resources[i], func(uns *unstructured.Unstructured) error {
			ann, _, err := unstructured.NestedStringMap(uns.Object, "metadata", "annotations")
			if err != nil {
				return fmt.Errorf("error getting existing annotations: %w", err)
			}
			if ann == nil {
				ann = make(map[string]string)
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
