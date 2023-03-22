package templates

import (
	"encoding/json"
	"fmt"
	"strings"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
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
		// parse resource as yaml and get the top level keys
		obj, err := kyaml.Parse(string(resources[i]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse object: %v %v", string(resources[i]), err)
		}

		// get the kind of the resource and apiVersion
		kindNode, err := obj.Pipe(kyaml.Get("kind"))
		if err != nil {
			return nil, fmt.Errorf("failed to get kind: %v", err)
		}

		kind, err := kindNode.String()
		if err != nil {
			return nil, fmt.Errorf("failed to convert kind to string: %v", err)
		}

		apiVersionNode, err := obj.Pipe(kyaml.Get("apiVersion"))
		if err != nil {
			return nil, fmt.Errorf("failed to get apiVersion: %v", err)
		}

		apiVersion, err := apiVersionNode.String()
		if err != nil {
			return nil, fmt.Errorf("failed to convert apiVersion to string: %v", err)
		}

		// If the kind and apiVersion are not empty, add the annotation
		trimmedKind := strings.TrimSpace(kind)
		trimmedAPIVersion := strings.TrimSpace(apiVersion)

		if trimmedKind != "" && trimmedAPIVersion != "" {
			metadataNode, err := obj.Pipe(kyaml.Get("metadata"))
			if err != nil {
				return nil, fmt.Errorf("failed to get metadata: %v", err)
			}

			annotationsNode, err := metadataNode.Pipe(kyaml.Get("annotations"))
			if err != nil {
				return nil, fmt.Errorf("failed to get annotations: %v", err)
			}

			// If the annotations node doesn't exist and we are in the first resource,
			// create it and add it to the metadata node
			if annotationsNode == nil && i == 0 {
				annotationsNode = kyaml.NewMapRNode(nil)
				err = metadataNode.PipeE(kyaml.SetField("annotations", annotationsNode))
				if err != nil {
					return nil, fmt.Errorf("failed to add annotations: %v", err)
				}
			}

			err = annotationsNode.PipeE(kyaml.SetField(annotation, kyaml.NewScalarRNode(string(b))))
			if err != nil {
				return nil, fmt.Errorf("failed to add annotation: %v", err)
			}

			updated[i] = []byte(obj.MustString())
		} else {
			updated[i] = resources[i]
		}
	}

	return updated, nil
}
