package root_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
)

func TestInsecureSkipVerifyFalse(t *testing.T) {
	client := adapters.NewHTTPClient()

	cmd := root.Command(client)
	cmd.SetArgs([]string{
		"add", "cluster",
	})

	// Command is incomplete and should raise an error, it helps us short circuit here to quickly
	// test that the client has been set
	err := cmd.Execute()
	assert.Error(t, err)

	transport, ok := client.GetBaseClient().Transport.(*http.Transport)
	assert.True(t, ok)
	// Its set to nil and uses whatever the golang defaults are (InsecureSkipVerify: false)
	assert.Nil(t, transport.TLSClientConfig)
}
