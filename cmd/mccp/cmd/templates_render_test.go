package cmd_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/cmd"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
)

func TestSetSeparateValues(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../test/testdata/rendered_template.json"))
		},
	)

	cmd := cmd.RootCmd(client)
	cmd.SetArgs([]string{
		"templates", "render", "cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev",
		"--set=AWS_REGION=us-east-1",
		"--set=AWS_SSH_KEY_NAME=ssh_key",
		"--set=KUBERNETES_VERSION=1.19",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSetMultipleValues(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../test/testdata/rendered_template.json"))
		},
	)

	cmd := cmd.RootCmd(client)
	cmd.SetArgs([]string{
		"templates", "render", "cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev,AWS_REGION=us-east-1,AWS_SSH_KEY_NAME=ssh_key,KUBERNETES_VERSION=1.19",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSetMultipleAndSeparateValues(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../test/testdata/rendered_template.json"))
		},
	)

	cmd := cmd.RootCmd(client)
	cmd.SetArgs([]string{
		"templates", "render", "cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev,AWS_REGION=us-east-1",
		"--set=AWS_SSH_KEY_NAME=ssh_key",
		"--set=KUBERNETES_VERSION=1.19",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}
