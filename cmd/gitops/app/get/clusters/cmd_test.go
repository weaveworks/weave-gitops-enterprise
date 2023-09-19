package clusters_test

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
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
			url:      "http://localhost:8000/v1/namespaces/default/clusters/dev-cluster/kubeconfig",
			status:   http.StatusOK,
			response: httpmock.File("../../../pkg/adapters/testdata/cluster_kubeconfig.json"),
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--print-kubeconfig",
				"--endpoint", "http://localhost:8000",
			},
		},
		{
			name: "http error",
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--print-kubeconfig",
				"--endpoint", "not_a_valid_url",
			},
			errString: "failed to parse endpoint: parse \"not_a_valid_url\": invalid URI for request",
		},
		{
			name: "no endpoint",
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--print-kubeconfig",
			},
			errString: "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder(
				http.MethodGet,
				tt.url,
				func(r *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(tt.status, tt.response)
				},
			)

			cmd := root.Command(client)
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
