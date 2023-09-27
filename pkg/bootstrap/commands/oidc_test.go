package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/domain"
)

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
		input  domain.OIDCConfigParams
		expect domain.OIDCConfig
		err    error
	}{
		{
			name: "OIDCConfigParams with all fields",
			input: domain.OIDCConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
			expect: domain.OIDCConfig{
				IssuerURL:    "https://dex-01.wge.dev.weave.works",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "http://localhost:8000/oauth2/callback",
			},
			err: nil,
		},
		{
			name: "OIDCConfigParams with invalid DiscoveryURL",
			input: domain.OIDCConfigParams{
				DiscoveryURL: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration-invalid",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				UserDomain:   "localhost",
			},
			expect: domain.OIDCConfig{
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
			result, err := getOIDCSecrets(tt.input)
			if err != nil {
				assert.NotNil(t, err)
				return
			}

			assert.Equal(t, tt.expect, result)
		})
	}

}
