package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
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
			err:      errors.New("configmap capi-templates not found in default namespace"),
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
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")

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
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")

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
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace"),
		},
		{
			name: "1 parameter",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: &capiv1_protos.Template{
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
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")
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
			err:  errors.New("error looking up template cluster-template-1: configmap capi-templates not found in default namespace"),
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
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")

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
		clusterState     []runtime.Object
		expected         string
		err              error
		expectedErrorStr string
		credentials      *capiv1_protos.Credential
	}{
		{
			name: "render template",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n  name: test-cluster\n",
		},
		{
			// some client might send empty credentials objects
			name: "render template with empty credentials",
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
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n  name: test-cluster\n",
		},
		{
			name: "render template with credentials",
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
			expected: "apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4\nkind: AWSCluster\nmetadata:\n  name: boop\nspec:\n  identityRef:\n    kind: AWSClusterStaticIdentity\n    name: cred-name\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")

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
	s := createServer(clusterState, "capi-templates", "default", nil, nil, "")

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
	if diff := cmp.Diff(err.Error(), "error rendering template cluster-template-1 due to missing variables: [CLUSTER_NAME]"); diff != "" {
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
			expected:    "apiVersion: fooversion\nkind: fookind\nmetadata:\n  annotations:\n    capi.weave.works/display-name: ClusterName\n  name: test-cluster\n",
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
			s := createServer(tt.clusterState, "capi-templates", "default", nil, nil, "")

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

func TestCreatePullRequest(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		provider     git.Provider
		req          *capiv1_protos.CreatePullRequestRequest
		expected     string
		err          error
		dbRows       int
	}{
		{
			name:   "validation errors",
			req:    &capiv1_protos.CreatePullRequestRequest{},
			err:    errors.New("6 errors occurred:\ntemplate name must be specified\nparameter values must be specified\nhead branch must be specified\ntitle must be specified\ndescription must be specified\ncommit message must be specified"),
			dbRows: 0,
		},
		{
			name: "name validation errors",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo bar bad name",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			err:    errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "foo bar bad name", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
			dbRows: 0,
		},
		{
			name: "pull request failed",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			provider: NewFakeGitProvider("", nil, errors.New("oops")),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			dbRows: 0,
			err:    errors.New(`unable to create pull request and cluster rows for "cluster-template-1": oops`),
		},
		{
			name: "create pull request",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			dbRows:   1,
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := createDatabase(t)
			s := createServer(tt.clusterState, "capi-templates", "default", tt.provider, db, "")

			// request
			createPullRequestResponse, err := s.CreatePullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, createPullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}
			}

			// Check the db looks good
			var clusters []models.Cluster
			tx := db.Find(&clusters)
			if tx.Error != nil {
				t.Fatalf("error querying db:\n%v", tx.Error)
			}
			if diff := cmp.Diff(len(clusters), tt.dbRows); diff != "" {
				t.Fatalf("Rows mismatch:\n%s\nwas: %d", diff, len(clusters))
			}
		})
	}
}

func TestGetKubeconfig(t *testing.T) {
	testCases := []struct {
		name                    string
		clusterState            []runtime.Object
		clusterObjectsNamespace string // Namespace that cluster objects are created in
		req                     *capiv1_protos.GetKubeconfigRequest
		ctx                     context.Context
		expected                []byte
		err                     error
	}{
		{
			name: "get kubeconfig as JSON",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(fmt.Sprintf(`{"kubeconfig":"%s"}`, base64.StdEncoding.EncodeToString([]byte("foo")))),
		},
		{
			name: "get kubeconfig as binary",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("accept", "application/octet-stream")),
			expected: []byte("foo"),
		},
		{
			name:                    "secret not found",
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("unable to get secret \"dev-kubeconfig\" for Kubeconfig: secrets \"dev-kubeconfig\" not found"),
		},
		{
			name: "secret found but is missing key",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "val", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("secret \"default/dev-kubeconfig\" was found but is missing key \"value\""),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			db := createDatabase(t)
			gp := NewFakeGitProvider("", nil, nil)
			s := createServer(tt.clusterState, "capi-templates", "default", gp, db, tt.clusterObjectsNamespace)

			res, err := s.GetKubeconfig(tt.ctx, tt.req)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get the kubeconfig:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, res.Data, protocmp.Transform()); diff != "" {
					t.Fatalf("kubeconfig didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteClustersPullRequest(t *testing.T) {
	testCases := []struct {
		name     string
		dbState  []interface{}
		provider git.Provider
		req      *capiv1_protos.DeleteClustersPullRequestRequest
		expected string
		err      error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.DeleteClustersPullRequestRequest{},
			err:  errors.New("5 errors occurred:\nat least one cluster name must be specified\nhead branch must be specified\ntitle must be specified\ndescription must be specified\ncommit message must be specified"),
		},
		{
			name:     "cluster does not exist",
			dbState:  []interface{}{},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
				ClusterNames:  []string{"foo"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			err: gorm.ErrRecordNotFound,
		},
		{
			name: "create delete pull request",
			dbState: []interface{}{
				&models.Cluster{Name: "foo", Token: "foo-token"},
				&models.Cluster{Name: "bar", Token: "bar-token"},
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
				ClusterNames:  []string{"foo", "bar"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := createDatabase(t)
			s := createServer([]runtime.Object{}, "capi-templates", "default", tt.provider, db, "")
			for _, o := range tt.dbState {
				db.Create(o)
			}

			// delete request
			deletePullRequestResponse, err := s.DeleteClustersPullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, deletePullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}

				var clusters []models.Cluster
				db.Preload(clause.Associations).Find(&clusters)
				for _, cluster := range clusters {
					if len(cluster.PullRequests) != 1 {
						t.Fatalf("got the wrong number of pull requests:%d", len(cluster.PullRequests))
					}
					if cluster.PullRequests[0].Type != "delete" {
						t.Fatalf("got the wrong type of pull request:%s", cluster.PullRequests[0].Type)
					}
				}
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name     string
		template *capiv1.CAPITemplate
		provider string
	}{
		{
			name: "AWSCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSCluster"
							}`),
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AWSManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSManagedCluster"
							}`),
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AzureCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureCluster"
							}`),
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "AzureManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureManagedCluster"
							}`),
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "DOCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "DOCluster"
							}`),
						},
					},
				},
			},
			provider: "digitalocean",
		},
		{
			name: "GCPCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "GCPCluster"
							}`),
						},
					},
				},
			},
			provider: "gcp",
		},
		{
			name: "OpenStackCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "OpenStackCluster"
							}`),
						},
					},
				},
			},
			provider: "openstack",
		},
		{
			name: "PacketCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "PacketCluster"
							}`),
						},
					},
				},
			},
			provider: "packet",
		},
		{
			name: "VSphereCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "VSphereCluster"
							}`),
						},
					},
				},
			},
			provider: "vsphere",
		},
		{
			name: "FooCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "FooCluster"
							}`),
						},
					},
				},
			},
			provider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider != getProvider(tt.template) {
				t.Fatalf("expected %s but got %s", tt.provider, getProvider(tt.template))
			}
		})
	}
}

