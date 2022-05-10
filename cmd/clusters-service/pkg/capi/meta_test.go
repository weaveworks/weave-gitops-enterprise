package capi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

func TestParseTemplateMeta(t *testing.T) {
	parsed := mustParseFile(t, "testdata/template3.yaml")
	meta, err := ParseTemplateMeta(parsed)
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
		Params: []templates.Param{
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
	parsed, err := templates.ParseBytes([]byte("spec:\n  resourcetemplates:\n   - apiVersion: ${CLUSTER_NAME"), "testing.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseTemplateMeta(&capiv1.CAPITemplate{Template: *parsed})
	assert.EqualError(t, err, "failed to get parameters processing template: missing closing brace")
}

func mustParseFile(t *testing.T, filename string) *capiv1.CAPITemplate {
	t.Helper()
	parsed, err := templates.ParseFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return &capiv1.CAPITemplate{Template: *parsed}
}
