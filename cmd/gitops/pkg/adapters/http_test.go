package adapters_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestRetrieveTemplates(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		kind       templates.TemplateKind
		assertFunc func(t *testing.T, templates []templates.Template, err error)
	}{
		{
			name:      "templates returned",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/templates.json")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.ElementsMatch(t, ts, []templates.Template{
					{
						Name:        "cluster-template",
						Provider:    "",
						Description: "this is test template 1",
					},
					{
						Name:        "cluster-template-2",
						Provider:    "aws",
						Description: "this is test template 2",
					},
					{
						Name:        "cluster-template-3",
						Description: "this is test template 3",
						Provider:    "azure",
					},
				})
			},
		},
		{
			name:      "error returned for capi type",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET templates from \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\": Get \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\": oops")
			},
		},
		{
			name:      "error returned for gitops type",
			kind:      templates.GitOpsTemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET templates from \"https://weave.works/api/v1/templates?template_kind=GitOpsTemplate\": Get \"https://weave.works/api/v1/templates?template_kind=GitOpsTemplate\": oops")
			},
		},
		{
			name:      "unexpected status code",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			ts, err := client.RetrieveTemplates(tt.kind)
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRetrieveTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string // this isn't actually used, but it's a nice to have
		responder    httpmock.Responder
		kind         templates.TemplateKind
		assertFunc   func(t *testing.T, template *templates.Template, err error)
	}{
		{
			name:         "capi template returned",
			templateName: "cluster-template",
			kind:         templates.CAPITemplateKind,
			responder:    httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/single_capi_template.json")),
			assertFunc: func(t *testing.T, ts *templates.Template, err error) {
				assert.Equal(t, *ts, templates.Template{
					Name:        "cluster-template",
					Provider:    "aws",
					Description: "this is a test template",
				})
			},
		},
		{
			name:         "terraform template returned",
			kind:         templates.GitOpsTemplateKind,
			templateName: "terraform-template",
			responder:    httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/single_terraform_template.json")),
			assertFunc: func(t *testing.T, ts *templates.Template, err error) {
				assert.Equal(t, *ts, templates.Template{
					Name:        "terraform-template",
					Provider:    "aws",
					Description: "this is test terraform template",
				})
			},
		},
		{
			name:         "error returned for capi type",
			templateName: "cluster-template",
			kind:         templates.CAPITemplateKind,
			responder:    httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts *templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET template from \"https://weave.works/api/v1/templates/cluster-template?template_kind=CAPITemplate\": Get \"https://weave.works/api/v1/templates/cluster-template?template_kind=CAPITemplate\": oops")
			},
		},
		{
			name:         "error returned for gitops type",
			templateName: "terraform-template",
			kind:         templates.GitOpsTemplateKind,
			responder:    httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts *templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET template from \"https://weave.works/api/v1/templates/terraform-template?template_kind=GitOpsTemplate\": Get \"https://weave.works/api/v1/templates/terraform-template?template_kind=GitOpsTemplate\": oops")
			},
		},
		{
			name:         "unexpected status code",
			templateName: "cluster-template",
			kind:         templates.CAPITemplateKind,
			responder:    httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, ts *templates.Template, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates/cluster-template?template_kind=CAPITemplate\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates/"+tt.templateName, tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			ts, err := client.RetrieveTemplate(tt.templateName, tt.kind)
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRetrieveTemplatesByProvider(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.Template, err error)
	}{
		{
			name:      "templates returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/templates_by_provider.json")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.ElementsMatch(t, ts, []templates.Template{
					{
						Name:        "cluster-template-2",
						Provider:    "aws",
						Description: "this is test template 2",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET templates from \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\": Get \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates?template_kind=CAPITemplate\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			ts, err := client.RetrieveTemplates(templates.CAPITemplateKind)
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRetrieveTemplateParameters(t *testing.T) {
	tests := []struct {
		name       string
		kind       templates.TemplateKind
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.TemplateParameter, err error)
	}{
		{
			name:      "template parameters returned for capi kind",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/template_parameters.json")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.ElementsMatch(t, ts, []templates.TemplateParameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
						Options:     []string{"option1", "option2"},
					},
				})
			},
		},
		{
			name:      "template parameters returned for gitops kind",
			kind:      templates.GitOpsTemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/template_parameters.json")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.ElementsMatch(t, ts, []templates.TemplateParameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
						Options:     []string{"option1", "option2"},
					},
				})
			},
		},
		{
			name:      "error returned for capi kind",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "unable to GET template parameters from \"https://weave.works/api/v1/templates/cluster-template/params?template_kind=CAPITemplate\": Get \"https://weave.works/api/v1/templates/cluster-template/params?template_kind=CAPITemplate\": oops")
			},
		},
		{
			name:      "error returned for gitops kind",
			kind:      templates.GitOpsTemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "unable to GET template parameters from \"https://weave.works/api/v1/templates/cluster-template/params?template_kind=GitOpsTemplate\": Get \"https://weave.works/api/v1/templates/cluster-template/params?template_kind=GitOpsTemplate\": oops")
			},
		},
		{
			name:      "unexpected status code",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates/cluster-template/params?template_kind=CAPITemplate\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates/cluster-template/params?template_kind="+tt.kind.String(), tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			ts, err := client.RetrieveTemplateParameters(tt.kind, "cluster-template")
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRenderTemplateWithParameters(t *testing.T) {
	tests := []struct {
		name       string
		kind       templates.TemplateKind
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "rendered template returned for capi kind",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/rendered_template_capi.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, `apiVersion: cluster.x-k8s.io/v1alpha4
kind: Cluster
metadata:
  name: dev
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
    kind: AWSManagedControlPlane
    name: dev-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
    kind: AWSManagedCluster
    name: dev

---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSManagedCluster
metadata:
  name: dev

---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: AWSManagedControlPlane
metadata:
  name: dev-control-plane
spec:
  region: us-east-1
  sshKeyName: ssh_key
  version: "1.19"

---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSFargateProfile
metadata:
  name: dev-fargate-0
spec:
  clusterName: mb-test-1
  selectors:
  - namespace: default
`)
			},
		},
		{
			name:      "rendered template returned for gitops kind",
			kind:      templates.GitOpsTemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/rendered_template_gitops.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, `apiVersion: tfcontroller.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  name: dev
spec:
  interval: 1h
  path: ./
  approvePlan: "auto"
  vars:
    - name: cluster_identifier
      value: cluster-name
  sourceRef:
    kind: GitRepository
    name: git-repo-name
    namespace: git-repo-namespace
`)
			},
		},
		{
			name:      "service error",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("./testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=CAPITemplate\": something bad happened")
			},
		},
		{
			name:      "error returned for capi kind",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=CAPITemplate\": Post \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=CAPITemplate\": oops")
			},
		},
		{
			name:      "error returned for gitops kind",
			kind:      templates.GitOpsTemplateKind,
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=GitOpsTemplate\": Post \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=GitOpsTemplate\": oops")
			},
		},
		{
			name:      "unexpected status code",
			kind:      templates.CAPITemplateKind,
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/templates/cluster-template/render?template_kind=CAPITemplate\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/v1/templates/cluster-template/render?template_kind="+tt.kind.String(), tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			result, err := client.RenderTemplateWithParameters(tt.kind, "cluster-template", nil, templates.Credentials{})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestRetrieveCredentials(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, credentials []templates.Credentials, err error)
	}{
		{
			name:      "credentials returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/credentials.json")),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.ElementsMatch(t, creds, []templates.Credentials{
					{
						Group:     "infrastructure.cluster.x-k8s.io",
						Version:   "v1alpha3",
						Kind:      "AWSClusterStaticIdentity",
						Name:      "aws-creds",
						Namespace: "default",
					},
					{
						Group:     "infrastructure.cluster.x-k8s.io",
						Version:   "v1alpha4",
						Kind:      "AzureClusterIdentity",
						Name:      "azure-creds",
						Namespace: "default",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.EqualError(t, err, "unable to GET credentials from \"https://weave.works/api/v1/credentials\": Get \"https://weave.works/api/v1/credentials\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/credentials\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/credentials", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			creds, err := client.RetrieveCredentials()
			tt.assertFunc(t, creds, err)
		})
	}
}

