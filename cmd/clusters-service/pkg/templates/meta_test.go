package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParseTemplateTerraformMeta(t *testing.T) {
	parsed := mustParseFile(t, "testdata/tf-controller-multiple.yaml")
	meta, err := ParseTemplateMeta(parsed, TFControllerDisplayNameAnnotation)
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
		Params: []Param{
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
	parsed := mustParseFile(t, "testdata/template3.yaml")
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
		Params: []Param{
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
	parsed, err := ParseBytes([]byte("spec:\n  resourcetemplates:\n   - apiVersion: ${CLUSTER_NAME"), "testing.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseTemplateMeta(parsed, TFControllerDisplayNameAnnotation)
	assert.EqualError(t, err, "failed to get parameters processing template: missing closing brace")
}
