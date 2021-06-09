package templates_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
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
			expected: "No templates found.",
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
			r := NewFakeTemplateRetriever(tt.ts, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.ListTemplates(r, w)
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
			expected: "NAME\tDESCRIPTION\ntemplate-param-a\ntemplate-param-b\n",
		},
		{
			name: "templates include all fields",
			tps: []templates.TemplateParameter{
				{
					Name:        "template-param-a",
					Description: "a desc",
				},
				{
					Name:        "template-param-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-param-a\ta desc\ntemplate-param-b\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve template parameters from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFakeTemplateRetriever(nil, tt.tps, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.ListTemplateParameters("foo", r, w)
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
			r := NewFakeTemplateRetriever(nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.RenderTemplate("foo", nil, r, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeTemplateRetriever struct {
	ts  []templates.Template
	tps []templates.TemplateParameter
	s   string
	err error
}

func NewFakeTemplateRetriever(ts []templates.Template, tps []templates.TemplateParameter, s string, err error) FakeTemplateRetriever {
	return FakeTemplateRetriever{
		ts:  ts,
		tps: tps,
		s:   s,
		err: err,
	}
}

func (r FakeTemplateRetriever) Source() string {
	return "In-memory fake"
}

func (r FakeTemplateRetriever) RetrieveTemplates() ([]templates.Template, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.ts, nil
}

func (r FakeTemplateRetriever) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.tps, nil
}

func (r FakeTemplateRetriever) RenderTemplateWithParameters(name string, values map[string]string) (string, error) {
	if r.err != nil {
		return "", r.err
	}

	return r.s, nil
}
