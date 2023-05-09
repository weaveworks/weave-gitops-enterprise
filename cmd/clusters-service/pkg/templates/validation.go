package templates

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ValidateRenderedTemplates takes a slice of byte contents and tries to
// validate each as a Kubernetes resource.
func ValidateRenderedTemplates(processedTemplate [][]byte) error {
	for _, v := range processedTemplate {
		dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		uv := &unstructured.Unstructured{}
		if _, _, err := dec.Decode(v, nil, uv); err != nil {
			return fmt.Errorf("failed to unmarshal resourceTemplate: %w", err)
		}

		if name := uv.GetName(); name != "" {
			// Simple validation for now, to expand
			errs := validation.IsDNS1123Subdomain(name)
			if len(errs) > 0 {
				return fmt.Errorf("invalid value for metadata.name: \"%s\", %s", name, strings.Join(errs, ". "))
			}
		}
	}

	return nil
}
