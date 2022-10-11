package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

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

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func RenderAnnotationValues(annotations map[string]string) (map[string]string, error) {
	for k, annot := range annotations {
		if strings.Contains(k, "capi.weave.works/profile-") {
			var annotJson map[string]string
			err := json.Unmarshal([]byte(annot), &annotJson)
			if err != nil {
				return nil, err
			}
			if annotJson["values"] != "" {
				var annotValues map[string]string

				err := yaml.Unmarshal([]byte(annotJson["values"]), &annotValues)
				if err != nil {
					return nil, err
				}
				for resultKey, resultValue := range annotValues {
					regExpMatch, err := regexp.MatchString("[\\$]{\\S+}", resultValue)
					if err != nil {
						return nil, err
					}
					if regExpMatch {
						// Remove ${} surrounding env variable, retrieve its value
						trimmedValue := strings.Trim(resultValue, "${}")
						envValue := GetEnv(trimmedValue, "")

						if envValue != "" {
							newValue := resultKey + ": " + envValue
							annotations[k] = strings.Replace(annot, resultKey+": "+resultValue, newValue, 1)
						} else {
							annotations[k] = strings.Replace(annot, resultKey+": "+resultValue, "", 1)
						}
					}

				}
			}
		}

	}
	return annotations, nil

}
