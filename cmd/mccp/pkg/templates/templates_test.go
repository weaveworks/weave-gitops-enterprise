package templates_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/templates"
)

func TestListTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No templates found.\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name: "template-a",
				},
				{
					Name: "template-b",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ntemplate-b\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Description: "a desc",
				},
				{
					Name:        "template-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ta desc\ntemplate-b\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.ts, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.ListTemplates("", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestListTemplateParameters(t *testing.T) {
	tests := []struct {
		name             string
		tps              []templates.TemplateParameter
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No template parameters found.",
		},
		{
			name: "template parameters include just name",
			tps: []templates.TemplateParameter{
				{
					Name: "template-param-a",
				},
				{
					Name: "template-param-b",
				},
			},
			expected: "NAME\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ntemplate-param-b\n",
		},
		{
			name: "templates include all fields",
			tps: []templates.TemplateParameter{
				{
					Name:        "template-param-a",
					Description: "a desc",
					Options:     []string{"op-1", "op-2"},
				},
				{
					Name:        "template-param-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ta desc\top-1, op-2\ntemplate-param-b\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve template parameters from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(nil, tt.tps, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.ListTemplateParameters("foo", c, w)
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
		result           string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no result returned",
			expected: "No template found.",
		},
		{
			name:             "error returned",
			err:              errors.New("expected param CLUSTER_NAME to be passed"),
			expectedErrorStr: "unable to render template: expected param CLUSTER_NAME to be passed",
		},
		{
			name: "result is rendered to output",
			result: `apiVersion: cluster.x-k8s.io/v1alpha3
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
			expected: `apiVersion: cluster.x-k8s.io/v1alpha3
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
			c := NewFakeClient(nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.RenderTemplate("foo", nil, templates.Credentials{}, c, w)
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
			c := NewFakeClient(nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.CreatePullRequest(templates.CreatePullRequestForTemplateParams{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestListCredentials(t *testing.T) {
	tests := []struct {
		name             string
		creds            []templates.Credentials
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no credentials",
			expected: "No credentials found.",
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
			c := NewFakeClient(nil, nil, tt.creds, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.ListCredentials(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeClient struct {
	ts    []templates.Template
	tps   []templates.TemplateParameter
	creds []templates.Credentials
	s     string
	err   error
}

func NewFakeClient(ts []templates.Template, tps []templates.TemplateParameter, creds []templates.Credentials, s string, err error) *FakeClient {
	return &FakeClient{
		ts:    ts,
		tps:   tps,
		creds: creds,
		s:     s,
		err:   err,
	}
}

func (c *FakeClient) Source() string {
	return "In-memory fake"
}

func (c *FakeClient) RetrieveTemplates() ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *FakeClient) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.tps, nil
}

func (c *FakeClient) RenderTemplateWithParameters(name string, values map[string]string, creds templates.Credentials) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *FakeClient) CreatePullRequestForTemplate(params templates.CreatePullRequestForTemplateParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *FakeClient) RetrieveCredentials() ([]templates.Credentials, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.creds, nil
}
