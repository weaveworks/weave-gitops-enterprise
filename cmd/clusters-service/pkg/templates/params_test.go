package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParamsFromTemplate(t *testing.T) {
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
				{Name: "AWS_SSH_KEY_NAME", Description: "A description"},
				{Name: "AWS_NODE_MACHINE_TYPE", Options: []string{"big", "small"}},
				{Name: "CLUSTER_NAME"},
			},
		},
		{
			"testdata/text-template.yaml", []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
					Required:    true,
					Options:     []string{},
				},
			},
		},
	}

	for _, tt := range templateTests {
		t.Run(tt.filename, func(t *testing.T) {
			parsed := parseCAPITemplateFromFile(t, tt.filename)
			got, err := ParamsFromTemplate(parsed)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("failed to get params from spec:\n%s", diff)
			}
		})
	}
}

func TestParamsFromTemplate_with_bad_template(t *testing.T) {
	parsed := parseCAPITemplateFromFile(t, "testdata/bad_template.yaml")
	_, err := ParamsFromTemplate(parsed)
	assert.EqualError(t, err, "failed to get params from template: processing template: missing closing brace")
}
