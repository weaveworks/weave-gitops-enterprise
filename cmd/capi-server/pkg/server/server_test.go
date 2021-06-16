package server

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"google.golang.org/protobuf/testing/protocmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var template = `
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: cluster-template-1
spec:
  description: this is test template 1
  params:
    - name: CLUSTER_NAME
      description: This is used for the cluster naming.
  resourcetemplates:
  - "hello ${CLUSTER_NAME}"
`

var emptyConfigMap = &v1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "capi-templates",
		Namespace: "default",
	},
	Data: map[string]string{},
}

var configMapWithParams = &v1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "capi-templates",
		Namespace: "default",
	},
	Data: map[string]string{
		"template1": template,
	},
}

func TestListTemplates(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         []*capiv1.Template
		err              error
		expectedErrorStr string
	}{
		{
			name:     "no configmap",
			err:      errors.New("configmap capi-templates not found in default namespace"),
			expected: []*capiv1.Template{},
		},
		{
			name: "no templates",
			clusterState: []runtime.Object{
				emptyConfigMap,
			},
			expected: []*capiv1.Template{},
		},
		{
			name: "1 template",
			clusterState: []runtime.Object{
				configMapWithParams,
			},
			expected: []*capiv1.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Body:        "ImhlbGxvICR7Q0xVU1RFUl9OQU1FfSI=",
					Parameters: []*capiv1.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default")

			listTemplatesRequest := new(capiv1.ListTemplatesRequest)

			listTemplatesResponse, err := s.ListTemplates(context.Background(), listTemplatesRequest)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to read the templates:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("failed to read the templates:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, listTemplatesResponse.Templates, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestListTemplateParams(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         []*capiv1.Parameter
		err              error
		expectedErrorStr string
	}{
		{
			name: "1 parameter",
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				configMapWithParams,
			},
			expected: []*capiv1.Parameter{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default")

			listTemplateParamsRequest := new(capiv1.ListTemplateParamsRequest)
			listTemplateParamsRequest.TemplateName = "cluster-template-1"

			listTemplateParamsResponse, err := s.ListTemplateParams(context.Background(), listTemplateParamsRequest)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to read the templates:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, listTemplateParamsResponse.Parameters, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         string
		err              error
		expectedErrorStr string
	}{
		{
			name: "render template",
			clusterState: []runtime.Object{
				configMapWithParams,
			},
			expected: "hello test-cluster\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default")

			renderTemplateRequest := new(capiv1.RenderTemplateRequest)
			renderTemplateRequest.TemplateName = "cluster-template-1"
			renderTemplateRequest.Values = &capiv1.ParameterValues{
				Values: map[string]string{
					"CLUSTER_NAME": "test-cluster",
				},
			}

			renderTemplateResponse, err := s.RenderTemplate(context.Background(), renderTemplateRequest)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to read the templates:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, renderTemplateResponse.RenderedTemplate, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func createServer(clusterState []runtime.Object, configMapName, namespace string) capiv1.ClustersServiceServer {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		apiv1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	s := NewClusterServer(&templates.ConfigMapLibrary{
		Client:        cl,
		ConfigMapName: configMapName,
		Namespace:     namespace,
	})

	return s
}
