package git

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_getGitProviderClient(t *testing.T) {

	azureGitProvider := GitProvider{
		Token:    os.Getenv("AZURE_DEVOPS_TOKEN"),
		Type:     "azure",
		Hostname: "dev.azure.com",
	}

	client, err := getGitProviderClient(azureGitProvider)

	require.NoError(t, err)
	require.NotNil(t, client.ProviderID())
}
