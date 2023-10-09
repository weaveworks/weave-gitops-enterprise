package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type authConfigParams struct {
	Type         string
	UserDomain   string
	WGEVersion   string
	DiscoveryURL string
	ClientID     string
	ClientSecret string
}

// TestCreateOIDCConfig tests the CreateOIDCConfig function.
func TestCreateOIDCConfig(t *testing.T) {
	test := []struct {
		name   string
		input  authConfigParams
		expect OIDCConfig
		err    error
	}{
		{
			name: "AuthConfigParams with all fields",
			input: authConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
		},
		{
			name: "AuthConfigParams with invalid DiscoveryURL",
			input: authConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration-invalid",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			logger.NewCLILogger(os.Stdout)
			// config := Config{
			// 	Logger: logger,
			// }

			// err := createOIDCConfig()
			// if err != nil {
			// 	assert.NotNil(t, err)
			// 	return
			// }

			// make sure that a secret with the name oidc exists
			// secret, err := utils.GetSecret(config.KubernetesClient, "oidc", "flux-system")
			// if err != nil {
			// 	assert.NotNil(t, err)
			// 	return
			// }
			//asert secret data contains clientID and clientSecret
			//assert.Equal(t, tt.expect.ClientID, string(secret.Data["clientID"]), "Expected clientID %s, but got %s", tt.expect.ClientID, string(secret.Data["clientID"]))
		})
	}
}
func TestGetIssuer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, `{"issuer": "https://example.com/issuer"}`)
	}))
	defer mockServer.Close()

	issuer, err := getIssuer(mockServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedIssuer := "https://example.com/issuer"
	assert.Equal(t, expectedIssuer, issuer, "Expected issuer %s, got %s", expectedIssuer, issuer)
}

func TestGetOIDCSecrets(t *testing.T) {

	tests := []struct {
		name   string
		input  authConfigParams
		expect OIDCConfig
		err    error
	}{
		{
			name: "AuthConfigParams with all fields",
			input: authConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
			expect: OIDCConfig{
				IssuerURL:    "https://dex-01.wge.dev.weave.works",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "http://localhost:8000/oauth2/callback",
			},
			err: nil,
		},
		{
			name: "AuthConfigParams with invalid DiscoveryURL",
			input: authConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration-invalid",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
			expect: OIDCConfig{
				IssuerURL:    "https://dex-01.wge.dev.weave.works",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "http://localhost:8000/oauth2/callback",
			},
			err: fmt.Errorf("error: OIDC discovery URL returned status 404"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.NewCLILogger(os.Stdout)
			// config := Config{
			// 	Logger: logger,
			// }

			// result, err := config.getOIDCSecrets(tt.input)
			// if err != nil {
			// 	assert.NotNil(t, err)
			// 	return
			// }

			//assert.Equal(t, tt.expect, result)
		})
	}

}