func TestGenerateProfileFiles(t *testing.T) {
	c := createClient(makeTestHelmRepository("base"))
	file, err := generateProfileFiles(
		context.TODO(),
		"testing",
		"test-ns",
		"cluster-foo",
		c,
		[]*capiv1_protos.ProfileValues{
			{
				Name:    "foo",
				Version: "0.0.1",
				Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
			},
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: cluster-foo-foo
  namespace: wego-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta1
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  interval: 1m0s
  values:
    foo: bar
status: {}
`

	assert.Equal(t, *file.Content, expected)
}

func createClient(clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1beta1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
}

func createServer(clusterState []runtime.Object, configMapName, namespace string, provider git.Provider, db *gorm.DB, ns string) capiv1_protos.ClustersServiceServer {

	c := createClient(clusterState...)

	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	s := NewClusterServer(logr.Discard(),
		&templates.ConfigMapLibrary{
			Log:           logr.Discard(),
			Client:        c,
			ConfigMapName: configMapName,
			Namespace:     namespace,
		}, provider, c, dc, db, ns, "weaveworks-charts", "")

	return s
}

func createDatabase(t *testing.T) *gorm.DB {
	db, err := utils.OpenDebug("", true)
	if err != nil {
		t.Fatal(err)
	}
	err = utils.MigrateTables(db)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func makeTestHelmRepository(base string, opts ...func(*sourcev1beta1.HelmRepository)) *sourcev1beta1.HelmRepository {
	hr := &sourcev1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta1.HelmRepositoryKind,
			APIVersion: sourcev1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "test-ns",
		},
		Spec: sourcev1beta1.HelmRepositorySpec{
			URL:      base + "/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1beta1.HelmRepositoryStatus{
			URL: base + "/index.yaml",
		},
	}
	for _, o := range opts {
		o(hr)
	}
	return hr
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

func makeTemplate(t *testing.T, opts ...func(*capiv1.CAPITemplate)) string {
	t.Helper()
	basicRaw := `
	{
		"apiVersion":"fooversion",
		"kind":"fookind",
		"metadata":{
		   "name":"${CLUSTER_NAME}",
		   "annotations":{
			  "capi.weave.works/display-name":"ClusterName"
		   }
		}
	 }`
	ct := &capiv1.CAPITemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CAPITemplate",
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template-1",
		},
		Spec: capiv1.CAPITemplateSpec{
			Description: "this is test template 1",
			Params: []capiv1.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []capiv1.CAPIResourceTemplate{
				{
					RawExtension: rawExtension(basicRaw),
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

func makeNamespace(n string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
	}
}

func makeSecret(n string, ns string, s ...string) *corev1.Secret {
	data := make(map[string][]byte)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = []byte(s[i+1])
	}

	nsObj := makeNamespace(ns)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: nsObj.GetName(),
		},
		Data: data,
	}
}

func rawExtension(s string) runtime.RawExtension {
	return runtime.RawExtension{
		Raw: []byte(s),
	}
}

func NewFakeGitProvider(url string, repo *git.GitRepo, err error) git.Provider {
	return &FakeGitProvider{
		url:  url,
		repo: repo,
		err:  err,
	}
}

type FakeGitProvider struct {
	url  string
	repo *git.GitRepo
	err  error
}

func (p *FakeGitProvider) WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req git.WriteFilesToBranchAndCreatePullRequestRequest) (*git.WriteFilesToBranchAndCreatePullRequestResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &git.WriteFilesToBranchAndCreatePullRequestResponse{WebURL: p.url}, nil
}

func (p *FakeGitProvider) CloneRepoToTempDir(req git.CloneRepoToTempDirRequest) (*git.CloneRepoToTempDirResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &git.CloneRepoToTempDirResponse{Repo: p.repo}, nil
}
