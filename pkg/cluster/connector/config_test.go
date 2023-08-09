package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
)

func TestConfigForContext_missing_context(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/nonexisting-kube-config.yaml"

	_, err := ConfigForContext(opts, "context2")
	assert.Error(t, err, "failed to get context context2")

}

func TestConfigForContext(t *testing.T) {
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := ConfigForContext(opts, "context2")
	if err != nil {
		t.Fatal(err)
	}

	if restCfg.Host != "https://cluster2.example.com" {
		t.Fatalf("expected Host = %s, got %s", "https://cluster2.example.com", restCfg.Host)
	}
}
