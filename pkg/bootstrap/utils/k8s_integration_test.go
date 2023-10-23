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
	//	kubeConfigFileContent := `apiVersion: v1
	//kind: Config
	//clusters:
	//- cluster:
	//    server: http://example.com
	//  name: my-cluster
	//contexts:
	//- context:
	//    cluster: my-cluster
	//    user: my-user
	//  name: my-context
	//current-context: my-context
	//users:
	//- name: my-user
	//  user:
	//    username: test
	//    password: test
	//`
	tests := []struct {
		name           string
		kubeconfigPath string
		shouldError    bool
	}{
		//{
		//	name:           "should create kubernetes http without kubeconfig ",
		//	kubeconfigPath: "",
		//	shouldError:    false,
		//},
		{
			name:           "should create kubernetes http kubeconfig ",
			kubeconfigPath: createKubeconfigFileForRestConfig(*cfg),
			shouldError:    false,
		},
		{
			name:           "should not create kubernetes http with invalid kubeconfig ",
			shouldError:    true,
			kubeconfigPath: "idontexist.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//fakeKubeconfigfile := filepath.Join(os.TempDir(), fmt.Sprintf("test-kubeconfig-%s.yaml", random.RandomString(6)))
			//file, err := os.Create(fakeKubeconfigfile)
			//assert.NoError(t, err, "error creating file")
			//
			//defer file.Close()
			//defer os.Remove(fakeKubeconfigfile)
			//
			//_, err = file.WriteString(kubeConfigFileContent)
			//assert.NoError(t, err, "error creating to file")

			kubehttp, err := GetKubernetesHttp(tt.kubeconfigPath)
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
