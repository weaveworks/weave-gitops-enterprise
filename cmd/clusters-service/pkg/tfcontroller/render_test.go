package tfcontroller

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestRender(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":       "testing",
		"GIT_REPO_NAME":      "git-repo",
		"GIT_REPO_NAMESPACE": "git-namespace",
		"NAMESPACE":          "namespace",
		"TEMPLATE_NAME":      "test-tf-template",
		"TEMPLATE_PATH":      "./",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: test-tf-template
  namespace: namespace
spec:
  approvePlan: auto
  interval: 1h
  path: ./
  sourceRef:
    kind: GitRepository
    name: git-repo
    namespace: git-namespace
  vars:
  - name: cluster_identifier
    value: testing
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestInjectPruneAnnotation(t *testing.T) {
	raw := []byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: NotCluster
metadata:
  name: testing
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
spec:
  replicas: 5`)
	updated, err := processUnstructured(raw, InjectPruneAnnotation)
	if err != nil {
		t.Fatal(err)
	}

	want := `apiVersion: cluster.x-k8s.io/v1alpha3
kind: NotCluster
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: testing
`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestRender_InjectPruneAnnotation(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":       "testing",
		"GIT_REPO_NAME":      "git-repo",
		"GIT_REPO_NAMESPACE": "git-namespace",
		"NAMESPACE":          "namespace",
		"TEMPLATE_NAME":      "test-tf-template",
		"TEMPLATE_PATH":      "./",
	}, InjectPruneAnnotation)
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: test-tf-template
  namespace: namespace
spec:
  approvePlan: auto
  interval: 1h
  path: ./
  sourceRef:
    kind: GitRepository
    name: git-repo
    namespace: git-namespace
  vars:
  - name: cluster_identifier
    value: testing
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestInNamespace(t *testing.T) {
	raw := []byte(`
apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: testing
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
spec:
  replicas: 5`)
	updated, err := processUnstructured(raw, InNamespace("new-namespace"))
	if err != nil {
		t.Fatal(err)
	}

	want := `apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: testing
  namespace: new-namespace
`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestInNamespacePreservesExistingNamespace(t *testing.T) {
	raw := []byte(`
apiVersion: tfcontroller.x-k8s.io/v1alpha3
kind: Terraform
metadata:
  name: testing
  namespace: old-namespace
`)
	updated, err := processUnstructured(raw, InNamespace("new-namespace"))
	if err != nil {
		t.Fatal(err)
	}

	want := `apiVersion: tfcontroller.x-k8s.io/v1alpha3
kind: Terraform
metadata:
  name: testing
  namespace: old-namespace
`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestRenderInNamespace(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller-2.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME": "testing",
	}, InNamespace("new-namespace"))
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: test-template
  namespace: new-namespace
spec:
  approvePlan: auto
  interval: 1h
  path: ./
  sourceRef:
    kind: GitRepository
    name: git-repo-name
    namespace: git-repo-namespace
  vars:
  - name: cluster_identifier
    value: testing
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRenderWithOptions(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller-2.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME": "testing",
	},
		func(uns *unstructured.Unstructured) error {
			uns.SetName("just-a-test")
			return nil
		},
		func(uns *unstructured.Unstructured) error {
			uns.SetNamespace("not-a-real-namespace")
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
spec:
  approvePlan: auto
  interval: 1h
  path: ./
  sourceRef:
    kind: GitRepository
    name: git-repo-name
    namespace: git-repo-namespace
  vars:
  - name: cluster_identifier
    value: testing
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_unknown_parameter(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller-2.yaml")

	_, err := Render(parsed.Spec, map[string]string{})
	assert.EqualError(t, err, "processing template: value for variables [CLUSTER_NAME] is not set. Please set the value using os environment variables or the clusterctl config file")
}

func writeMultiDoc(t *testing.T, objs [][]byte) string {
	t.Helper()
	var out bytes.Buffer
	for _, v := range objs {
		if _, err := out.Write([]byte("---\n")); err != nil {
			t.Fatal(err)
		}
		if _, err := out.Write(v); err != nil {
			t.Fatal(err)
		}
	}
	return out.String()
}
