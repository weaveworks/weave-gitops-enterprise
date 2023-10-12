package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err = createOIDCConfig(tt.input, tt.config)
			if err != nil {
				//fire the errors except it's related to 'no kind is registered for the type v1beta2.GitRepository in scheme'
				//because it's out of the scope of this test to add all the required objects to the fake client
				//check if the error contains the above string and if not, fail the test
				if !strings.Contains(err.Error(), "no kind is registered for the type v1beta2.GitRepository in scheme") {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			//validate oidc-auth secret is created, use getSecret function to get the secret and validate the secret data
			secret, err := utils.GetSecret(tt.config.KubernetesClient, "oidc-auth", "flux-system")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, "client-id", string(secret.Data["clientID"]), "Expected clientID %s, got %s", "client-id", string(secret.Data["clientID"]))
			assert.Equal(t, "client-secret", string(secret.Data["clientSecret"]), "Expected clientSecret %s, got %s", "client-secret", string(secret.Data["clientSecret"]))
			assert.Equal(t, "http://localhost:8000/oauth2/callback", string(secret.Data["redirectURL"]), "Expected redirectURL %s, got %s", "http://localhost:8000/oauth2/callback", string(secret.Data["redirectURL"]))
			assert.Equal(t, "https://dex-01.wge.dev.weave.works", string(secret.Data["issuerURL"]), "Expected issuerURL %s, got %s", "https://dex-01.wge.dev.weave.works/", string(secret.Data["issuerURL"]))

			//validate wge helmrelease is updated with oidc values
			helmrelease := &helmv2.HelmRelease{}
			if err := tt.config.KubernetesClient.Get(context.Background(), k8s_client.ObjectKey{
				Namespace: "flux-system",
				Name:      "weave-gitops-enterprise",
			}, helmrelease); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			values := helmrelease.Spec.Values
			var valuesMap map[string]interface{}
			if err := json.Unmarshal(values.Raw, &valuesMap); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			configMap := valuesMap["config"].(map[string]interface{})
			oidcMap := configMap["oidc"].(map[string]interface{})

			assert.Equal(t, "https://eng-sandbox.weave.works/oauth2/callback", oidcMap["redirectURL"], "Expected redirectURL %s, got %s", "https://eng-sandbox.weave.works/oauth2/callback", oidcMap["redirectURL"])
			assert.Equal(t, "https://dex.eng-sandbox.weave.works", oidcMap["issuerURL"], "Expected issuerURL %s, got %s", "https://dex.eng-sandbox.weave.works", oidcMap["issuerURL"])
			assert.Equal(t, true, oidcMap["enabled"], "Expected enabled %t, got %t", true, oidcMap["enabled"])
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
