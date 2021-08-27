package cmd_test

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/cmd"
)

func TestGetCluster(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		status    int
		response  interface{}
		args      []string
		result    string
		errString string
	}{
		{
			name:     "cluster kubeconfig",
			url:      "http://localhost:8000/v1/clusters/dev-cluster/kubeconfig",
			status:   http.StatusOK,
			response: httpmock.File("../test/testdata/cluster_kubeconfig.json"),
			args: []string{
				"clusters", "get", "dev-cluster",
				"--kubeconfig",
				"--endpoint", "http://localhost:8000",
			},
		},
		{
			name: "http error",
			args: []string{
				"clusters", "get", "dev-cluster",
				"--kubeconfig",
				"--endpoint", "not_a_valid_url",
			},
			errString: "parse \"not_a_valid_url\": invalid URI for request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder(
				http.MethodGet,
				tt.url,
				func(r *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(tt.status, tt.response)
				},
			)

			cmd := cmd.RootCmd(client)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.errString == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.errString)
			}
		})
	}
}
