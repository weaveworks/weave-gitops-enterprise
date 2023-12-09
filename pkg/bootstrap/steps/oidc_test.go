package steps

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const hrFileContentTest = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: weave-gitops-enterprise
  namespace: flux-system
spec:
  chart:
    spec:
      chart: mccp
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: weave-gitops-enterprise-charts
        namespace: flux-system
  install:
    crds: CreateReplace
  interval: 1h0m0s
  upgrade:
    crds: CreateReplace
  values:
    cluster-controller:
      controllerManager:
        manager:
          image: {}
    config:
      oidc:
        clientCredentialsSecret: oidc-auth
        enabled: true
        issuerURL: https://dex-01.wge.dev.weave.works
        redirectURL: http://localhost:8000/oauth2/callback
    global: {}
status: {}
`

func TestCreateOIDCConfig(t *testing.T) {
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

	hr1 := helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise",
			Namespace: "flux-system",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   "mccp",
					Version: ">= 0.0.0-0",
					SourceRef: helmv2.CrossNamespaceObjectReference{
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
		name        string
		input       []StepInput
		output      []StepOutput
		err         bool
		helmrelease helmv2.HelmRelease
	}{
		{
			name: "wrong discovery url",
			input: []StepInput{
				{
					Name: inDiscoveryURL, Value: "https://wrong-url.com",
				},
				{
					Name: inClientID, Value: "client-id",
				},
				{
					Name: inClientSecret, Value: "client-secret",
				},
			},
			helmrelease: hr1,
			err:         true,
		},
		{
			name: "Case with all fields",
			input: []StepInput{
				{
					Name: inDiscoveryURL, Value: "https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				},
				{
					Name: inClientID, Value: "client-id",
				},
				{
					Name: inClientSecret, Value: "client-secret",
				},
			},
			err:         false,
			helmrelease: hr1,
			output: []StepOutput{
				{
					Name: oidcSecretName,
					Type: typeSecret,
					Value: corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      oidcSecretName,
							Namespace: WGEDefaultNamespace,
						},
						Data: map[string][]byte{
							"issuerURL":    []byte("https://dex-01.wge.dev.weave.works"),
							"clientID":     []byte("client-id"),
							"clientSecret": []byte("client-secret"),
							"redirectURL":  []byte("http://localhost:8000/oauth2/callback"),
						},
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						CommitMsg: oidcCommitMsg,
						Content:   hrFileContentTest,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			schemeBuilder := runtime.SchemeBuilder{
				corev1.AddToScheme,
				kustomizev1.AddToScheme,
				sourcev1.AddToScheme,
				helmv2.AddToScheme,
			}
			err := schemeBuilder.AddToScheme(scheme)
			if err != nil {
				t.Fatal(err)
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&tt.helmrelease).Build()
			cliLogger := logger.NewCLILogger(os.Stdout)
			config := Config{
				Logger:           cliLogger,
				KubernetesClient: fakeClient,
			}

			out, err := createOIDCConfig(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}

			// validate secret output
			for i, item := range out {
				if i == 0 {
					// validate secret
					assert.Equal(t, item.Name, tt.output[i].Name)
					assert.Equal(t, item.Type, tt.output[i].Type)
					outSecret, ok := item.Value.(corev1.Secret)
					if !ok {
						t.Fatalf("failed getting result secret data")
					}
					inSecret, ok := tt.output[i].Value.(corev1.Secret)
					if !ok {
						t.Fatalf("failed getting output secret data")
					}
					assert.Equal(t, outSecret.Name, inSecret.Name, "mismatch name")
					assert.Equal(t, outSecret.Namespace, inSecret.Namespace, "mismatch namespace")
					assert.Equal(t, outSecret.Data["issuerURL"], inSecret.Data["issuerURL"], "mismatch url")
					assert.Equal(t, outSecret.Data["clientID"], inSecret.Data["clientID"], "mismatch id")
					assert.Equal(t, outSecret.Data["clientSecret"], inSecret.Data["clientSecret"], "mismatch secret")
					assert.Equal(t, outSecret.Data["redirectURL"], inSecret.Data["redirectURL"], "mismatch url")
				}
				if i == 1 {
					assert.Equal(t, item.Name, tt.output[i].Name)
					assert.Equal(t, item.Type, tt.output[i].Type)
					outFile, ok := item.Value.(fileContent)
					if !ok {
						t.Fatalf("failed getting result secret data")
					}
					inFile, ok := tt.output[i].Value.(fileContent)
					if !ok {
						t.Fatalf("failed getting output secret data")
					}
					assert.Equal(t, outFile.Name, inFile.Name, "mismatch name")
					assert.Equal(t, outFile.CommitMsg, inFile.CommitMsg, "mismatch commit msg")
					assert.Equal(t, outFile.Content, inFile.Content, "mismatch content")
				}
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
