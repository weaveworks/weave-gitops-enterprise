package git

import (
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateBitbucketServerClient(t *testing.T) {

	hostname := "https://my-bit-bucket-server/scm/flux/fleet-tenants-irsa.git"
	token := "myToken"
	tokenType := "myTokenType"

	providerOpts := []ProviderWithFn{}

	providerOpts = append(providerOpts, WithUsername("git"))
	providerOpts = append(providerOpts, WithToken(tokenType, token))
	providerOpts = append(providerOpts, WithDomain(hostname))
	providerOpts = append(providerOpts, WithConditionalRequests())

	_, err := NewFactory(testr.New(t)).Create(BitBucketServerProviderName, providerOpts...)
	require.NoError(t, err)
}
