package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
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
			name:         "no templates",
			clusterState: []runtime.Object{},
			expected:     []*capiv1_protos.Template{},
		},
		{
			name: "1 template",
			clusterState: []runtime.Object{
				makeTemplateWithProvider(t, "AWSCluster"),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-1",
					Description:  "this is test template 1",
					Provider:     "aws",
					TemplateKind: "CAPITemplate",
					Namespace:    "default",
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
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.ObjectMeta.Namespace = "test-ns"
					ct.Spec.Description = "this is test template 2"
					ct.Labels = map[string]string{"weave.works/template-kind": "cluster"}
				}),
				makeCAPITemplate(t),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-1",
					Description:  "this is test template 1",
					Provider:     "",
					TemplateKind: "CAPITemplate",
					Namespace:    "default",
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
					Name:         "cluster-template-2",
					Description:  "this is test template 2",
					Provider:     "",
					TemplateKind: "CAPITemplate",
					Labels:       map[string]string{"weave.works/template-kind": "cluster"},
					Namespace:    "test-ns",
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

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			listTemplatesRequest := &capiv1_protos.ListTemplatesRequest{
				TemplateKind: capiv1.Kind,
			}

			listTemplatesResponse, err := s.ListTemplates(ctx, listTemplatesRequest)
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

func TestListClusterTemplates(t *testing.T) {
	testCases := []struct {
		name             string
		clusterState     []runtime.Object
		expected         []*capiv1_protos.Template
		err              error
		expectedErrorStr string
	}{
		{
			name:     "no configmap",
			err:      errors.New("error getting capitemplate default/cluster-template-1: capitemplates.capi.weave.works \"cluster-template-1\" not found"),
			expected: []*capiv1_protos.Template{},
		},
		{
			name:         "no templates",
			clusterState: []runtime.Object{},
			expected:     []*capiv1_protos.Template{},
		},
		{
			name: "1 template",
			clusterState: []runtime.Object{
				makeClusterTemplateWithProvider(t, "AWSCluster"),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-1",
					Description:  "this is test template 1",
					Provider:     "aws",
					TemplateKind: "GitOpsTemplate",
					Namespace:    "default",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:       "${RESOURCE_NAME}",
							ApiVersion: "fooversion",
							Kind:       "AWSCluster",
							Parameters: []string{"RESOURCE_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "RESOURCE_NAME",
							Description: "This is used for the resource naming.",
						},
					},
				},
			},
		},
		{
			name: "2 templates",
			clusterState: []runtime.Object{
				makeClusterTemplates(t, func(ct *gapiv1.GitOpsTemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.ObjectMeta.Namespace = "test-ns"
					ct.Spec.Description = "this is test template 2"
				}),
				makeClusterTemplates(t),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-1",
					Description:  "this is test template 1",
					Provider:     "",
					TemplateKind: "GitOpsTemplate",
					Namespace:    "default",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:        "${RESOURCE_NAME}",
							DisplayName: "ClusterName",
							ApiVersion:  "fooversion",
							Kind:        "fookind",
							Parameters:  []string{"RESOURCE_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "RESOURCE_NAME",
							Description: "This is used for the resource naming.",
						},
					},
				},
				{
					Name:         "cluster-template-2",
					Description:  "this is test template 2",
					Provider:     "",
					TemplateKind: "GitOpsTemplate",
					Namespace:    "test-ns",
					Objects: []*capiv1_protos.TemplateObject{
						{
							Name:        "${RESOURCE_NAME}",
							DisplayName: "ClusterName",
							ApiVersion:  "fooversion",
							Kind:        "fookind",
							Parameters:  []string{"RESOURCE_NAME"},
						},
					},
					Parameters: []*capiv1_protos.Parameter{
						{
							Name:        "RESOURCE_NAME",
							Description: "This is used for the resource naming.",
						},
					},
				},
			},
		},
	}

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			listTemplatesRequest := &capiv1_protos.ListTemplatesRequest{
				TemplateKind: gapiv1.Kind,
			}

			listTemplatesResponse, err := s.ListTemplates(ctx, listTemplatesRequest)
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
				makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}),
				makeCAPITemplate(t),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-2",
					Description:  "this is test template 2",
					Provider:     "aws",
					TemplateKind: "CAPITemplate",
					Namespace:    "default",
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
				makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}),
				makeCAPITemplate(t),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:         "cluster-template-2",
					Description:  "this is test template 2",
					Provider:     "aws",
					TemplateKind: "CAPITemplate",
					Namespace:    "default",
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
				makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}),
				makeCAPITemplate(t),
			},
			expected: []*capiv1_protos.Template{},
		},
		{
			name:     "Provider name not recognised",
			provider: "foo",
			clusterState: []runtime.Object{
				makeTemplateWithProvider(t, "AWSCluster", func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-2"
					ct.Spec.Description = "this is test template 2"
				}),
				makeCAPITemplate(t),
			},
			err: fmt.Errorf("provider %q is not recognised", "foo"),
		},
	}

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			listTemplatesRequest := new(capiv1_protos.ListTemplatesRequest)
			listTemplatesRequest.Provider = tt.provider

			listTemplatesResponse, err := s.ListTemplates(ctx, listTemplatesRequest)
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
			err:  errors.New("error looking up template cluster-template-1: error getting capitemplate default/cluster-template-1: capitemplates.capi.weave.works \"cluster-template-1\" not found"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
					c.Annotations = map[string]string{"hi": "there"}
					c.Labels = map[string]string{"weave.works/template-kind": "cluster"}
				}),
			},
			expected: &capiv1_protos.Template{
				Name:         "cluster-template-1",
				Annotations:  map[string]string{"hi": "there"},
				Labels:       map[string]string{"weave.works/template-kind": "cluster"},
				Description:  "this is test template 1",
				Provider:     "",
				TemplateKind: "CAPITemplate",
				Namespace:    "default",
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
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})
			getTemplateRes, err := s.GetTemplate(context.Background(), &capiv1_protos.GetTemplateRequest{TemplateName: "cluster-template-1", TemplateNamespace: "default"})
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
			err:  errors.New("error looking up template cluster-template-1: error getting capitemplate default/cluster-template-1: capitemplates.capi.weave.works \"cluster-template-1\" not found"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
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
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			listTemplateParamsRequest := new(capiv1_protos.ListTemplateParamsRequest)
			listTemplateParamsRequest.TemplateName = "cluster-template-1"
			listTemplateParamsRequest.TemplateNamespace = "default"

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
			err:  errors.New("error looking up template cluster-template-1: error getting capitemplate default/cluster-template-1: capitemplates.capi.weave.works \"cluster-template-1\" not found"),
		},
		{
			name: "1 profile",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
					c.Annotations = map[string]string{
						"capi.weave.works/profile-0": "{\"name\": \"profile-a\", \"version\": \"v0.0.1\" }",
					}
				}),
			},
			expected: []*capiv1_protos.TemplateProfile{
				{
					Name:     "profile-a",
					Version:  "v0.0.1",
					Required: true,
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			listTemplateProfilesRequest := new(capiv1_protos.ListTemplateProfilesRequest)
			listTemplateProfilesRequest.TemplateName = "cluster-template-1"
			listTemplateProfilesRequest.TemplateNamespace = "default"

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
		templateKind     string
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
				makeCAPITemplate(t),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "render template with apps",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "render template with optional value",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.Spec.Params = append(ct.Spec.Params, templatesv1.TemplateParam{
						Name:     "OPTIONAL_PARAM",
						Required: false,
					})
					ct.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{{RawExtension: rawExtension(`{
							"apiVersion":"fooversion",
							"kind":"fookind",
							"metadata":{
								"name":"${CLUSTER_NAME}",
								"namespace":"${NAMESPACE}",
								"annotations":{
									"capi.weave.works/display-name":"ClusterName${OPTIONAL_PARAM}"
								}
							}
						}`),
							}},
						},
					}
				}),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "render template with default value",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.Spec.Params = append(ct.Spec.Params, templatesv1.TemplateParam{
						Name:     "OPTIONAL_PARAM",
						Required: true, // Default being set overrides this field
						Default:  "foo",
					})
					ct.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{{RawExtension: rawExtension(`{
							"apiVersion":"fooversion",
							"kind":"fookind",
							"metadata":{
								"name":"${CLUSTER_NAME}",
								"namespace":"${NAMESPACE}",
								"annotations":{
									"capi.weave.works/display-name":"ClusterName${OPTIONAL_PARAM}"
								}
							}
						}`),
							}},
						},
					}
				}),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterNamefoo
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			// some client might send empty credentials objects
			name:             "render template with empty credentials",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			credentials: &capiv1_protos.Credential{
				Group:     "",
				Version:   "",
				Kind:      "",
				Name:      "",
				Namespace: "",
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "render template with credentials",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				u,
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Name = "cluster-template-1"
					ct.Spec.Description = "this is test template 1"
					ct.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{RawExtension: rawExtension(`{
							"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
							"kind": "AWSCluster",
							"metadata": { "name": "boop" }
						}`)},
							},
						},
					}
				}),
			},
			credentials: &capiv1_protos.Credential{
				Group:     "infrastructure.cluster.x-k8s.io",
				Version:   "v1alpha4",
				Kind:      "AWSClusterStaticIdentity",
				Name:      "cred-name",
				Namespace: "cred-namespace",
			},
			expected: `apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSCluster
metadata:
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: boop
  namespace: test-ns
  annotations:
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
spec:
  identityRef:
    kind: AWSClusterStaticIdentity
    name: cred-name
`,
		},
		{
			name:             "enable prune injections",
			pruneEnvVar:      "enabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "enable prune injections via anno",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.ObjectMeta.Annotations = map[string]string{
						"templates.weave.works/inject-prune-annotation": "true",
					}
				}),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
		{
			name:             "enable prune injections with non-CAPI template",
			pruneEnvVar:      "enabled",
			clusterNamespace: "test-ns",
			templateKind:     gapiv1.Kind,
			clusterState: []runtime.Object{
				makeClusterTemplateWithProvider(t, "AWSCluster", func(gt *gapiv1.GitOpsTemplate) {
					gt.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
							"apiVersion":"fooversion",
							"kind":"fookind",
							"metadata":{
								"name": "${CLUSTER_NAME}"
							}
						}`),
								},
							},
						},
					}
				}),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
  annotations:
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
`,
		},
		{
			name:             "render template with renderType: templating",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.Spec.RenderType = "templating"
					ct.Spec.Params = append(ct.Spec.Params, templatesv1.TemplateParam{
						Name:     "OPTIONAL_PARAM",
						Required: false,
					})
					ct.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{{RawExtension: rawExtension(`{
							"apiVersion":"fooversion",
							"kind":"fookind",
							"metadata":{
								"name": "{{ .params.CLUSTER_NAME }}",
								"namespace": "{{ .params.NAMESPACE }}",
								"annotations":{
									"capi.weave.works/display-name":"ClusterName{{ .params.OPTIONAL_PARAM }}"
								}
							}
						}`),
							}},
						},
					}
				}),
			},
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"test-ns/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: test-ns
`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault("inject-prune-annotation", tt.pruneEnvVar)
			viper.SetDefault("capi-clusters-namespace", tt.clusterNamespace)
			viper.SetDefault("capi-templates-namespace", "default")

			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "test-cluster",
				},
				Credentials:       tt.credentials,
				TemplateKind:      tt.templateKind,
				TemplateNamespace: "default",
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
				if diff := cmp.Diff(tt.expected, renderTemplateResponse.RenderedTemplates[0].Content, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestCostEstimation(t *testing.T) {
	// Works very similarly to the render tests, but we need to set up a fake
	// cost estimator server to test the cost estimation functionality.

	testCases := []struct {
		name         string
		clusterState []runtime.Object
		expectedCost *capiv1_protos.CostEstimate
		estimator    estimation.Estimator
		err          error
	}{
		{
			name: "no annotation",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			expectedCost: &capiv1_protos.CostEstimate{
				Message: "no estimate returned",
			},
		},
		{
			name: "has annotation",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.SetAnnotations(map[string]string{
						"templates.weave.works/cost-estimation-enabled": "true",
					})
				}),
			},
			estimator: testEstimator{low: 50, high: 150000, currency: "GBP"},
			expectedCost: &capiv1_protos.CostEstimate{
				Currency: "GBP",
				Range: &capiv1_protos.CostEstimate_Range{
					Low:  50,
					High: 150000,
				},
			},
		},
		{
			name: "estimator errors come back in the response",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.SetAnnotations(map[string]string{
						"templates.weave.works/cost-estimation-enabled": "true",
					})
				}),
			},
			estimator: errorEstimator{err: errors.New("on no.")},
			expectedCost: &capiv1_protos.CostEstimate{
				Message: "failed to calculate estimate for cluster costs: on no.",
			},
		},
		{
			name: "estimator doesn't return anything",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
					ct.SetAnnotations(map[string]string{
						"templates.weave.works/cost-estimation-enabled": "true",
					})
				}),
			},
			estimator: errorEstimator{err: nil},
			expectedCost: &capiv1_protos.CostEstimate{
				Message: "no estimate returned",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault("capi-templates-namespace", "default")

			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
				estimator:    tt.estimator,
			})

			renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "test-cluster",
				},
				TemplateKind:      "CAPITemplate",
				TemplateNamespace: "default",
			}

			renderTemplateResponse, err := s.RenderTemplate(context.Background(), renderTemplateRequest)
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.expectedCost, renderTemplateResponse.CostEstimate, protocmp.Transform()); diff != "" {
				t.Fatalf("cost estimate does not match:\n%s", diff)
			}

		})
	}
}

func TestRenderTemplateWithAppsAndProfiles(t *testing.T) {
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
		expected         *capiv1_protos.RenderTemplateResponse
		err              error
		expectedErrorStr string
		credentials      *capiv1_protos.Credential
		req              *capiv1_protos.RenderTemplateRequest
	}{
		{
			name:             "render template with apps",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			req: &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				TemplateNamespace: "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:            "./apps/capi",
							SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
							TargetNamespace: "foo-ns",
						},
					},
					{
						Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:      "./apps/billing",
							SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
						},
					},
				},
			},
			expected: &capiv1_protos.RenderTemplateResponse{
				RenderedTemplates: []*capiv1_protos.CommitFile{
					{
						Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: ""
  name: dev
  namespace: clusters-namespace
`,
						Path: "clusters-namespace/dev.yaml",
					},
				},
				KustomizationFiles: []*capiv1_protos.CommitFile{
					{
						Path: "clusters/clusters-namespace/dev/apps-capi-flux-system-kustomization.yaml",
						Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-capi
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/capi
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  targetNamespace: foo-ns
status: {}
`,
					},
					{
						Path: "clusters/clusters-namespace/dev/apps-billing-flux-system-kustomization.yaml",
						Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-billing
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/billing
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
status: {}
`,
					},
				},
				ProfileFiles: []*capiv1_protos.CommitFile{},
			},
		},
		{
			name:             "render template with profiles",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			req: &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				TemplateNamespace: "default",
				Profiles: []*capiv1_protos.ProfileValues{
					{
						Name:      "demo-profile",
						Version:   "0.0.1",
						Values:    base64.StdEncoding.EncodeToString([]byte(``)),
						Namespace: "test-system",
					},
				},
			},
			expected: &capiv1_protos.RenderTemplateResponse{
				RenderedTemplates: []*capiv1_protos.CommitFile{
					{
						Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: ""
  name: dev
  namespace: clusters-namespace
`,
						Path: "clusters-namespace/dev.yaml",
					},
				},
				KustomizationFiles: []*capiv1_protos.CommitFile{
					{
						Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
						Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
					},
				},
				ProfileFiles: []*capiv1_protos.CommitFile{
					{
						Path: "clusters/clusters-namespace/dev/profiles.yaml",
						Content: `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weaveworks-charts
  namespace: default
spec:
  interval: 10m0s
  url: http://127.0.0.1:{{ .Port }}/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: demo-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: demo-profile
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
      version: 0.0.1
  install:
    crds: CreateReplace
    createNamespace: true
  interval: 1m0s
  targetNamespace: test-system
  upgrade:
    crds: CreateReplace
  values: {}
status: {}
`,
					},
				},
			},
		},
		{
			name:             "render template with dummy cost estimate data returned",
			pruneEnvVar:      "disabled",
			clusterNamespace: "test-ns",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			req: &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				Profiles:          []*capiv1_protos.ProfileValues{},
				TemplateNamespace: "default",
			},
			expected: &capiv1_protos.RenderTemplateResponse{
				RenderedTemplates: []*capiv1_protos.CommitFile{
					{
						Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    templates.weave.works/created-files: "{\"files\":[\"clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: ""
  name: dev
  namespace: clusters-namespace
`,
						Path: "clusters-namespace/dev.yaml",
					},
				},
				KustomizationFiles: []*capiv1_protos.CommitFile{},
				ProfileFiles:       []*capiv1_protos.CommitFile{},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault("add-bases-kustomization", "disabled")
			viper.SetDefault("inject-prune-annotation", tt.pruneEnvVar)
			viper.SetDefault("capi-clusters-namespace", tt.clusterNamespace)
			viper.SetDefault("capi-repository-clusters-path", "clusters")

			ts := httptest.NewServer(makeServeMux(t))
			hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1.HelmRepository) {
				hr.Name = "weaveworks-charts"
				hr.Namespace = "default"
			})
			tt.clusterState = append(tt.clusterState, hr)
			fakeCache := testNewFakeChartCache(t,
				nsn("management", ""),
				helm.ObjectReference{
					Name:      "weaveworks-charts",
					Namespace: "default",
				},
				[]helm.Chart{})
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
				chartsCache:  fakeCache,
				profileHelmRepository: &types.NamespacedName{
					Name:      "weaveworks-charts",
					Namespace: "default",
				},
			})

			renderTemplateResponse, err := s.RenderTemplate(context.Background(), tt.req)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to read the templates:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected.RenderedTemplates, renderTemplateResponse.RenderedTemplates, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.KustomizationFiles, renderTemplateResponse.KustomizationFiles, protocmp.Transform()); len(renderTemplateResponse.KustomizationFiles) > 0 && diff != "" {
					t.Fatalf("template kustomizations didn't match expected:\n%s", diff)
				}

				if diff := cmp.Diff(prepCommitedFiles(t, ts.URL, tt.expected.ProfileFiles), renderTemplateResponse.ProfileFiles, protocmp.Transform()); len(tt.expected.ProfileFiles) > 0 && diff != "" {
					t.Fatalf("templates profiles didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestRenderTemplate_MissingRequiredVariable(t *testing.T) {
	clusterState := []runtime.Object{
		makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
			ct.Spec.Params[0].Required = true
		}),
	}
	s := createServer(t, serverOptions{
		clusterState: clusterState,
		namespace:    "default",
	})

	renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
		TemplateName: "cluster-template-1",
		Credentials: &capiv1_protos.Credential{
			Group:     "",
			Version:   "",
			Kind:      "",
			Name:      "",
			Namespace: "",
		},
		TemplateNamespace: "default",
	}

	_, err := s.RenderTemplate(context.Background(), renderTemplateRequest)
	if diff := cmp.Diff(err.Error(), "failed to render template with parameter values: error rendering template cluster-template-1, missing required parameter: CLUSTER_NAME"); diff != "" {
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
				makeCAPITemplate(t),
			},
			clusterName: "test-cluster",
			expected: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/created-files: "{\"files\":[\"default/test-cluster.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: test-cluster
  namespace: default
