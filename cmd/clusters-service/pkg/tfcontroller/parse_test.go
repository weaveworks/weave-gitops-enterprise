package tfcontroller

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	tapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/tfcontroller/v1alpha1"
)

func TestParseFile(t *testing.T) {
	c, err := ParseFile("testdata/tf-controller.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := &tapiv1.TFTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TFTemplate",
			APIVersion: "tfcontroller.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-wge-tf-controller-template",
			Namespace: "default",
		},
		Spec: tapiv1.TFTemplateSpec{
			Description: "This is a sample WGE template that will be translated into a tf-controller specific template.",
			Params: []tapiv1.TemplateParam{
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
			ResourceTemplates: []tapiv1.ResourceTemplate{},
		},
	}
	if diff := cmp.Diff(want, c, cmpopts.IgnoreFields(tapiv1.TFTemplateSpec{}, "ResourceTemplates")); diff != "" {
		t.Fatalf("failed to read the template:\n%s", diff)
	}
}

func TestParseFileResourceTemplate(t *testing.T) {
	c, err := ParseFile("testdata/tf-controller.yaml")
	if err != nil {
		t.Fatal(err)
	}

	tfControllerResultContent, err := os.ReadFile(filepath.Join("testdata", "tf-controller-result.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := &tapiv1.TFTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TFTemplate",
			APIVersion: "tfcontroller.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-wge-tf-controller-template",
			Namespace: "default",
		},
		Spec: tapiv1.TFTemplateSpec{
			Description: "This is a sample WGE template that will be translated into a tf-controller specific template.",
			Params: []tapiv1.TemplateParam{
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
			ResourceTemplates: []tapiv1.ResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: tfControllerResultContent,
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, c); diff != "" {
		t.Fatalf("failed to read the template:\n%s", diff)
	}
}

func TestParseFile_with_unknown_file(t *testing.T) {
	_, err := ParseFile("testdata/unknownyaml")
	assert.EqualError(t, err, "failed to read template: open testdata/unknownyaml: no such file or directory")
}

func TestParseFileFromFS_with_unknown_file(t *testing.T) {
	_, err := ParseFileFromFS(os.DirFS("testdata"), "unknown.yaml")
	assert.EqualError(t, err, "failed to read template: open testdata/unknown.yaml: no such file or directory")
}

func TestParams(t *testing.T) {
	paramTests := []struct {
		filename string
		want     []string
	}{
		{
			filename: "testdata/tf-controller.yaml",
			want: []string{
				"CLUSTER_NAME",
				"GIT_REPO_NAME",
				"GIT_REPO_NAMESPACE",
				"NAMESPACE",
				"RESOURCE_NAME",
				"TEMPLATE_PATH",
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
