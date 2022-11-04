package templates

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"sigs.k8s.io/yaml"
)

func TestParseTemplateTerraformMeta(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/cluster-template-multiple.yaml")
	meta, err := ParseTemplateMeta(parsed, GitOpsTemplateNameAnnotation)
	if err != nil {
		t.Fatal(err)
	}

	want := &TemplateMeta{
		Name:        "sample-wge-tf-controller-template",
		Description: "This is a sample WGE template that will be translated into a tf-controller specific template.",
		Objects: []Object{
			{
				Kind:       "Terraform",
				APIVersion: "tfcontroller.contrib.fluxcd.io/v1alpha1",
				Name:       "${RESOURCE_NAME}-1",
				Params:     []string{"RESOURCE_NAME"},
			},
			{
				Kind:       "Terraform",
				APIVersion: "tfcontroller.contrib.fluxcd.io/v1alpha1",
				Name:       "${RESOURCE_NAME}-2",
				Params:     []string{"RESOURCE_NAME"},
			},
		},
		Params: []templates.TemplateParam{
			{
				Name:        "RESOURCE_NAME",
				Description: "Name of the template.",
				Required:    true,
			},
		},
	}
	if diff := cmp.Diff(want, meta); diff != "" {
		t.Fatalf("failed to parse metadata:\n%s", diff)
	}
}

func TestParseTemplateCAPIMeta(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/template3.yaml")
	meta, err := ParseTemplateMeta(parsed, CAPIDisplayNameAnnotation)
	if err != nil {
		t.Fatal(err)
	}

	want := &TemplateMeta{
		Name:        "cluster-template2",
		Description: "this is a test template",
		Objects: []Object{
			{
				Kind:       "Cluster",
				APIVersion: "cluster.x-k8s.io/v1alpha3",
				Name:       "${CLUSTER_NAME}",
				Params:     []string{"CLUSTER_NAME"},
			},
			{
				Kind:       "GitopsCluster",
				APIVersion: "gitops.weave.works/v1alpha1",
				Name:       "${CLUSTER_NAME}-gitops",
				Params:     []string{"CLUSTER_NAME"},
			},
			{
				Kind:       "AWSMachineTemplate",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
				Params:     []string{"CLUSTER_NAME"},
				Name:       "${CLUSTER_NAME}-md-0",
			},
			{
				Kind:       "KubeadmControlPlane",
				APIVersion: "controlplane.cluster.x-k8s.io/v1alpha4",
				Name:       "${CLUSTER_NAME}-control-plane",
				Params:     []string{"CLUSTER_NAME", "CONTROL_PLANE_MACHINE_COUNT"},
			},
		},
		Params: []templates.TemplateParam{
			{
				Name:        "CLUSTER_NAME",
				Description: "This is used for the cluster naming.",
				Required:    true,
			},
			{
				Name:        "CONTROL_PLANE_MACHINE_COUNT",
				Description: "How many machine replicas to setup.",
				Required:    true,
			},
		},
	}
	if diff := cmp.Diff(want, meta); diff != "" {
		t.Fatalf("failed to parse metadata:\n%s", diff)
	}
}

func TestParseTemplateMeta_bad_parameter(t *testing.T) {
	parsed := parseCAPITemplateFromBytes(t, []byte("apiVersion: capi.weave.works/v1alpha1\nkind: CAPITemplate\nspec:\n  resourcetemplates:\n   - apiVersion: ${CLUSTER_NAME"))

	_, err := ParseTemplateMeta(parsed, GitOpsTemplateNameAnnotation)
	assert.EqualError(t, err, "failed to parse params in template: processing template: missing closing brace")
}

func parseCAPITemplateFromBytes(t *testing.T, b []byte) *capiv1.CAPITemplate {
	var c capiv1.CAPITemplate
	err := yaml.Unmarshal(b, &c)
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}
	return &c
}

func parseCAPITemplateFromFile(t *testing.T, filename string) *capiv1.CAPITemplate {
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return parseCAPITemplateFromBytes(t, b)
}