func TestRetrieveCredentialsByName(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, credentials templates.Credentials, err error)
	}{
		{
			name:      "credentials returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/credentials.json")),
			assertFunc: func(t *testing.T, creds templates.Credentials, err error) {
				assert.Equal(t, creds, templates.Credentials{
					Group:     "infrastructure.cluster.x-k8s.io",
					Version:   "v1alpha3",
					Kind:      "AWSClusterStaticIdentity",
					Name:      "aws-creds",
					Namespace: "default",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/credentials", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			creds, err := client.RetrieveCredentialsByName("aws-creds")
			tt.assertFunc(t, creds, err)
		})
	}
}

func TestRetrieveClusters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, cs []clusters.Cluster, err error)
	}{
		{
			name:      "clusters returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/clusters.json")),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.ElementsMatch(t, cs, []clusters.Cluster{
					{
						Name: "cluster-a",
						Conditions: []clusters.Condition{
							{
								Type:    "Ready",
								Status:  "True",
								Message: "Cluster Found",
							},
						},
					},
					{
						Name: "cluster-b",
						Conditions: []clusters.Condition{
							{
								Type:    "Ready",
								Status:  "True",
								Message: "Cluster Found",
							},
						},
					},
					{
						Name: "cluster-c",
						Conditions: []clusters.Condition{
							{
								Type:    "Ready",
								Status:  "True",
								Message: "Cluster Found",
							},
						},
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.EqualError(t, err, "unable to GET clusters from \"https://weave.works/api/v1/clusters\": Get \"https://weave.works/api/v1/clusters\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/clusters", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			cs, err := client.RetrieveClusters()
			tt.assertFunc(t, cs, err)
		})
	}
}

