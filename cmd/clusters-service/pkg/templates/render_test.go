package templates

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCAPIRender(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(parsed)
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

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
`),
				[]byte(`apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
`),
			},
			Path: "./clusters/testing/capi-clusters.yaml",
		},
		{
			Data: [][]byte{
				[]byte(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
`),
				[]byte(`apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
spec:
  replicas: 5
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestTextTemplateStringReplace(t *testing.T) {
	processor, err := NewProcessorForTemplate(parseCAPITemplateFromBytes(t, []byte(`---
apiVersion: capi.weave.works/v1alpha2
kind: CAPITemplate
metadata:
  name: cluster-template-1
spec:
  description: this is test template 1
  renderType: templating
  params:
  - name: CLUSTER_NAME
    description: This is used for the cluster naming.
  resourcetemplates:
  - content:
    - apiVersion: cluster.x-k8s.io/v1alpha3
      kind: Cluster
      metadata:
        name: '{{ .params.CLUSTER_NAME | replace "." "-" }}'
    path: "./clusters/{{ .params.CLUSTER_NAME }}/capi-cluster.yaml"
`)))
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing.name",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing-name
`),
			},
			Path: "./clusters/testing.name/capi-cluster.yaml",
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestTextTemplateMissingFunction(t *testing.T) {
	processor, err := NewProcessorForTemplate(parseCAPITemplateFromBytes(t, []byte(`---
apiVersion: capi.weave.works/v1alpha2
kind: CAPITemplate
metadata:
  name: cluster-template-1
spec:
  description: this is test template 1
  renderType: templating
  params:
  - name: CLUSTER_NAME
    description: This is used for the cluster naming.
  resourcetemplates:
  - content:
    - apiVersion: cluster.x-k8s.io/v1alpha3
      kind: Cluster
      metadata:
        name: '{{ env "TESTING" }}'
`)))
	if err != nil {
		t.Fatal(err)
	}

	_, err = processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing.name",
	})
	assert.ErrorContains(t, err, `template: cluster-template-1:4: function "env" not defined`)
}

func TestTextTemplate_metadata(t *testing.T) {
	processor, err := NewProcessorForTemplate(parseCAPITemplateFromBytes(t, []byte(`---
apiVersion: capi.weave.works/v1alpha2
kind: CAPITemplate
metadata:
  name: cluster-template-1
  namespace: test-namespace
spec:
  description: this is test template 1
  renderType: templating
  resourcetemplates:
  - content:
    - apiVersion: cluster.x-k8s.io/v1alpha3
      kind: Cluster
      metadata:
        name: "{{ .template.meta.namespace }}-cluster"
    path: "./clusters/{{ .template.meta.name }}/capi-cluster.yaml"
`)))
	if err != nil {
		t.Fatal(err)
	}

	rendered, err := processor.RenderTemplates(map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: test-namespace-cluster
`),
			},
			Path: "./clusters/cluster-template-1/capi-cluster.yaml",
		},
	}

	if diff := cmp.Diff(want, rendered); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestGitopsRender(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/cluster-template.yaml")
	processor, err := NewProcessorForTemplate(parsed)
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

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
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
`),
			},
			Path: "./clusters/testing/tf.yaml",
		},
	}
	if diff := cmp.Diff(want, b); diff != "" {
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
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(parsed)
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

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
`),
				[]byte(`apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
`),
			},
			Path: "./clusters/testing/capi-clusters.yaml",
		},
		{
			Data: [][]byte{
				[]byte(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: testing-md-0
`),
				[]byte(`apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  annotations:
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: testing-control-plane
spec:
  replicas: 5
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
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
  replicas: 5
`)

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1alpha3", Kind: "Cluster"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "controlplane.cluster.x-k8s.io", Version: "v1alpha4", Kind: "KubeadmControlPlane"}, meta.RESTScopeNamespace)

	updated, err := processUnstructured(raw, InNamespace("new-namespace", mapper))
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
	parsed := parseCAPITemplateFromFile(t, "testdata/cluster-template-2.yaml")
	processor, err := NewProcessorForTemplate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1alpha3", Kind: "Cluster"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "gitops.weave.works", Version: "v1alpha1", Kind: "GitopsCluster"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "tfcontroller.contrib.fluxcd.io", Version: "v1alpha1", Kind: "Terraform"}, meta.RESTScopeNamespace)

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing",
	}, InNamespace("new-namespace", mapper))
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
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
`),
			},
			Path: "./clusters/testing/tf.yaml",
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
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

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1alpha3", Kind: "Cluster"}, meta.RESTScopeNamespace)

	updated, err := processUnstructured(raw, InNamespace("new-namespace", mapper))
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

func TestInNamespaceNonNamespacedResource(t *testing.T) {
	namespace := "testNamespace"
	inputRaw := `
apiVersion: v1
kind: Node
metadata:
  name: testNode
