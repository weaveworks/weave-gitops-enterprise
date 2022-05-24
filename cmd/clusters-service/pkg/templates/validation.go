package templates

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/validation"
)

func ValidateRenderedTemplates(processedTemplate [][]byte) error {
	for _, v := range processedTemplate {
		dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		uv := &unstructured.Unstructured{}
		if _, _, err := dec.Decode(v, nil, uv); err != nil {
			return fmt.Errorf("failed to unmarshal resourceTemplate: %w", err)
		}
		// Simple validation for now, to expand
		errs := validation.IsDNS1123Subdomain(uv.GetName())
		if len(errs) > 0 {
			return fmt.Errorf("invalid value for metadata.name: \"%s\", %s", uv.GetName(), strings.Join(errs, ". "))
		}
	}
	return nil
}
