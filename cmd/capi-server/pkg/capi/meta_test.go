package capi

import (
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
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
				Params:     []string{"CLUSTER_NAME"},
			},
			{
				Kind:       "AWSMachineTemplate",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
				Params:     []string{"CLUSTER_NAME"},
			},
		},
		Params: []Param{
			{
				Name:        "CLUSTER_NAME",
				Description: "This is used for the cluster naming."},
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
	_, err = ParseTemplateMeta(parsed)
	assert.EqualError(t, err, "failed to get parameters processing template: bad substitution")
}

func readFixture(t *testing.T, fname string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
