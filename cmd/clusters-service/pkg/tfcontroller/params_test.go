package capi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
)

func TestParamsFromSpec(t *testing.T) {
	templateTests := []struct {
		filename string
		want     []Param
	}{
		{
			"testdata/template1.yaml", []Param{
				{Name: "CLUSTER_NAME", Description: "This is used for the cluster naming."},
			},
		},
		{
			"testdata/template2.yaml", []Param{
				{Name: "AWS_NODE_MACHINE_TYPE", Options: []string{"big", "small"}},
				{Name: "AWS_SSH_KEY_NAME", Description: "A description"},
				{Name: "CLUSTER_NAME"},
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

func mustParseFile(t *testing.T, filename string) *capiv1.CAPITemplate {
	t.Helper()
	parsed, err := ParseFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
