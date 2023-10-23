//go:build integration

package utils

import (
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// TestGetKubernetesClient test TestGetKubernetesClient
func TestGetKubernetesClientIt(t *testing.T) {

	tests := []struct {
		name        string
		setup       func() string
		reset       func()
		shouldError bool
	}{
		{
			name: "should create kubernetes http without kubeconfig",
			setup: func() string {
				kp := createKubeconfigFileForRestConfig(*cfg)
				os.Setenv("KUBECONFIG", kp)
				return ""
			},
			reset: func() {
				os.Unsetenv("KUBECONFIG")
			},
			shouldError: false,
		},
		{
			name: "should create kubernetes http kubeconfig",
			setup: func() string {
				return createKubeconfigFileForRestConfig(*cfg)
			},
			reset:       func() {},
			shouldError: false,
		},
		{
			name: "should not create kubernetes http with invalid kubeconfig ",
			setup: func() string {
				return "idontexist.yaml"
			},
			reset:       func() {},
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := tt.setup()
			defer tt.reset()

			kubehttp, err := GetKubernetesHttp(kubeconfigPath)
			if tt.shouldError {
				assert.Error(t, err, "error getting Kubernetes client")
				return
			}
			assert.NoError(t, err, "should have Kubernetes client")
			assert.NotNil(t, kubehttp, "should have Kubernetes client")
		})
	}

}

// createKubeconfigFileForRestConfig creates a kubeconfig file so we could use it for calling any command with --kubeconfig pointing to it
func createKubeconfigFileForRestConfig(restConfig rest.Config) string {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	kubeConfigFile, _ := os.CreateTemp("", "kubeconfig")
	_ = clientcmd.WriteToFile(clientConfig, kubeConfigFile.Name())
	return kubeConfigFile.Name()
}
