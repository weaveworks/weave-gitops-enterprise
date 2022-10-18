package templates

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

var _ Processor = (*TextTemplateProcessor)(nil)
var _ Processor = (*EnvsubstTemplateProcessor)(nil)

func TestNewProcessorForTemplate(t *testing.T) {
	processorTests := []struct {
		renderType string
		want       interface{}
		wantErr    string
	}{
		{renderType: templates.RenderTypeEnvsubst, want: NewEnvsubstTemplateProcessor()},
		{renderType: "", want: NewEnvsubstTemplateProcessor()},
		{renderType: templates.RenderTypeTemplating, want: NewTextTemplateProcessor()},
		{renderType: "unknown", wantErr: "unknown template renderType: unknown"},
	}

	for _, tt := range processorTests {
		t.Run("processor for "+tt.renderType, func(t *testing.T) {
			v, err := NewProcessorForTemplate(&gapiv1.GitOpsTemplate{Spec: templates.TemplateSpec{RenderType: tt.renderType}})

			if err != nil {
				if tt.wantErr == "" {
					t.Fatal(err)
				}
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("want error %s, got %s", msg, tt.wantErr)
				}
				return
			}

			if !reflect.DeepEqual(tt.want, v.Processor) {
				t.Fatalf("got %T, want %T", v, tt.want)
			}
		})
	}
}

func TestProcessor_Params(t *testing.T) {
	paramTests := []struct {
		filename string
		want     []Param
	}{
		{
			filename: "testdata/template1.yaml",
			want:     []Param{{Name: "CLUSTER_NAME", Description: "This is used for the cluster naming."}},
		},
		{
			filename: "testdata/template2.yaml",
			want: []Param{
				{
					Name:    "AWS_NODE_MACHINE_TYPE",
					Options: []string{"big", "small"},
				},
				{
					Name:        "AWS_SSH_KEY_NAME",
					Description: "A description",
				},
				{
					Name: "CLUSTER_NAME",
				},
			},
		},
		{
			filename: "testdata/text-template.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
					Required:    true,
					Options:     []string{},
				},
			},
		},
		{
			filename: "testdata/text-template2.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
					Required:    true,
					Options:     []string{},
				},
				{
					Name: "S3_BUCKET_NAME",
				},
				{
					Name:        "TEST_VALUE",
					Description: "boolean string",
					Required:    false,
					Options:     []string{"true", "false"},
				},
			},
		},
		{
			filename: "testdata/text-template3.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
					Required:    true,
				},
				{
					Name:        "CONTROL_PLANE_MACHINE_COUNT",
					Description: "Number of control planes",
					Required:    false,
					Options:     []string{"1", "2", "3"},
				},
				{
					Name:        "KUBERNETES_VERSION",
					Description: "Kubernetes version to use for the cluster",
					Required:    false,
					Options:     []string{"1.19.11", "1.21.1", "1.22.0", "1.23.3"},
				},
				{
					Name:        "NAMESPACE",
					Description: "Namespace to create the cluster in",
					Required:    false,
				},
				{
					Name:        "WORKER_MACHINE_COUNT",
					Description: "Number of control planes",
					Required:    false,
				},
			},
		},
		{
			filename: "testdata/template-with-annotation-params.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
				{
					Name: "TEST_PARAMETER",
				},
			},
		},
	}

	for _, tt := range paramTests {
		t.Run(tt.filename, func(t *testing.T) {
			c := parseCAPITemplateFromFile(t, tt.filename)
			proc, err := NewProcessorForTemplate(c)
			if err != nil {
				t.Fatal(err)
			}
			params, err := proc.Params()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, params); diff != "" {
				t.Fatalf("failed to extract params:\n%s", diff)
			}
		})
	}
}
