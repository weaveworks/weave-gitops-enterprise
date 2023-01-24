package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var testConfig string = `config:
  devMode: true
  logLevel: foo
  entitlementSecret:
    name: foo
    namespace: foo
  profileCacheLocation: foo
  htmlRootPath: foo
  runtimeNamespace: foo
  helmRepo:
    name: foo
    namespace: foo
  cluster:
    name: foo
  git:
    type: foo
    hostname: foo
  capi:
    templates:
      namespace: foo
      injectPruneAnnotation: foo
      addBasesKustomization: foo
    clusters:
      namespace: foo
    repositoryURL: foo
    repositoryApiURL: foo
    repositoryPath: foo
    repositoryClustersPath: foo
    baseBranch: foo
  checkpoint:
    enabled: true
  oidc:
    enabled: true
    issuerURL: foo
    redirectURL: foo
    cookieDuration: 42h0m0s
    claimUsername: foo
    claimGroups: foo
    clientCredentialsSecret: foo
    customScopes: foo
  auth:
    userAccount:
      enabled: true
    tokenPassthrough:
      enabled: true
  pipelineController:
    address: foo
  ui:
    logoURL: foo
    footer:
      backgroundColor: foo
      color: foo
      content: foo
      hideVersion: true
  costEstimation:
    estimationFilter: foo
    apiRegion: foo
    csvFile: foo
global:
  useK8sCachedClients: true
tls:
  enabled: true
  noTLS: true
  cert: foo
  key: foo
deprecated:
  authMethods:
  - foo
`

func TestUnmarshalConfig(t *testing.T) {
	var values Values

	// yaml unmarshal
	err := kyaml.Unmarshal([]byte(testConfig), &values)
	if err != nil {
		t.Fatalf("Error unmarshaling config: %v", err)
	}

	// marshal to yaml
	bs, err := yaml.Marshal(values)
	if err != nil {
		t.Fatalf("unable to marshal config to YAML: %v", err)
	}

	// compare
	assert.Equal(t, string(bs), testConfig)
}

func TestViperFlags(t *testing.T) {
	var values Values

	// yaml unmarshal
	err := kyaml.Unmarshal([]byte(testConfig), &values)
	if err != nil {
		t.Fatalf("Error unmarshaling config: %v", err)
	}

	items := map[string]interface{}{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ErrorUnset:  true,
		Result:      &items,
	})

	if err != nil {
		t.Fatalf("Error creating decoder: %v", err)
	}

	err = decoder.Decode(values)
	if err != nil {
		t.Fatalf("Error creating decoder: %v", err)
	}

	err = mapstructure.Decode(values, &items)
	if err != nil {
		t.Fatalf("Error creating decoder: %v", err)
	}

	v := viper.New()
	v.MergeConfigMap(items)
	v.Set("log-level", "foo")

	// c := v.AllSettings()
	// bs, err := yaml.Marshal(c)
	// if err != nil {
	// 	t.Fatalf("unable to marshal config to YAML: %v", err)
	// }
}
