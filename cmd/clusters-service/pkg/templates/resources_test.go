package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
    alpha: "true" # this is a comment
`),
		sb(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
`),
		sb(`apiVersion:
kind:
metadata:
  name: testing-md-1
  annotations:
    alpha: "false" # this is a comment
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
  name: testing
  annotations:
    alpha: "true" # this is a comment
    example.com/test: "{\"name\":\"testing\",\"namespace\":\"test-ns\"}"
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
---
apiVersion:
kind:
metadata:
  name: testing-md-1
  annotations:
    alpha: "false" # this is a comment
`
	if diff := cmp.Diff(want, writeMultiDoc(t, updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestInjectJSONAnnotation_non_existing_annotations(t *testing.T) {
	sb := func(s string) []byte {
		return []byte(s)
	}

	raw := [][]byte{
		sb(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
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
  name: testing
  annotations:
    example.com/test: "{\"name\":\"testing\",\"namespace\":\"test-ns\"}"
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

func TestInjectJSONAnnotation_annotated_resources(t *testing.T) {
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
  annotations:
    example.com/test: ""
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
  name: testing
  annotations:
    alpha: "true"
    example.com/test: "{\"name\":\"testing\",\"namespace\":\"test-ns\"}"
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
  annotations:
    example.com/test: "{\"name\":\"testing\",\"namespace\":\"test-ns\"}"
`
	if diff := cmp.Diff(want, writeMultiDoc(t, updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}
