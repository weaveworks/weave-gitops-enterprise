package templates

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
)

var _ Processor = (*TextTemplateProcessor)(nil)
var _ Processor = (*EnvsubstTemplateProcessor)(nil)

func TestNewProcessorForTemplate(t *testing.T) {
	processorTests := []struct {
		renderType string
		template   *gapiv1.GitOpsTemplate
		want       interface{}
		wantErr    string
	}{
		{renderType: templatesv1.RenderTypeEnvsubst, want: NewEnvsubstTemplateProcessor()},
		{renderType: "", want: NewEnvsubstTemplateProcessor()},
		{renderType: templatesv1.RenderTypeTemplating, want: NewTextTemplateProcessor(nil)},
		{renderType: "unknown", wantErr: "unknown template renderType: unknown"},
	}

	for _, tt := range processorTests {
		t.Run("processor for "+tt.renderType, func(t *testing.T) {
			v, err := NewProcessorForTemplate(&gapiv1.GitOpsTemplate{Spec: templatesv1.TemplateSpec{RenderType: tt.renderType}})
			if err != nil {
				if tt.wantErr == "" {
					t.Fatal(err)
				}
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("want error %s, got %s", msg, tt.wantErr)
				}
				return
			}

			if reflect.TypeOf(v.Processor) != reflect.TypeOf(tt.want) {
				t.Fatalf("got %T, want %T", v.Processor, tt.want)
			}
		})
	}
}

func TestProcessor_RenderTemplates(t *testing.T) {
	paramTests := []struct {
		filename string
		params   map[string]string
		want     string
		wantErr  error
	}{
		{
			filename: "testdata/text-template4.yaml",
			params: map[string]string{
				"CLUSTER_NAME":                "testing",
				"NAMESPACE":                   "testing",
				"CONTROL_PLANE_MACHINE_COUNT": "5",
				"KUBERNETES_VERSION":          "1.2.5",
			},
			want:    "---\napiVersion: controlplane.cluster.x-k8s.io/v1beta1\nkind: KubeadmControlPlane\nmetadata:\n  name: testing-control-plane\n  namespace: testing\nspec:\n  machineTemplate:\n    infrastructureRef:\n      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1\n      kind: DockerMachineTemplate\n      name: testing-control-plane\n      namespace: testing\n  replicas: 5\n  version: 1.2.5\n",
			wantErr: nil,
		},
		{
			filename: "testdata/text-template5.yaml",
			params: map[string]string{
				"CLUSTER_NAME":                "testing",
				"NAMESPACE":                   "testing",
				"CONTROL_PLANE_MACHINE_COUNT": "5",
				"KUBERNETES_VERSION":          "1.2.5",
			},
			want:    "---\napiVersion: controlplane.cluster.x-k8s.io/v1beta1\nkind: KubeadmControlPlane\nmetadata:\n  name: testing-control-plane\n  namespace: testing\n  labels:\n    cluster.x-k8s.io/cluster-name: testing #{\"$test-comment\": \"bar\"}\nspec:\n  machineTemplate:\n    infrastructureRef:\n      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1\n      kind: DockerMachineTemplate\n      name: testing-control-plane\n      namespace: testing\n  replicas: 5\n  version: 1.2.5 #{\"$promotion\": \"foo\"}\n",
			wantErr: nil,
		},
		{
			filename: "testdata/text-template6.yaml",
			params: map[string]string{
				"CLUSTER_NAME":                "testing",
				"NAMESPACE":                   "testing",
				"CONTROL_PLANE_MACHINE_COUNT": "5",
				"KUBERNETES_VERSION":          "1.2.5",
			},
			wantErr: errors.New("cannot specify both raw and content in the same resource template: cluster-template-1/default"),
		},
	}

	for _, tt := range paramTests {
		t.Run(tt.filename, func(t *testing.T) {
			c := parseCAPITemplateFromFile(t, tt.filename)
			proc, err := NewProcessorForTemplate(c)
			if err != nil {
				t.Fatal(err)
			}
			result, err := proc.RenderTemplates(tt.params)
			if err != nil {
				if tt.wantErr == nil {
					t.Fatal(err)
				}
				if msg := err.Error(); msg != tt.wantErr.Error() {
					t.Fatalf("want error %s, got %s", msg, tt.wantErr.Error())
				}
				return
			}
			resultData := [][]byte{}
			for _, r := range result {
				resultData = append(resultData, r.Data...)
			}
			t.Logf("%s", writeMultiDoc(t, resultData))
			if diff := cmp.Diff(tt.want, writeMultiDoc(t, resultData)); diff != "" {
				t.Fatalf("failed to render templates:\n%s", diff)
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
					Name:        "AWS_SSH_KEY_NAME",
					Description: "A description",
				},
				{
					Name:    "AWS_NODE_MACHINE_TYPE",
					Options: []string{"big", "small"},
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
					Name:        "TEST_VALUE",
					Description: "boolean string",
					Required:    false,
					Options:     []string{"true", "false"},
				},
				{
					Name: "S3_BUCKET_NAME",
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
					Name:        "NAMESPACE",
					Description: "Namespace to create the cluster in",
					Required:    false,
				},
				{
					Name:        "KUBERNETES_VERSION",
					Description: "Kubernetes version to use for the cluster",
					Required:    false,
					Options:     []string{"1.19.11", "1.21.1", "1.22.0", "1.23.3"},
				},
				{
					Name:        "CONTROL_PLANE_MACHINE_COUNT",
					Description: "Number of control planes",
					Required:    false,
					Options:     []string{"1", "2", "3"},
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
		}, {
			filename: "testdata/template-with-profiles-params.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
				{Name: "HELM_REPO_PATH"},
				{Name: "INTERVAL"},
				{Name: "SPECIAL_CLUSTER_PATH"},
				{Name: "TEST_PARAMETER"},
				{Name: "TEST_PATH"},
			},
		},
		{
			filename: "testdata/template-with-alt-annotation-params.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
		},
		{
			filename: "testdata/text-template5.yaml",
			want: []Param{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
					Required:    true,
					Options:     []string{},
				},
				{Name: "NAMESPACE", Required: true},
				{Name: "CONTROL_PLANE_MACHINE_COUNT", Required: true},
				{Name: "KUBERNETES_VERSION", Required: true},
			},
		},
		{
			filename: "testdata/text-template7.yaml",
			want: []Param{
				{Name: "CLUSTER_NAME"},
				{Name: "NAMESPACE"},
				{Name: "NEW_PARAM"},
				{Name: "OTHER_PARAM"},
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