`,
		},
		{
			name: "value contains non alphanumeric",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			clusterName: "t&est-cluster",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "t&est-cluster", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "value does not end alphanumeric",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			clusterName: "test-cluster-",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "test-cluster-", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "value contains uppercase letter",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			clusterName: "Test-Cluster",
			err:         errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "Test-Cluster", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault("capi-templates-namespace", "default")

			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			renderTemplateRequest := &capiv1_protos.RenderTemplateRequest{
				TemplateName: "cluster-template-1",
				Values: map[string]string{
					"CLUSTER_NAME": tt.clusterName,
				},
				TemplateNamespace: "default",
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
				if diff := cmp.Diff(tt.expected, renderTemplateResponse.RenderedTemplates[0].Content, protocmp.Transform()); diff != "" {
					t.Fatalf("templates didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestGetFiles_required_profiles(t *testing.T) {
	viper.SetDefault("runtime-namespace", "flux-system")
	ts := httptest.NewServer(makeServeMux(t))
	hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1.HelmRepository) {
		hr.SetName("weaveworks-charts")
		hr.SetNamespace("flux-system")
	})
	c := createClient(t, hr)

	log := logr.Discard()
	testEstimator := testEstimator{low: 1, high: 2, currency: "USD"}
	getFilesRequest := GetFilesRequest{
		ClusterNamespace: "ns-foo",
		ParameterValues: map[string]string{
			"CLUSTER_NAME": "cluster-foo",
			"NAMESPACE":    "ns-foo",
		},
		Credentials: &capiv1_protos.Credential{
			Group:     "",
			Version:   "",
			Kind:      "",
			Name:      "",
			Namespace: "",
		},
		Profiles:       []*capiv1_protos.ProfileValues{},
		Kustomizations: []*capiv1_protos.Kustomization{},
	}

	expectedPath := "ns-foo/cluster-foo/profiles.yaml"
	expectedTemplateContent := `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weaveworks-charts
  namespace: flux-system
