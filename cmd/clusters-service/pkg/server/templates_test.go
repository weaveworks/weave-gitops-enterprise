package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestListTemplates(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         []*capiv1_protos.Template
		err              error
		expectedErrorStr string
	}{
		{
			name:     "no configmap",
			err:      errors.New("configmap capi-templates not found in default namespace: configmaps \"capi-templates\" not found"),
			expected: []*capiv1_protos.Template{},
		},
		{
			name: "no templates",
			clusterState: []runtime.Object{
				makeTemplateConfigMap(),
			},
			expected: []*capiv1_protos.Template{},
		},
		{
			name: "1 template",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplateWithProvider(t, "AWSCluster")),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Provider:    "aws",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:       string("${CLUSTER_NAME}"),
							ApiVersion: "fooversion",
							Kind:       "AWSCluster",
							Parameters: []string{"CLUSTER_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
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
				makeTemplateConfigMap("template2", makeTemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Provider:    "",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:        string("${CLUSTER_NAME}"),
							DisplayName: string("ClusterName"),
							ApiVersion:  "fooversion",
							Kind:        "fookind",
							Parameters:  []string{"CLUSTER_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
				{
					Name:        "cluster-template-2",
					Description: "this is test template 2",
					Provider:    "",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:        string("${CLUSTER_NAME}"),
							DisplayName: string("ClusterName"),
							ApiVersion:  "fooversion",
							Kind:        "fookind",
							Parameters:  []string{"CLUSTER_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
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
			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			listTemplatesRequest := new(capiv1_protos.ListTemplatesRequest)

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

func TestListTemplates_FilterByProvider(t *testing.T) {
	testCases := []struct {
		name         string
		provider     string
		clusterState []runtime.Object
		expected     []*capiv1_protos.Template
		err          error
	}{
		{
			name:     "Provider name with upper case letters",
			provider: "AWS",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template2", makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:        "cluster-template-2",
					Description: "this is test template 2",
					Provider:    "aws",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:       string("${CLUSTER_NAME}"),
							ApiVersion: "fooversion",
							Kind:       "AWSCluster",
							Parameters: []string{"CLUSTER_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
			},
		},
		{
			name:     "Provider name with lower case letters",
			provider: "aws",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template2", makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:        "cluster-template-2",
					Description: "this is test template 2",
					Provider:    "aws",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:       string("${CLUSTER_NAME}"),
							ApiVersion: "fooversion",
							Kind:       "AWSCluster",
							Parameters: []string{"CLUSTER_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "CLUSTER_NAME",
							Description: "This is used for the cluster naming.",
						},
					},
				},
			},
		},
		{
			name:     "Provider name with no templates",
			provider: "Azure",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template2", makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Template{},
		},
		{
			name:     "Provider name not recognised",
			provider: "foo",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template2", makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}), "template1", makeTemplate(t)),
			},
			err: fmt.Errorf("provider %q is not recognised", "foo"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			listTemplatesRequest := new(capiv1_protos.ListTemplatesRequest)
			listTemplatesRequest.Provider = tt.provider

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

func TestGetTemplate(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         *capiv1_protos.Template
		err              error
		expectedErrorStr string
	}{
		{
			name: "No templates",
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace: configmaps \"capi-templates\" not found"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t, func(c *capiv1.CAPITemplate) {
					c.Annotations = map[string]string{"hi": "there"}
				})),
			},
			expected: &capiv1_protos.Template{
				Name:        "cluster-template-1",
				Annotations: map[string]string{"hi": "there"},
				Description: "this is test template 1",
				Provider:    "",
				Objects: []*capiv1_protos.TemplateObject{
					{
						Name:        string("${CLUSTER_NAME}"),
						DisplayName: string("ClusterName"),
						ApiVersion:  "fooversion",
						Kind:        "fookind",
						Parameters:  []string{"CLUSTER_NAME"},
					},
				},
				Parameters: []*capiv1_protos.Parameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)
			getTemplateRes, err := s.GetTemplate(context.Background(), &capiv1_protos.GetTemplateRequest{TemplateName: "cluster-template-1"})
			if err != nil && tt.err == nil {
				t.Fatalf("failed to read the templates:\n%s", err)
			} else if err != nil && tt.err != nil {
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, getTemplateRes.Template, protocmp.Transform()); diff != "" {
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
		expected         []*capiv1_protos.Parameter
		err              error
		expectedErrorStr string
	}{
		{
			name: "1 parameter err",
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace: configmaps \"capi-templates\" not found"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Parameter{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			listTemplateParamsRequest := new(capiv1_protos.ListTemplateParamsRequest)
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

func TestListTemplateProfiles(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         []*capiv1_protos.TemplateProfile
		err              error
		expectedErrorStr string
	}{
		{
			name: "1 profile err",
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace: configmaps \"capi-templates\" not found"),
		},
		{
			name: "1 profile",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t, func(c *capiv1.CAPITemplate) {
					c.Annotations = map[string]string{
						"capi.weave.works/profile-0": "{\"name\": \"profile-a\", \"version\": \"v0.0.1\" }",
					}
				})),
			},
			expected: []*capiv1_protos.TemplateProfile{
				{
					Name:    "profile-a",
					Version: "v0.0.1",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			listTemplateProfilesRequest := new(capiv1_protos.ListTemplateProfilesRequest)
			listTemplateProfilesRequest.TemplateName = "cluster-template-1"

			listTemplateProfilesResponse, err := s.ListTemplateProfiles(context.Background(), listTemplateProfilesRequest)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to read the profiles:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, listTemplateProfilesResponse.Profiles, protocmp.Transform()); diff != "" {
					t.Fatalf("profiles didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "cred-name",
			"namespace": "cred-namespace",
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AWSClusterStaticIdentity",
		Version: "v1alpha4",
	})

	testCases := []struct {
		name             string
		pruneEnvVar      string
		clusterNamespace string
		clusterState     []runtime.Object
		expected         string
		err              error
		expectedErrorStr string
		credentials      *capiv1_protos.Credential
	}{
		{
			name:             "render template",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n  name: test-cluster\n  namespace: test-ns\n",
		},
		{
			// some client might send empty credentials objects
			name:             "render template with empty credentials",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			credentials: &capiv1_protos.Credential{
				Group:     "",
				Version:   "",
				Kind:      "",
				Name:      "",
				Namespace: "",
			},
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n  name: test-cluster\n  namespace: test-ns\n",
		},
		{
			name:             "render template with credentials",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				u,
				makeTemplateConfigMap("template1",
					makeTemplate(t, func(ct *capiv1.CAPITemplate) {
						ct.ObjectMeta.Name = "cluster-template-1"
						ct.Spec.Description = "this is test template 1"
						ct.Spec.ResourceTemplates = []capiv1.CAPIResourceTemplate{
							{
								RawExtension: rawExtension(`{
							"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
							"kind": "AWSCluster",
							"metadata": { "name": "boop" }
						}`),
							},
						}
					}),
				),
			},
			credentials: &capiv1_protos.Credential{
				Group:     "infrastructure.cluster.x-k8s.io",
				Version:   "v1alpha4",
				Kind:      "AWSClusterStaticIdentity",
				Name:      "cred-name",
				Namespace: "cred-namespace",
			},
			expected: "apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4\nkind: AWSCluster\nmetadata:\n  name: boop\n  namespace: test-ns\nspec:\n  identityRef:\n    kind: AWSClusterStaticIdentity\n    name: cred-name\n",
		},
		{
			name:             "enable prune injections",
			pruneEnvVar:      "enabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n    kustomize.toolkit.fluxcd.io/prune: disabled\n  name: test-cluster\n  namespace: test-ns\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault("inject-prune-annotation", tt.pruneEnvVar)
			viper.SetDefault("capi-clusters-namespace", tt.clusterNamespace)

			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "test-cluster",
				},
				Credentials: tt.credentials,
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

func TestRenderTemplate_MissingVariables(t *testing.T) {
	clusterState := []runtime.Object{
		makeTemplateConfigMap("template1", makeTemplate(t)),
	}
	s := createServer(t, clusterState, "capi-templates", "default", nil, nil, "", nil)

	renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
		TemplateName: "cluster-template-1",
		Credentials: &capiv1_protos.Credential{
			Group:     "",
			Version:   "",
			Kind:      "",
			Name:      "",
			Namespace: "",
		},
	}

	_, err := s.RenderTemplate(context.Background(), renderTemplateRequest)
	if diff := cmp.Diff(err.Error(), "error rendering template cluster-template-1, processing template: failed to process template values: value for variables [CLUSTER_NAME] is not set. Please set the value using os environment variables or the clusterctl config file"); diff != "" {
		t.Fatalf("got the wrong error:\n%s", diff)
	}
}

func TestRenderTemplate_ValidateVariables(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		expected     string
		err          error
		clusterName  string
	}{
		{
			name: "valid value",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "test-cluster",
			expected:    "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n    kustomize.toolkit.fluxcd.io/prune: disabled\n  name: test-cluster\n",
		},
		{
			name: "value contains non alphanumeric",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "t&est-cluster",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "t&est-cluster", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "value does not end alphanumeric",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "test-cluster-",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "test-cluster-", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "value contains uppercase letter",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "Test-Cluster",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "Test-Cluster", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			s := createServer(t, tt.clusterState, "capi-templates", "default", nil, nil, "", nil)

			renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": tt.clusterName,
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

func TestGetProfilesFromTemplate(t *testing.T) {
	annotations := map[string]string{
		"capi.weave.works/profile-0": "{\"name\": \"profile-a\", \"version\": \"v0.0.1\" }",
	}

	expected := []*capiv1_protos.TemplateProfile{
		{
			Name:    "profile-a",
			Version: "v0.0.1",
		},
	}

	result, err := getProfilesFromTemplate(annotations)
	assert.NoError(t, err)

	assert.Equal(t, result, expected)
}

func makeTemplateWithProvider(t *testing.T, clusterKind string, opts ...func(*capiv1.CAPITemplate)) string {
	t.Helper()
	basicRaw := `
	{
		"apiVersion": "fooversion",
		"kind": "` + clusterKind + `",
		"metadata": {
		  "name": "${CLUSTER_NAME}"
		}
	  }`
	return makeTemplate(t, append(opts, func(c *capiv1.CAPITemplate) {
		c.Spec.ResourceTemplates = []capiv1.CAPIResourceTemplate{
			{
				RawExtension: rawExtension(basicRaw),
			},
		}
	})...)
}
