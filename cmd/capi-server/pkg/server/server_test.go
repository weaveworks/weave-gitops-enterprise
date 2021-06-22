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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

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
				makeTemplateConfigMap(),
			},
			expected: []*capiv1.Template{},
		},
		{
			name: "1 template",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: []*capiv1.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Body:        "eyJoZWxsbyI6IiR7Q0xVU1RFUl9OQU1FfSJ9",
					Parameters: []*capiv1.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
			},
		},
		{
			name: "2 templates",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template2", makeTemplate(t, func(ct *apiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			expected: []*capiv1.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Body:        "eyJoZWxsbyI6IiR7Q0xVU1RFUl9OQU1FfSJ9",
					Parameters: []*capiv1.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
				{
					Name:        "cluster-template-2",
					Description: "this is test template 2",
					Body:        "eyJoZWxsbyI6IiR7Q0xVU1RFUl9OQU1FfSJ9",
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
				makeTemplateConfigMap("template1", makeTemplate(t)),
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
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: "hello: test-cluster\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default")

			renderTemplateRequest := &capiv1.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: &capiv1.ParameterValues{
					Values: map[string]string{
						"CLUSTER_NAME": "test-cluster",
					},
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
		corev1.AddToScheme,
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

func makeTemplateConfigMap(s ...string) *corev1.ConfigMap {
	data := make(map[string]string)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = s[i+1]
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-templates",
			Namespace: "default",
		},
		Data: data,
	}
}

func makeTemplate(t *testing.T, opts ...func(*apiv1.CAPITemplate)) string {
	t.Helper()
	ct := &apiv1.CAPITemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CAPITemplate",
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template-1",
		},
		Spec: apiv1.CAPITemplateSpec{
			Description: "this is test template 1",
			Params: []apiv1.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []apiv1.CAPIResourceTemplate{
				{
					RawExtension: rawExtension(`{"hello": "${CLUSTER_NAME}"}`),
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	b, err := yaml.Marshal(ct)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func rawExtension(s string) runtime.RawExtension {
	return runtime.RawExtension{
		Raw: []byte(s),
	}
}