spec:
  interval: 10m0s
  url: {{ .URL }}/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: demo-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: demo-profile
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: flux-system
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`
	expectedContent := simpleTemplate(t, expectedTemplateContent, struct{ URL string }{URL: ts.URL})
	expected := &GetFilesReturn{
		RenderedTemplate: nil,
		ProfileFiles: []git.CommitFile{
			{
				Path:    expectedPath,
				Content: &expectedContent,
			},
		},
		CostEstimate: &capiv1_protos.CostEstimate{
			Currency: "USD",
			Range: &capiv1_protos.CostEstimate_Range{
				Low:  1,
				High: 2,
			},
		},
	}

	fakeChartCache := testNewFakeChartCache(t,
		nsn("cluster-foo", "ns-foo"),
		helm.ObjectReference{
			Name:      "weaveworks-charts",
			Namespace: "flux-system",
		},
		[]helm.Chart{})
	values := []byte("foo: bar")
	profile := fmt.Sprintf(`{"name": "demo-profile", "version": "0.0.1", "values": "%s" }`, values)
	files, err := GetFiles(
		context.TODO(),
		c,
		c.RESTMapper(),
		log,
		testEstimator,
		fakeChartCache,
		types.NamespacedName{Name: "cluster-foo", Namespace: "ns-foo"},
		makeTestTemplateWithProfileAnnotation(
			templatesv1.RenderTypeEnvsubst,
			"capi.weave.works/profile-0",
			profile,
		),
		getFilesRequest,
		nil)
	assert.NoError(t, err)
	if diff := cmp.Diff(expected, files, protocmp.Transform()); diff != "" {
		t.Fatalf("files did not match:\n%s", diff)
	}
}

func makeTemplateWithProvider(t *testing.T, clusterKind string, opts ...func(*capiv1.CAPITemplate)) *capiv1.CAPITemplate {
	t.Helper()
	basicRaw := `
	{
		"apiVersion": "fooversion",
		"kind": "` + clusterKind + `",
		"metadata": {
		  "name": "${CLUSTER_NAME}"
		}
	  }`
	return makeCAPITemplate(t, append(opts, func(c *capiv1.CAPITemplate) {
		c.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
			{
				Content: []templatesv1.ResourceTemplateContent{{RawExtension: rawExtension(basicRaw)}},
			},
		}
	})...)
}

func makeClusterTemplateWithProvider(t *testing.T, clusterKind string, opts ...func(template *gapiv1.GitOpsTemplate)) *gapiv1.GitOpsTemplate {
	t.Helper()
	basicRaw := `
	{
		"apiVersion": "fooversion",
		"kind": "` + clusterKind + `",
		"metadata": {
		  "name": "${RESOURCE_NAME}"
		}
	  }`

	defaultOpts := []func(template *gapiv1.GitOpsTemplate){
		func(c *gapiv1.GitOpsTemplate) {
			c.Spec.ResourceTemplates = []templatesv1.ResourceTemplate{
				{
					Content: []templatesv1.ResourceTemplateContent{{RawExtension: rawExtension(basicRaw)}},
				},
			}
		},
	}

	return makeClusterTemplates(t, append(defaultOpts, opts...)...)
}

type testEstimator struct {
	high     float32
	low      float32
	currency string
}

func (t testEstimator) Estimate(context.Context, []*unstructured.Unstructured) (*estimation.CostEstimate, error) {
	return &estimation.CostEstimate{Low: t.low, High: t.high, Currency: t.currency}, nil
}

type errorEstimator struct {
	err error
}

func (t errorEstimator) Estimate(context.Context, []*unstructured.Unstructured) (*estimation.CostEstimate, error) {
	return nil, t.err
}