`
	expected := `apiVersion: v1
kind: Node
metadata:
  name: testNode
`

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"}, meta.RESTScopeRoot)

	updated, err := processUnstructured([]byte(inputRaw), InNamespace(namespace, mapper))
	if err != nil {
		t.Fatalf(err.Error())
	}

	if diff := cmp.Diff(expected, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}

}

func TestInNamespaceNoKindMatchError(t *testing.T) {
	namespace := "testNamespace"
	inputRaw := `
apiVersion: v1
kind: Node
metadata:
  name: testNode
`
	expected := `apiVersion: v1
kind: Node
metadata:
  name: testNode
  namespace: testNamespace
`

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})

	updated, err := processUnstructured([]byte(inputRaw), InNamespace(namespace, mapper))
	if err != nil {
		t.Fatalf(err.Error())
	}

	if diff := cmp.Diff(expected, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestRender_in_namespace(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1alpha3", Kind: "Cluster"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "gitops.weave.works", Version: "v1alpha1", Kind: "GitopsCluster"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "infrastructure.cluster.x-k8s.io", Version: "v1alpha3", Kind: "AWSMachineTemplate"}, meta.RESTScopeNamespace)
	mapper.Add(schema.GroupVersionKind{Group: "controlplane.cluster.x-k8s.io", Version: "v1alpha4", Kind: "KubeadmControlPlane"}, meta.RESTScopeNamespace)

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME":                "testing",
		"CONTROL_PLANE_MACHINE_COUNT": "5",
	},
		InNamespace("new-test-namespace", mapper))
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  namespace: new-test-namespace
`),
				[]byte(`apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: testing-gitops
  namespace: new-test-namespace
`),
			},
			Path: "./clusters/testing/capi-clusters.yaml",
		},
		{
			Data: [][]byte{
				[]byte(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: testing-md-0
  namespace: new-test-namespace
`),
				[]byte(`apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
  namespace: new-test-namespace
spec:
  replicas: 5
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_with_options(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")
	processor, err := NewProcessorForTemplate(parsed)
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

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
`),
				[]byte(`apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
`),
			},
			Path: "./clusters/testing/capi-clusters.yaml",
		},
		{
			Data: [][]byte{
				[]byte(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
`),
				[]byte(`apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: just-a-test
  namespace: not-a-real-namespace
spec:
  replicas: 2
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRenderWithCRD(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template0.yaml")
	processor, err := NewProcessorForTemplate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing"})
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
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
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestTextTemplateRender(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/text-template.yaml")
	processor, err := NewProcessorForTemplate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	b, err := processor.RenderTemplates(map[string]string{
		"CLUSTER_NAME": "testing-templating",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
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
`),
			},
			Path: "./clusters/testing-templating/cluster.yaml",
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestTextTemplateRenderConditional(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/text-template2.yaml")
	processor, err := NewProcessorForTemplate(parsed)
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

	want := []RenderedTemplate{
		{
			Data: [][]byte{
				[]byte(`apiVersion: cluster.x-k8s.io/v1alpha3
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
`),
			},
		},
	}

	if diff := cmp.Diff(want, b); diff != "" {
		t.Fatalf("rendering failure:\n%s", diff)
	}
}

func TestRender_unknown_parameter(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")

	processor, err := NewProcessorForTemplate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	_, err = processor.RenderTemplates(map[string]string{})
	assert.ErrorContains(t, err, "missing required parameter: CLUSTER_NAME")
}

func TestInjectLabels(t *testing.T) {
	raw := []byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing
  namespace: new-namespace
  labels:
    com.example/testing: tested
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: KubeadmControlPlane
metadata:
  name: testing-control-plane
  labels:
    com.example/other: tested
spec:
  replicas: 5`)
	updated, err := processUnstructured(raw, InjectLabels(map[string]string{
		"new-label": "test-label",
	}))
	if err != nil {
		t.Fatal(err)
	}

	want := `apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  labels:
    com.example/testing: tested
    new-label: test-label
  name: testing
  namespace: new-namespace
`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Fatalf("rendering with option failed:\n%s", diff)
	}
}

func TestConvertToUnstructured(t *testing.T) {
	raw := [][]byte{[]byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing1
  namespace: default
`), []byte(`
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: testing2
  namespace: default
`)}

	converted, err := ConvertToUnstructured(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []*unstructured.Unstructured{
		{
			Object: map[string]any{
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"kind":       "Cluster",
				"metadata": map[string]any{
					"name":      "testing1",
					"namespace": "default",
				},
			},
		},
		{
			Object: map[string]any{
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"kind":       "Cluster",
				"metadata": map[string]any{
					"name":      "testing2",
					"namespace": "default",
				},
			},
		},
	}
	if diff := cmp.Diff(want, converted); diff != "" {
		t.Fatalf("failed to convert to unstructured:\n%s", diff)
	}
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
