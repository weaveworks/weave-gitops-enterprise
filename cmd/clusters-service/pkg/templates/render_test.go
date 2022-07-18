package templates

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCAPIRender(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
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
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
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

func TestGitopsRender(t *testing.T) {
	parsed := mustParseFile(t, "testdata/cluster-template.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME":       "testing",
		"GIT_REPO_NAME":      "git-repo",
		"GIT_REPO_NAMESPACE": "git-namespace",
		"NAMESPACE":          "namespace",
		"RESOURCE_NAME":      "test-tf-template",
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
`)
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

func TestInjectPruneAnnotation_invalid_yaml(t *testing.T) {
	raw := []byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: NotCluster
metadata:
  name: testing
  annotations:
    test.annotation: true
`)
	_, err := processUnstructured(raw, InjectPruneAnnotation)

	assert.ErrorContains(t, err, "failed trying to inject prune annotation: .metadata.annotations")
}

func TestRender_InjectPruneAnnotation(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
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

func TestInNamespaceGitOps(t *testing.T) {
	parsed := mustParseFile(t, "testdata/cluster-template-2.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
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
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
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
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
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
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
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
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
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
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing"})
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

func TestTextTemplateRender(t *testing.T) {
	parsed := mustParseFile(t, "testdata/text-template.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing-templating",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing-templating
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
    kind: KubeadmControlPlane
    name: testing-templating-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: AWSCluster
    name: testing-templating
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestTextTemplateRenderConditional(t *testing.T) {
	parsed := mustParseFile(t, "testdata/text-template2.yaml")
	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME":   "testing-templating",
		"TEST_VALUE":     "false",
		"S3_BUCKET_NAME": "test-bucket",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing-templating
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.1.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
    kind: KubeadmControlPlane
    name: testing-templating-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: AWSCluster
    name: testing-templating
  notARealField:
    name: test-bucket-test
`
	if diff := cmp.Diff(want, writeMultiDoc(t, b)); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_unknown_parameter(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")

	processor, err := NewProcessorForTemplate(*parsed)
	if err != nil {
		t.Fatal(err)
	}

	_, err = processor.RenderTemplates(map[string]string{})
	assert.ErrorContains(t, err, "value for variables [CLUSTER_NAME] is not set")
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
