package app_test

import (
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
)

func TestNoIssuerURL(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoIssuerURL)
}

func TestNoClientID(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoClientID)
}

func TestNoClientSecret(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoClientSecret)
}

func TestNoRedirectURL(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer os.Remove(tempDir)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
		"--oidc-client-secret=client-secret",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoRedirectURL)
}
