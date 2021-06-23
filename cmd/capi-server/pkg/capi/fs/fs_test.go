package fs

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi"
)

func TestFlavours(t *testing.T) {
	f := New(os.DirFS("testdata"), "flavours")

	got, err := f.Flavours()
	if err != nil {
		t.Fatal(err)
	}
	want := []*capi.Flavour{
		{
			Name:        "cluster-template1",
			Description: "this is test template 1",
			Version:     "1.2.3",
			Params: []capi.Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
		},
		{
			Name:        "cluster-template1",
			Description: "this is test template 1",
			Version:     "2.1.0",
			Params: []capi.Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
		},
		{
			Name:        "cluster-template2",
			Description: "this is test template 2",
			Version:     "1.2.3",
			Params: []capi.Param{
				{
					Name: "AWS_NODE_MACHINE_TYPE",
				},
				{
					Name: "AWS_SSH_KEY_NAME",
				},
				{
					Name: "CLUSTER_NAME",
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("failed to parse flavours:\n%s", diff)
	}
}

func TestFlavours_with_unknown_directory(t *testing.T) {
	f := New(os.DirFS("testdata"), "unknown")
	_, err := f.Flavours()
	assert.EqualError(t, err, "failed to ReadDir in Flavours(): open testdata/unknown: no such file or directory")
}

func TestFlavours_with_error_cases(t *testing.T) {
	errorTests := []struct {
		description string
		dirname     string
		errMsg      string
	}{
		{"invalid yaml", "badtemplates1", "failed to parse: failed to unmarshal badtemplates1/0.0.1/bad_template.yaml: error converting YAML to JSON: yaml: cannot decode !!str `error` as a !!float"},
		{"invalid params", "badtemplates2", "failed to get params from template: processing template: bad substitution"},
	}

	for _, tt := range errorTests {
		t.Run(tt.description, func(t *testing.T) {
			f := New(os.DirFS("testdata"), tt.dirname)
			_, err := f.Flavours()
			assert.EqualError(t, err, tt.errMsg)
		})
	}
}
