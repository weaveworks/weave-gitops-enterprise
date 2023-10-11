package templates_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestCreatePullRequestFromTemplate_CAPI(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("../adapters/testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("../adapters/testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/templates/pull-request\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/templates/pull-request\": Post \"https://weave.works/api/v1/templates/pull-request\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/templates/pull-request\" was 400")
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
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/v1/templates/pull-request", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)

			result, err := client.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{TemplateKind: templates.CAPITemplateKind.String()})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestCreatePullRequestFromTemplate_Terraform(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("../adapters/testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("../adapters/testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/tfcontrollers\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/tfcontrollers\": Post \"https://weave.works/api/v1/tfcontrollers\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/tfcontrollers\" was 400")
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
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/v1/tfcontrollers", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			result, err := client.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{TemplateKind: templates.GitOpsTemplateKind.String()})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestGetTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No templates were found.\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
				{
					Name: "template-b",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\t\ntemplate-b\t\t\t\t\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Description: "a desc",
					Provider:    "azure",
					Error:       "",
				},
				{
					Name:        "template-b",
					Description: "b desc",
					Error:       "something went wrong",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\tazure\t\ta desc\t\ntemplate-b\t\t\tb desc\tsomething went wrong\n",
		},
		{
			name: "templates with template type",
			ts: []templates.Template{
				{
					Name:         "template-a",
					Description:  "a desc",
					Provider:     "azure",
					TemplateType: "cluster",
					Error:        "",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\tazure\tcluster\ta desc\t\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(templates.CAPITemplateKind, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name             string
		tmplName         string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			tmplName: "",
			expected: "No templates were found.\n",
		},
		{
			name:     "templates includes just name",
			tmplName: "template-a",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\t\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(templates.CAPITemplateKind, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplatesByProvider(t *testing.T) {
	tests := []struct {
		name             string
		provider         string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			provider: "aws",
			expected: "No templates were found for provider \"aws\".\n",
		},
		{
			name:     "templates includes just name",
			provider: "aws",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
				{
					Name:     "template-b",
					Provider: "aws",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\t\ntemplate-b\taws\t\t\t\n",
		},
		{
			name:     "templates include all fields",
			provider: "azure",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Provider:    "azure",
					Description: "a desc",
					Error:       "",
				},
				{
					Name:        "template-b",
					Provider:    "azure",
					Description: "b desc",
					Error:       "something went wrong",
				},
			},
			expected: "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\ntemplate-a\tazure\t\ta desc\t\ntemplate-b\tazure\t\tb desc\tsomething went wrong\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplatesByProvider(templates.CAPITemplateKind, tt.provider, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplateParameters(t *testing.T) {
	tests := []struct {
		name             string
		tps              []templates.TemplateParameter
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No template parameters were found.\n",
		},
		{
			name: "template parameters include just name",
			tps: []templates.TemplateParameter{
				{
					Name:     "template-param-a",
					Required: true,
				},
				{
					Name: "template-param-b",
				},
			},
			expected: "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ttrue\ntemplate-param-b\tfalse\n",
		},
		{
			name: "templates include all fields",
			tps: []templates.TemplateParameter{
				{
					Name:        "template-param-a",
					Required:    true,
					Description: "a desc",
					Options:     []string{"op-1", "op-2"},
				},
				{
					Name:        "template-param-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ttrue\ta desc\top-1, op-2\ntemplate-param-b\tfalse\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve parameters for template \"foo\" from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, tt.tps, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplateParameters(templates.CAPITemplateKind, "foo", "default", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name             string
		result           *templates.RenderTemplateResponse
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no result returned",
			expected: "No template was found.\n",
		},
		{
			name:             "error returned",
			err:              errors.New("expected param CLUSTER_NAME to be passed"),
			expectedErrorStr: "unable to render template \"foo\": expected param CLUSTER_NAME to be passed",
		},
		{
			name: "result is rendered to output",
			result: &templates.RenderTemplateResponse{
				RenderedTemplate: []templates.CommitFile{
					{
						Path: "foo.yaml",
						Content: `apiVersion: cluster.x-k8s.io/v1alpha3
				kind: Cluster
				metadata:
					name: foo
				spec:
					clusterNetwork:
					pods:
						cidrBlocks:
						- 192.168.0.0/16
					controlPlaneRef:
					apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
					kind: KubeadmControlPlane
					name: foo-control-plane
					infrastructureRef:
					apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
					kind: AWSCluster
					name: foo`,
					},
				},
			},
			expected: `
---
# foo.yaml

apiVersion: cluster.x-k8s.io/v1alpha3
				kind: Cluster
				metadata:
					name: foo
				spec:
					clusterNetwork:
					pods:
						cidrBlocks:
						- 192.168.0.0/16
					controlPlaneRef:
					apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
					kind: KubeadmControlPlane
					name: foo-control-plane
					infrastructureRef:
					apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
					kind: AWSCluster
					name: foo`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, nil, tt.result, "", tt.err)
			w := new(bytes.Buffer)
			req := templates.RenderTemplateRequest{
				TemplateName: "foo",
				TemplateKind: templates.CAPITemplateKind,
			}
			err := templates.RenderTemplateWithParameters(req, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestCreatePullRequest(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			err:              errors.New("something went wrong"),
			expectedErrorStr: "unable to create pull request: something went wrong",
		},
		{
			name:     "pull request created",
			result:   "https://github.com/org/repo/pull/1",
			expected: "Created pull request: https://github.com/org/repo/pull/1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name             string
		creds            []templates.Credentials
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no credentials",
			expected: "No credentials were found.\n",
		},
		{
			name: "credentials found",
			creds: []templates.Credentials{
				{
					Name: "creds-a",
					Kind: "AWSCluster",
				},
				{
					Name: "creds-b",
					Kind: "AzureCluster",
				},
			},
			expected: "NAME\tINFRASTRUCTURE PROVIDER\ncreds-a\tAWS\ncreds-b\tAzure\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve credentials from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, tt.creds, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetCredentials(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplateProfiles(t *testing.T) {
	tests := []struct {
		name             string
		fs               []templates.Profile
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no profiles",
			expected: "No template profiles were found.\n",
		},
		{
			name: "profiles includes just name",
			fs: []templates.Profile{
				{
					Name:              "profile-a",
					AvailableVersions: []string{"v0.0.15"},
				},
				{
					Name: "profile-b",
				},
			},
			expected: "NAME\tLATEST_VERSIONS\nprofile-a\tv0.0.15\nprofile-b\t\n",
		},
		{
			name: "profiles include more than 5 versions",
			fs: []templates.Profile{
				{
					Name:              "profile-a",
					AvailableVersions: []string{"v0.0.9", "v0.0.10", "v0.0.11", "v0.0.12", "v0.0.13", "v0.0.14", "v0.0.15"},
				},
				{
					Name:              "profile-b",
					AvailableVersions: []string{"v0.0.13", "v0.0.14", "v0.0.15"},
				},
			},
			expected: "NAME\tLATEST_VERSIONS\nprofile-a\tv0.0.11, v0.0.12, v0.0.13, v0.0.14, v0.0.15\nprofile-b\tv0.0.13, v0.0.14, v0.0.15\n",
		},
		{
			name:             "error retrieving profiles",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve profiles for template \"profile-b\" from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, tt.fs, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplateProfiles("profile-b", "default", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestParseProfileFlags(t *testing.T) {

	// make an example values.yaml in a temp dir
	tmpDir := t.TempDir()
	valuesFile := filepath.Join(tmpDir, "values.yaml")
	err := os.WriteFile(valuesFile, []byte("foo: bar"), 0644)
	assert.NoError(t, err)

	tests := []struct {
		testName    string
		profiles    []string
		expected    []templates.ProfileValues
		expectedErr string
	}{
		{
			testName: "no profiles",
			profiles: []string{"name=profile-a,version=v0.0.15"},
			expected: []templates.ProfileValues{
				{
					Name:    "profile-a",
					Version: "v0.0.15",
				},
			},
		},
		{
			testName: "with values",
			profiles: []string{fmt.Sprintf("name=profile-a,version=v0.0.15,values=%s", valuesFile)},
			expected: []templates.ProfileValues{
				{
					Name:    "profile-a",
					Version: "v0.0.15",
					// base64 encoded version of "foo: bar"
					Values: "Zm9vOiBiYXI=",
				},
			},
		},
		{
			testName:    "missing name",
			profiles:    []string{"version=v0.0.15"},
			expectedErr: "profile name must be specified",
		},
		{
			testName:    "invalid name",
			profiles:    []string{"name='how do you do'"},
			expectedErr: `invalid value for name "'how do you do'"`,
		},
		{
			testName:    "invalid version",
			profiles:    []string{"name=profile-a,version=foo"},
			expectedErr: `invalid semver for version "foo": improper constraint: foo`,
		},
		{
			testName:    "invalid namespace",
			profiles:    []string{"name=profile-a,namespace='how do you do'"},
			expectedErr: `invalid value for namespace "'how do you do'"`,
		},
		{
			testName: "multiple profiles",
			profiles: []string{"name=profile-a,version=v0.0.15", "name=profile-b,version=v0.0.16"},
			expected: []templates.ProfileValues{
				{Name: "profile-a", Version: "v0.0.15"},
				{Name: "profile-b", Version: "v0.0.16"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			actual, err := templates.ParseProfileFlags(tt.profiles)
			if tt.expectedErr != "" {
				assert.Regexp(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

type fakeClient struct {
	ts  []templates.Template
	ps  []templates.TemplateParameter
	cs  []templates.Credentials
	fs  []templates.Profile
	rt  *templates.RenderTemplateResponse
	w   string
	err error
}

func newFakeClient(ts []templates.Template, ps []templates.TemplateParameter, cs []templates.Credentials, fs []templates.Profile, rt *templates.RenderTemplateResponse, w string, err error) *fakeClient {
	return &fakeClient{
		ts:  ts,
		ps:  ps,
		cs:  cs,
		fs:  fs,
		rt:  rt,
		w:   w,
		err: err,
	}
}

func (c *fakeClient) Source() string {
	return "In-memory fake"
}

func (c *fakeClient) RetrieveTemplates(kind templates.TemplateKind) ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplate(name string, kind templates.TemplateKind, namespace string) (*templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.ts[0].Name == name {
		return &c.ts[0], nil
	}

	return nil, errors.New("not found")
}

func (c *fakeClient) RetrieveTemplatesByProvider(kind templates.TemplateKind, provider string) ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplateParameters(kind templates.TemplateKind, name string, namespace string) ([]templates.TemplateParameter, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ps, nil
}

func (c *fakeClient) RenderTemplateWithParameters(req templates.RenderTemplateRequest) (*templates.RenderTemplateResponse, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.rt, nil
}

func (c *fakeClient) CreatePullRequestFromTemplate(params templates.CreatePullRequestFromTemplateParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.w, nil
}

func (c *fakeClient) RetrieveCredentials() ([]templates.Credentials, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.cs, nil
}

func (c *fakeClient) RetrieveTemplateProfiles(name string, namespace string) ([]templates.Profile, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.fs, nil
}
