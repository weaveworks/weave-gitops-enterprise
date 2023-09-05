package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConfigForContext_missing_context(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/nonexisting-kube-config.yaml"

	_, err := configForContext(context.TODO(), opts, "hub")
	assert.Error(t, err, "failed to get context hub")

}

func TestConfigForContext(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := configForContext(context.TODO(), opts, "hub")
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

// TestKubeConfigWithToken tests the kubeConfigWithToken function.
// spoke is considered the remote cluster to connect to
func TestKubeConfigWithToken(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := configForContext(context.TODO(), opts, "spoke")
	assert.NoError(t, err)
	config, err := kubeConfigWithToken(context.TODO(), restCfg, "spoke", []byte("testing-token"))
	assert.NoError(t, err)

	want := clientcmdapi.Config{
		Kind:       "",
		APIVersion: "",
		Clusters: map[string]*clientcmdapi.Cluster{
			"spoke": {
				Server:                   "https://spoke.example.com",
				CertificateAuthorityData: []byte("CADATA2"),
				InsecureSkipTLSVerify:    true,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"spoke-cluster-user": {
				Token: "testing-token",
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"spoke": {
				Cluster:  "spoke",
				AuthInfo: "spoke-cluster-user",
			},
		},
		CurrentContext: "spoke",
		Preferences: clientcmdapi.Preferences{
			Extensions: map[string]runtime.Object{},
		},
		Extensions: map[string]runtime.Object{},
	}

	assert.Equal(t, want, *config)
}
