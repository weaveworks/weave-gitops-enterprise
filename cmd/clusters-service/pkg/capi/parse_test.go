package capi

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
)

func TestParseFile(t *testing.T) {
	c, err := ParseFile("testdata/template1.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := &capiv1.CAPITemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CAPITemplate",
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template",
		},
		Spec: capiv1.CAPITemplateSpec{
			Description: "this is test template 1",
			Params: []capiv1.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []capiv1.CAPIResourceTemplate{},
		},
	}
	if diff := cmp.Diff(want, c, cmpopts.IgnoreFields(capiv1.CAPITemplateSpec{}, "ResourceTemplates")); diff != "" {
		t.Fatalf("failed to read the template:\n%s", diff)
	}
}

func TestParseFile_with_unknown_file(t *testing.T) {
	_, err := ParseFile("testdata/unknownyaml")
	assert.EqualError(t, err, "failed to read template: open testdata/unknownyaml: no such file or directory")
}

func TestParseFileFromFS(t *testing.T) {
	c, err := ParseFileFromFS(os.DirFS("testdata"), "template2.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := &capiv1.CAPITemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CAPITemplate",
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template2",
		},
		Spec: capiv1.CAPITemplateSpec{
			Description: "this is test template 2",
			Params: []capiv1.TemplateParam{
				{
					Name:        "AWS_SSH_KEY_NAME",
					Description: "A description",
				},
				{
					Name:    "AWS_NODE_MACHINE_TYPE",
					Options: []string{"big", "small"},
				},
			},

			ResourceTemplates: []capiv1.CAPIResourceTemplate{},
		},
	}
	if diff := cmp.Diff(want, c, cmpopts.IgnoreFields(capiv1.CAPITemplateSpec{}, "ResourceTemplates")); diff != "" {
		t.Fatalf("failed to read the template:\n%s", diff)
	}
}

func TestParseFileFromFS_with_unknown_file(t *testing.T) {
	_, err := ParseFileFromFS(os.DirFS("testdata"), "unknown.yaml")
	assert.EqualError(t, err, "failed to read template: open testdata/unknown.yaml: no such file or directory")
}

func TestParseConfigMap(t *testing.T) {
	cmBytes := readFixture(t, "testdata/configmap1.yaml")
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(cmBytes, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	cm := obj.(*corev1.ConfigMap)

	tm, err := ParseConfigMap(*cm)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]*capiv1.CAPITemplate{
		"cluster-template": {
			TypeMeta: metav1.TypeMeta{
				Kind:       "CAPITemplate",
				APIVersion: "capi.weave.works/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster-template",
			},
			Spec: capiv1.CAPITemplateSpec{
				Description: "this is test template 1",
				Params: []capiv1.TemplateParam{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
				},
				ResourceTemplates: []capiv1.CAPIResourceTemplate{},
			},
		},
	}
	if diff := cmp.Diff(want, tm, cmpopts.IgnoreFields(capiv1.CAPITemplateSpec{}, "ResourceTemplates")); diff != "" {
		t.Fatalf("failed to read the template from the configmap:\n%s", diff)
	}
}

func TestParams(t *testing.T) {
	paramTests := []struct {
		filename string
		want     []string
	}{
		{
			filename: "testdata/template1.yaml",
			want:     []string{"CLUSTER_NAME"},
		},
		{
			filename: "testdata/template2.yaml",
			want: []string{
				"AWS_NODE_MACHINE_TYPE",
				"AWS_SSH_KEY_NAME",
				"CLUSTER_NAME",
			},
		},
	}

	for _, tt := range paramTests {
		t.Run(tt.filename, func(t *testing.T) {
			c := mustParseFile(t, tt.filename)
			params, err := Params(c.Spec)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, params); diff != "" {
				t.Fatalf("failed to extract params:\n%s", diff)
			}
		})
	}
}
