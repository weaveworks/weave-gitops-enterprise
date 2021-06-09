package adapters_test

import (
	"errors"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

const BaseURI = "https://weave.works/api"

func TestRetrieveTemplates(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.Template, err error)
	}{
		{
			name:      "templates returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("testdata/templates.json")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.ElementsMatch(t, ts, []templates.Template{
					{
						Name:        "cluster-template",
						Description: "this is test template 1",
					},
					{
						Name:        "cluster-template-2",
						Description: "this is test template 2",
					},
					{
						Name:        "cluster-template-3",
						Description: "this is test template 3",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET templates from \"https://weave.works/api/v1/templates\": Get https://weave.works/api/v1/templates: oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/templates", tt.responder)

			r, err := adapters.NewHttpTemplateRetriever(BaseURI, client)
			assert.NoError(t, err)
			ts, err := r.RetrieveTemplates()
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRetrieveTemplateParameters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.TemplateParameter, err error)
	}{
		{
			name:      "template parameters returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("testdata/template_parameters.json")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.ElementsMatch(t, ts, []templates.TemplateParameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "unable to GET template parameters from \"https://weave.works/api/v1/templates/cluster-template/params\": Get https://weave.works/api/v1/templates/cluster-template/params: oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates/cluster-template/params\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/templates/cluster-template/params", tt.responder)

			r, err := adapters.NewHttpTemplateRetriever(BaseURI, client)
			assert.NoError(t, err)
			ts, err := r.RetrieveTemplateParameters("cluster-template")
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRenderTemplateWithParameters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "rendered template returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("testdata/rendered_template.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, `apiVersion: cluster.x-k8s.io/v1alpha3
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
    name: foo
`)
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render\": Post https://weave.works/api/v1/templates/cluster-template/render: oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/templates/cluster-template/render\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", BaseURI+"/v1/templates/cluster-template/render", tt.responder)

			r, err := adapters.NewHttpTemplateRetriever(BaseURI, client)
			assert.NoError(t, err)
			result, err := r.RenderTemplateWithParameters("cluster-template", nil)
			tt.assertFunc(t, result, err)
		})
	}
}
