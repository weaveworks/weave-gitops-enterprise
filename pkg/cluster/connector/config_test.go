package connector

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
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
	if string(restCfg.CertData) != "USER2_CADATA" {
		t.Fatalf("expected CertData = %s, got %s", "USER2_CADATA", restCfg.CertData)
	}
	if string(restCfg.KeyData) != "USER1_CKDATA" {
		t.Fatalf("expected KeyData = %s, got %s", "USER1_CKDATA", restCfg.KeyData)
	}

}

func AddClusterToConfig(config *rest.Config, config2 *rest.Config, token []byte) (*rest.Config, error) {
	return nil, nil

}

func TestAddClusterToConfig(t *testing.T) {

	// opts := clientcmd.NewDefaultPathOptions()
	// opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	// restCfg, err := ConfigForContext(opts, "context1")
	// assert.NoError(t, err)
	// t.Logf("cfg : %s", restCfg.String())

	remoteClientSet := fake.NewSimpleClientset()
	serviceAccountName := "test-service-account"
	clusterConnectionOpts := ClusterConnectionOptions{
		ServiceAccountName:     serviceAccountName,
		ClusterRoleName:        serviceAccountName + "-cluster-role",
		ClusterRoleBindingName: serviceAccountName + "-cluster-role-binding",
		Namespace:              corev1.NamespaceDefault,
	}
	token, err := ReconcileServiceAccount(context.Background(), remoteClientSet, clusterConnectionOpts, logr.Logger{})

	opts2 := clientcmd.NewDefaultPathOptions()
	opts2.LoadingRules.ExplicitPath = "new-config.yaml"

	restCfg2, err := ConfigForContext(opts2, "context2")
	if err != nil {
		t.Fatal(err)
	}
	// restCfg2.CertData
	// restCfg2.
	// host := restCfg2.Host

	t.Logf("cfg2 : %s", restCfg2.String())
	t.Logf("cfg2 token : %s", restCfg2.BearerToken)

	newConfig := rest.Config{
		Host:        restCfg2.Host,
		BearerToken: string(token),
	}
	t.Logf("newConfig : %s", newConfig.String())

	// config, err := AddClusterToConfig(restCfg, restCfg2, token)

}
