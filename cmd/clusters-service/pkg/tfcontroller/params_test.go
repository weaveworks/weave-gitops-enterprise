package tfcontroller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
)

func TestParamsFromSpec(t *testing.T) {
	templateTests := []struct {
		filename string
		want     []Param
	}{
		{
			"testdata/tf-controller.yaml", []Param{
				{Name: "CLUSTER_NAME", Description: "Name of the cluster."},
				{Name: "GIT_REPO_NAME", Description: "Name of the configuring git repository."},
				{Name: "GIT_REPO_NAMESPACE", Description: "Namespace of the configuring git repository object."},
				{Name: "NAMESPACE", Description: "Namespace to create the Terraform resource in."},
				{Name: "TEMPLATE_NAME", Description: "Name of the template."},
				{Name: "TEMPLATE_PATH", Description: "Path to the generated tf-controller templates."},
			},
		},
		{
			"testdata/tf-controller-2.yaml", []Param{
				{Name: "CLUSTER_NAME", Description: "Name of the cluster."},
			},
		},
	}

	for _, tt := range templateTests {
		t.Run(tt.filename, func(t *testing.T) {
			parsed := mustParseFile(t, tt.filename)
			got, err := ParamsFromSpec(parsed.Spec)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("failed to get params from spec:\n%s", diff)
			}
		})
	}
}

func TestParamsFromSpec_with_bad_template(t *testing.T) {
	parsed := mustParseFile(t, "testdata/bad_template.yaml")
	_, err := ParamsFromSpec(parsed.Spec)
	assert.EqualError(t, err, "failed to get params from template: processing template: missing closing brace")
}

func mustParseFile(t *testing.T, filename string) *apiv1.TFTemplate {
	t.Helper()
	parsed, err := ParseFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
