package templates

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

func TestParseFile(t *testing.T) {
	c, err := ParseFile("testdata/template1.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := &templates.Template{
		TypeMeta: metav1.TypeMeta{
			Kind:       capiv1.Kind,
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template",
		},
		Spec: templates.TemplateSpec{
			Description: "this is test template 1",
			Params: []templates.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []templates.ResourceTemplate{},
		},
	}
	if diff := cmp.Diff(want, c, cmpopts.IgnoreFields(templates.TemplateSpec{}, "ResourceTemplates")); diff != "" {
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

	want := &templates.Template{
		TypeMeta: metav1.TypeMeta{
			Kind:       capiv1.Kind,
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template2",
		},
		Spec: templates.TemplateSpec{
			Description: "this is test template 2",
			Params: []templates.TemplateParam{
				{
					Name:        "AWS_SSH_KEY_NAME",
					Description: "A description",
				},
				{
					Name:    "AWS_NODE_MACHINE_TYPE",
					Options: []string{"big", "small"},
				},
			},

			ResourceTemplates: []templates.ResourceTemplate{},
		},
	}
	if diff := cmp.Diff(want, c, cmpopts.IgnoreFields(templates.TemplateSpec{}, "ResourceTemplates")); diff != "" {
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

	want := map[string]*templates.Template{
		"cluster-template": {
			TypeMeta: metav1.TypeMeta{
				Kind:       capiv1.Kind,
				APIVersion: "capi.weave.works/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster-template",
			},
			Spec: templates.TemplateSpec{
				Description: "this is test template 1",
				Params: []templates.TemplateParam{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
				},
				ResourceTemplates: []templates.ResourceTemplate{},
			},
		},
	}
	if diff := cmp.Diff(want, tm, cmpopts.IgnoreFields(templates.TemplateSpec{}, "ResourceTemplates")); diff != "" {
		t.Fatalf("failed to read the template from the configmap:\n%s", diff)
	}
}

func TestParseFileResourceTemplate(t *testing.T) {
	c, err := ParseFile("testdata/cluster-template.yaml")
	if err != nil {
		t.Fatal(err)
	}

	tfControllerResultContent, err := os.ReadFile(filepath.Join("testdata", "cluster-template-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	tfControllerResultContent = bytes.TrimSuffix(tfControllerResultContent, []byte("\n"))

	want := &gapiv1.GitOpsTemplate{
		Template: templates.Template{
			TypeMeta: metav1.TypeMeta{
				Kind:       "GitOpsTemplate",
				APIVersion: "clustertemplates.weave.works/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sample-wge-tf-controller-template",
				Namespace: "default",
			},
			Spec: templates.TemplateSpec{
				Description: "This is a sample WGE template that will be translated into a tf-controller specific template.",
				Params: []templates.TemplateParam{
					{
						Name:        "CLUSTER_NAME",
						Description: "Name of the cluster.",
					},
					{
						Name:        "RESOURCE_NAME",
						Description: "Name of the template.",
					},
					{
						Name:        "NAMESPACE",
						Description: "Namespace to create the Terraform resource in.",
					},
					{
						Name:        "GIT_REPO_NAMESPACE",
						Description: "Namespace of the configuring git repository object.",
					},
					{
						Name:        "GIT_REPO_NAME",
						Description: "Name of the configuring git repository.",
					},
					{
						Name:        "TEMPLATE_PATH",
						Description: "Path to the generated tf-controller templates.",
					},
				},
				ResourceTemplates: []templates.ResourceTemplate{
					{
						RawExtension: runtime.RawExtension{
							Raw: tfControllerResultContent,
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, &gapiv1.GitOpsTemplate{Template: *c}); diff != "" {
		t.Fatalf("failed to read the template:\n%s", diff)
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

func readFixture(t *testing.T, fname string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
