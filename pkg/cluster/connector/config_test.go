package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConfigForContext_missing_context(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/nonexisting-kube-config.yaml"

	_, err := ConfigForContext(opts, "hub")
	assert.Error(t, err, "failed to get context hub")

}

func TestConfigForContext(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := ConfigForContext(opts, "hub")
	if err != nil {
		t.Fatal(err)
	}

	if restCfg.Host != "https://hub.example.com" {
		t.Fatalf("expected Host = %s, got %s", "https://hub.example.com", restCfg.Host)
	}
	if string(restCfg.CertData) != "USER2_CADATA" {
		t.Fatalf("expected CertData = %s, got %s", "USER2_CADATA", restCfg.CertData)
	}
	if string(restCfg.KeyData) != "USER1_CKDATA" {
		t.Fatalf("expected KeyData = %s, got %s", "USER1_CKDATA", restCfg.KeyData)
	}

}

// kubeConfigWithToken takes a rest.Config and generates a KubeConfig with the
// named context and configured user credentials from the provided token.
func kubeConfigWithToken(config *rest.Config, context string, token []byte) (*clientcmdapi.Config, error) {
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters[context] = &clientcmdapi.Cluster{
		Server: config.Host,
	}

	return cfg, nil
}

func TestKubeConfigWithToken(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := ConfigForContext(opts, "spoke")
	assert.NoError(t, err)

	config, err := kubeConfigWithToken(restCfg, "spoke", []byte("testing-token"))
	assert.NoError(t, err)

	want := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"spoke": {
				Server:                   "https://spoke.example.com",
				CertificateAuthorityData: []byte("Q0FEQVRBMg=="),
				InsecureSkipTLSVerify:    true,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"user1": {
				Token: "testing-token",
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"context1": {
				Cluster:  "spoke",
				AuthInfo: "user1",
			},
		},
		CurrentContext: "spoke",
	}

	assert.Equal(t, want, config)
}
