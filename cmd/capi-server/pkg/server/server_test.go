package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/git"
	capiv1_protos "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
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
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			expected: []*capiv1_protos.Template{
				{
					Name:        "cluster-template-1",
					Description: "this is test template 1",
					Objects: []*capiv1_protos.TemplateObject{
						{
							ApiVersion: "fooversion",
							Kind:       "fookind",
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
					Objects: []*capiv1_protos.TemplateObject{
						{
							ApiVersion: "fooversion",
							Kind:       "fookind",
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
				{
					Name:        "cluster-template-2",
					Description: "this is test template 2",
					Objects: []*capiv1_protos.TemplateObject{
						{
							ApiVersion: "fooversion",
							Kind:       "fookind",
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
		name             string
		provider         string
		clusterState     []runtime.Object
		expected         []*capiv1_protos.Template
		err              error
		expectedErrorStr string
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
					Objects: []*capiv1_protos.TemplateObject{
						{
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
					Objects: []*capiv1_protos.TemplateObject{
						{
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
				Objects: []*capiv1_protos.TemplateObject{
					{
						ApiVersion: "fooversion",
						Kind:       "fookind",
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
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  labels:\n    name: test-cluster\n",
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
			expected: "apiVersion: fooversion\nkind: fookind\nmetadata:\n  labels:\n    name: test-cluster\n",
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
							"kind": "AWSCluster"
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
			expected: "apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4\nkind: AWSCluster\nspec:\n  identityRef:\n    kind: AWSClusterStaticIdentity\n    name: cred-name\n",
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
		clusterName      string
	}{
		{
			name: "valid value",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "test-cluster",
			expected:    "apiVersion: fooversion\nkind: fookind\nmetadata:\n  labels:\n    name: test-cluster\n",
		},
		{
			name: "first character is not alphabetic",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "2test-cluster",
			err:         errors.New("parameter value 2test-cluster must start with an alphanumeric character"),
		},
		{
			name: "value contains non alphanumeric",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "t&est-cluster",
			err:         errors.New("parameter value t&est-cluster must contain only alphanumeric characters, '-' or '.'"),
		},
		{
			name: "value does not end alphanumeric",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "test-cluster-",
			err:         errors.New("parameter value test-cluster- must end with an alphanumeric character"),
		},
		{
			name: "value contains uppercase letter",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "Test-Cluster",
			err:         errors.New("alphanumueric characters in parameter value Test-Cluster must be lowercase"),
		},
		{
			name: "multiple errors returned",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			clusterName: "2test-cluster.",
			err:         errors.New("2 errors occurred:\nparameter value 2test-cluster. must start with an alphanumeric character\nparameter value 2test-cluster. must end with an alphanumeric character"),
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
			err: errors.New("unable to get secret \"default/dev-kubeconfig\" for Kubeconfig: secrets \"dev-kubeconfig\" not found"),
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
		name         string
		clusterState []runtime.Object
		provider     git.Provider
		req          *capiv1_protos.DeleteClustersPullRequestRequest
		expected     string
		err          error
		dbRows       int
	}{
		{
			name:   "validation errors",
			req:    &capiv1_protos.DeleteClustersPullRequestRequest{},
			err:    errors.New("5 errors occurred:\nat least one cluster name must be specified\nhead branch must be specified\ntitle must be specified\ndescription must be specified\ncommit message must be specified"),
			dbRows: 0,
		},
		{
			name: "create delete pull request",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
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
			dbRows:   0,
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := createDatabase(t)
			s := createServer(tt.clusterState, "capi-templates", "default", tt.provider, db, "")

			req := &capiv1_protos.CreatePullRequestRequest{
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
			}

			// create request
			createPR, _ := s.CreatePullRequest(context.Background(), req)
			fmt.Println("======================")
			fmt.Println(createPR)
			fmt.Println("======================")

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

				var cluster models.Cluster
				db.Where("name = ?", "foo").Find(&cluster)

				if diff := cmp.Diff(cluster.Name, "foo"); diff != "" {
					t.Fatalf("got the wrong name:\n%s", diff)
				}

				var pr models.PRCluster
				db.Where("pr_id = ?", 2).Find(&pr)

				if diff := cmp.Diff(pr.ClusterID, uint(1)); diff != "" {
					t.Fatalf("got the wrong id:\n%s", diff)
				}
			}
		})
	}
}

func createServer(clusterState []runtime.Object, configMapName, namespace string, provider git.Provider, db *gorm.DB, ns string) capiv1_protos.ClustersServiceServer {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	s := NewClusterServer(&templates.ConfigMapLibrary{
		Client:        c,
		ConfigMapName: configMapName,
		Namespace:     namespace,
	}, provider, c, dc, db, ns)

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
  "apiVersion": "fooversion",
  "kind": "fookind",
  "metadata": {
    "labels": {
      "name": "${CLUSTER_NAME}"
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

func makeSecret(n string, ns string, s ...string) *corev1.Secret {
	data := make(map[string][]byte)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = []byte(s[i+1])
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: ns,
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
