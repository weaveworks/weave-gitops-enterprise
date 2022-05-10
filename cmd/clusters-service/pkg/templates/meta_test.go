package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParseTemplateMeta(t *testing.T) {
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

func TestParseTemplateMeta_bad_parameter(t *testing.T) {
	parsed, err := ParseBytes([]byte("spec:\n  resourcetemplates:\n   - apiVersion: ${CLUSTER_NAME"), "testing.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseTemplateMeta(parsed, TFControllerDisplayNameAnnotation)
	assert.EqualError(t, err, "failed to get parameters processing template: missing closing brace")
}