func TestGetClusterKubeconfig(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, s string, err error)
	}{
		{
			name:      "kubeconfig returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/cluster_kubeconfig.json")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.YAMLEq(t, s, httpmock.File("./testdata/cluster_kubeconfig.yaml").String())
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "unable to GET cluster kubeconfig from \"https://weave.works/api/v1/clusters/dev/kubeconfig\": Get \"https://weave.works/api/v1/clusters/dev/kubeconfig\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/clusters/dev/kubeconfig\" was 400")
			},
		},
		{
			name:      "base64 decode failure",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/cluster_kubeconfig_decode_failure.json")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "unable to base64 decode the cluster kubeconfig: illegal base64 data at input byte 3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/clusters/dev/kubeconfig", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			k, err := client.GetClusterKubeconfig("dev")
			tt.assertFunc(t, k, err)
		})
	}
}

func TestDeleteClusters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("./testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to Delete cluster and create pull request to \"https://weave.works/api/v1/clusters\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to Delete cluster and create pull request to \"https://weave.works/api/v1/clusters\": Delete \"https://weave.works/api/v1/clusters\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for Delete \"https://weave.works/api/v1/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("DELETE", testutils.BaseURI+"/v1/clusters", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			result, err := client.DeleteClusters(clusters.DeleteClustersParams{})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestEntitlementExpiredHeader(t *testing.T) {
	opts := &config.Options{
		Endpoint: testutils.BaseURI,
	}
	client := adapters.NewHTTPClient()
	response := httpmock.NewStringResponse(http.StatusOK, "")
	response.Header.Add("Entitlement-Expired-Message", "This is a test message")

	httpmock.ActivateNonDefault(client.GetBaseClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates", httpmock.ResponderFromResponse(response))

	var buf bytes.Buffer
	err := client.ConfigureClientWithOptions(opts, &buf)
	assert.NoError(t, err)
	_, err = client.RetrieveTemplates(templates.CAPITemplateKind)
	assert.NoError(t, err)
	b, err := io.ReadAll(&buf)
	assert.NoError(t, err)

	if string(b) != "This is a test message\n" {
		t.Errorf("Expected but got %s", string(b))
	}
}

func TestRetrieveTemplateProfiles(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, profile []templates.Profile, err error)
	}{
		{
			name:      "template profiles returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/template_profiles.json")),
			assertFunc: func(t *testing.T, ts []templates.Profile, err error) {
				assert.ElementsMatch(t, ts, []templates.Profile{
					{
						Name:        "profile-a",
						Home:        "https://github.com/org/repo",
						Sources:     []string{"https://github.com/org/repo1", "https://github.com/org/repo2"},
						Description: "this is test profile a",
						Maintainers: []templates.Maintainer{
							{
								Name:  "foo",
								Email: "foo@example.com",
								Url:   "example.com",
							},
						},
						Icon:        "test",
						KubeVersion: "1.19",
						HelmRepository: templates.HelmRepository{
							Name:      "test-repo",
							Namespace: "test-ns",
						},
						AvailableVersions: []string{"v0.0.14", "v0.0.15"},
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, fs []templates.Profile, err error) {
				assert.EqualError(t, err, "unable to GET template profiles from \"https://weave.works/api/v1/templates/cluster-template/profiles\": Get \"https://weave.works/api/v1/templates/cluster-template/profiles\": oops")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", testutils.BaseURI+"/v1/templates/cluster-template/profiles", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			tps, err := client.RetrieveTemplateProfiles("cluster-template")
			tt.assertFunc(t, tps, err)
		})
	}
}

func TestSignin(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, client *resty.Client, err error)
	}{
		{
			name: "sign in successful",
			responder: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"Set-Cookie": {
							"id_token=value",
						},
					},
				}, nil
			},
			assertFunc: func(t *testing.T, client *resty.Client, err error) {
				assert.NotEmpty(t, client.Cookies)
				assert.Equal(t, client.Cookies[0].Name, "id_token")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, client *resty.Client, err error) {
				assert.Equal(t, err.Error(), "error: could not configure auth for client: unable to sign in from \"https://weave.works/api/oauth2/sign_in\": Post \"https://weave.works/api/oauth2/sign_in\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, client *resty.Client, err error) {
				assert.Equal(t, err.Error(), "error: could not configure auth for client: response status for POST \"https://weave.works/api/oauth2/sign_in\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
				Username: "username",
				Password: "pass",
			}
			client := adapters.NewHTTPClient().EnableCLIAuth()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/oauth2/sign_in", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			tt.assertFunc(t, client.GetClient(), err)
		})
	}
}
