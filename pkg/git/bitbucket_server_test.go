//go:build integration
// +build integration

package git_test

import (
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"testing"
)

func TestCreateGitClientBitbucketServer(t *testing.T) {

	hostname := "https://git.gartner.com/scm/flux/fleet-tenants-irsa.git"
	token := "myToken"
	tokenType := "myTokenType"

	providerOpts := []git.ProviderWithFn{}

	providerOpts = append(providerOpts, git.WithUsername("git"))
	providerOpts = append(providerOpts, git.WithToken(tokenType, token))
	providerOpts = append(providerOpts, git.WithDomain(hostname))
	providerOpts = append(providerOpts, git.WithConditionalRequests())

	// Create a PR using our wrapper
	_, err := git.NewFactory(logr.Discard()).Create(git.BitBucketServerProviderName, providerOpts...)
	require.NoError(t, err)

}
