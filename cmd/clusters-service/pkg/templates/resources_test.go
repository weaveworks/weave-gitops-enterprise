package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func TestInjectJSONAnnotation(t *testing.T) {
	sb := func(s string) []byte {
		return []byte(s)
	}

	raw := [][]byte{
		sb(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  annotations:
    alpha: "true"
`),
		sb(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
`),
	}

	updated, err := InjectJSONAnnotation(raw, "example.com/test", testData{Name: "testing", Namespace: "test-ns"})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  annotations:
    alpha: "true"
    example.com/test: '{"name":"testing","namespace":"test-ns"}'
  name: testing
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
`
	if diff := cmp.Diff(want, writeMultiDoc(t, updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestInjectJSONAnnotation_no_elements(t *testing.T) {
	raw := [][]byte{}
	updated, err := InjectJSONAnnotation(raw, "example.com/test", testData{Name: "testing", Namespace: "test-ns"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff("", writeMultiDoc(t, updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestInjectJSONAnnotation_bad_annotation(t *testing.T) {
	sb := func(s string) []byte {
		return []byte(s)
	}
	raw := [][]byte{
		sb(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  annotations:
    testing: {{ testing }}
`),
		sb(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
`)}

	_, err := InjectJSONAnnotation(raw, "example.com/test", testData{Name: "testing", Namespace: "test-ns"})
	assert.ErrorContains(t, err, "failed to decode the YAML")
}
