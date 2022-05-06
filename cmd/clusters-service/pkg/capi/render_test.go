package capi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestRender(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":                "testing",
		"CONTROL_PLANE_MACHINE_COUNT": "5",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
spec:
  replicas: 5
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
	parsed := mustParseFile(t, "testdata/template3.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":                "testing",
		"CONTROL_PLANE_MACHINE_COUNT": "5",
	},
		InjectPruneAnnotation)
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
---
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: testing-md-0
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: testing-control-plane
spec:
  replicas: 5
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestInNamespace(t *testing.T) {
	raw := []byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
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

	want := `apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
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
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  namespace: old-namespace
`)
	updated, err := processUnstructured(raw, InNamespace("new-namespace"))
	if err != nil {
		t.Fatal(err)
	}

	want := `apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  namespace: old-namespace
`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestRender_in_namespace(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":                "testing",
		"CONTROL_PLANE_MACHINE_COUNT": "5",
	},
		InNamespace("new-test-namespace"))
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  namespace: new-test-namespace
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
  namespace: new-test-namespace
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
  namespace: new-test-namespace
spec:
  replicas: 5
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_with_options(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")

	b, err := Render(parsed.Spec, map[string]string{
		"CLUSTER_NAME":                "testing",
		"CONTROL_PLANE_MACHINE_COUNT": "2",
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
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
spec:
  replicas: 2
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRenderWithCRD(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template0.yaml")

	b, err := Render(parsed.Spec, map[string]string{"CLUSTER_NAME": "testing"})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
    kind: KubeadmControlPlane
    name: testing-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: AWSCluster
    name: testing
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_unknown_parameter(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")

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
