package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type authConfigParams struct {
	Type         string
	UserDomain   string
	WGEVersion   string
	DiscoveryURL string
	ClientID     string
	ClientSecret string
}

func TestCreateOIDCConfig(t *testing.T) {

	_ = helmv2beta1.AddToScheme(scheme.Scheme)

	valuesYAML := `config:
      oidc:
        clientCredentialsSecret: oidc-auth
        enabled: true
        issuerURL: https://dex.eng-sandbox.weave.works
        redirectURL: https://eng-sandbox.weave.works/oauth2/callback`

	valuesJSON, err := ConvertYAMLToJSON(valuesYAML)
	if err != nil {
		log.Fatalf("Failed to convert YAML to JSON: %v", err)
	}

	hr := &helmv2beta1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise",
			Namespace: "flux-system",
		},
		Spec: helmv2beta1.HelmReleaseSpec{
			Chart: helmv2beta1.HelmChartTemplate{
				Spec: helmv2beta1.HelmChartTemplateSpec{
					Chart:   "mccp",
					Version: ">= 0.0.0-0",
					SourceRef: helmv2beta1.CrossNamespaceObjectReference{
						Kind:      "HelmRepository",
						Name:      "weave-gitops-enterprise-charts",
						Namespace: "flux-system",
					},
				},
			},
			Values: &v1.JSON{
				Raw: valuesJSON,
			},
		},
	}

	tests := []struct {
		name   string
		input  []StepInput
		config *Config
		expect OIDCConfig
		err    string
	}{
		{
			name: "Case with all fields",
			input: []StepInput{
				{Name: DiscoveryURL, Value: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration"},
				{Name: ClientID, Value: "client-id"},
				{Name: ClientSecret, Value: "client-secret"},
			},
			config: &Config{
				UserDomain:       "localhost",
				KubernetesClient: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(hr).Build(),
				Logger:           &logger.CliLogger{},
			},
			expect: OIDCConfig{IssuerURL: "https://dex-01.wge.dev.weave.works/", ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "http://localhost:8000/oauth2/callback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := createOIDCConfig(tt.input, tt.config)
			if err != nil {
				if tt.err == "" {
					t.Fatalf("expected no error but got: %v", err)
				}
				if err.Error() != tt.err {
					t.Fatalf("expected error '%s' but got: %v", tt.err, err)
				}
				return
			}
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

func ConvertYAMLToJSON(yamlStr string) ([]byte, error) {
	var parsedYaml interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsedYaml); err != nil {
		return nil, err
	}
	converted := convertMapInterface(parsedYaml)
	return json.Marshal(converted)
}
func convertMapInterface(m interface{}) interface{} {
	switch m := m.(type) {
	case map[interface{}]interface{}:
		stringMap := make(map[string]interface{})
		for k, v := range m {
			key, ok := k.(string)
			if !ok {
				continue // Skip this key-value pair if the key isn't a string
			}
			stringMap[key] = convertMapInterface(v)
		}
		return stringMap
	case []interface{}:
		for i, v := range m {
			m[i] = convertMapInterface(v)
		}
	}
	return m
}
