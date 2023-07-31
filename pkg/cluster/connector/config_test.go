package connector

import (
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestConfigForContext_missing_context(t *testing.T) {
	t.Skip()
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
