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

func TestProcessor_AllParamNames(t *testing.T) {
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
		{
			filename: "testdata/text-template.yaml",
			want: []string{
				"CLUSTER_NAME",
			},
		},
		{
			filename: "testdata/text-template2.yaml",
			want: []string{
				"CLUSTER_NAME",
				"S3_BUCKET_NAME",
				"TEST_VALUE",
			},
		},
		{
			filename: "testdata/text-template3.yaml",
			want: []string{
				"CLUSTER_NAME",
				"CONTROL_PLANE_MACHINE_COUNT",
				"KUBERNETES_VERSION",
				"NAMESPACE",
				"WORKER_MACHINE_COUNT",
			},
		},
		{
			filename: "testdata/template-with-annotation-params.yaml",
			want: []string{
				"CLUSTER_NAME",
				"TEST_PARAMETER",
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
			params, err := proc.AllParamNames()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, params); diff != "" {
				t.Fatalf("failed to extract params:\n%s", diff)
			}
		})
	}
}
