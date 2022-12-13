package templates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"google.golang.org/protobuf/testing/protocmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_parseTemplate(t *testing.T) {
	type args struct {
		templateFile string
	}
	tests := []struct {
		name     string
		args     args
		expected *gapiv1.GitOpsTemplate
		wantErr  bool
	}{
		{
			name: "valid template",
			args: args{
				templateFile: "testdata/template.yaml",
			},
			expected: &gapiv1.GitOpsTemplate{
				TypeMeta:   metav1.TypeMeta{Kind: "GitOpsTemplate", APIVersion: "templates.weave.works/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-template", Namespace: "default"},
				Spec: templates.TemplateSpec{
					Description: "This is a sample WGE template to test parsing functionality.",
					Params: []templates.TemplateParam{
						{Name: "CLUSTER_NAME", Description: "Name of the cluster.", Required: false},
						{Name: "RESOURCE_NAME", Description: "Name of the template.", Required: false},
						{Name: "NAMESPACE", Description: "Namespace to create the resource in.", Required: false},
						{Name: "GIT_REPO_NAMESPACE", Description: "Namespace of the configuring git repository object.", Required: false},
						{Name: "GIT_REPO_NAME", Description: "Name of the configuring git repository.", Required: false},
						{Name: "PATH", Description: "Path to the generated resource.", Required: false},
					},
					ResourceTemplates: []templates.ResourceTemplate{
						{
							RawExtension: runtime.RawExtension{
								Raw: []byte(`{"apiVersion":"kustomize.toolkit.fluxcd.io/v1beta2","kind":"Kustomization","metadata":{"name":"${RESOURCE_NAME}","namespace":"${NAMESPACE}"},"spec":{"interval":"1h","path":"${TEMPLATE_PATH}","sourceRef":{"kind":"GitRepository","name":"${GIT_REPO_NAME}","namespace":"${GIT_REPO_NAMESPACE}"}}}`),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			args: args{
				templateFile: "testdata/invalid-template.yaml",
			},
			expected: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTemplate(tt.args.templateFile)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("got an error: %v, while wantErr is: %v", err, tt.wantErr)
				}
			}
			if diff := cmp.Diff(result, tt.expected, protocmp.Transform()); diff != "" {
				t.Fatalf("got: %v, want: %v", result, tt.expected)
			}
		})
	}
}
