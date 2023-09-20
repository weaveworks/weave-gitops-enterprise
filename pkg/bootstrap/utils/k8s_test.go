package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/loft-sh/vcluster/pkg/util/random"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetSecret test TestGetSecret
func TestGetSecret(t *testing.T) {
	secretName := "test-secret"
	secretNamespace := "flux-system"
	invalidSecretName := "invalid-secret"

	clientset := fake.NewSimpleClientset(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Type:       "Opaque",
		Data: map[string][]byte{
			"username": []byte("test-username"),
			"password": []byte("test-password"),
		},
	})

	secret, err := GetSecret(invalidSecretName, secretNamespace, clientset)
	assert.Error(t, err, "error fetching secret: %v", err)
	assert.Nil(t, secret, "error fetching secret: %v", err)

	secret, err = GetSecret(secretName, secretNamespace, clientset)

	expectedUsername := "test-username"
	expectedPassword := "test-password"
	assert.NoError(t, err, "error fetching secret: %v", err)
	assert.Equal(t, expectedUsername, string(secret.Data["username"]), "Expected username %s, but got %s", expectedUsername, string(secret.Data["username"]))
	assert.Equal(t, expectedPassword, string(secret.Data["password"]), "Expected password %s, but got %s", expectedPassword, string(secret.Data["password"]))

}

// TestCreateSecret test TestCreateSecret
func TestCreateSecret(t *testing.T) {
	secretName := "test-secret"
	secretNamespace := "flux-system"
	secretData := map[string][]byte{
		"username": []byte("test-username"),
		"password": []byte("test-password"),
	}

	clientset := fake.NewSimpleClientset()
	err := CreateSecret(secretName, secretNamespace, secretData, clientset)
	assert.NoError(t, err, "error creating secret: %v", err)

	secret, err := GetSecret(secretName, secretNamespace, clientset)
	expectedUsername := "test-username"
	expectedPassword := "test-password"

	assert.NoError(t, err, "error fetching secret: %v", err)
	assert.Equal(t, expectedUsername, string(secret.Data["username"]), "Expected username %s, but got %s", expectedUsername, string(secret.Data["username"]))
	assert.Equal(t, expectedPassword, string(secret.Data["password"]), "Expected password %s, but got %s", expectedPassword, string(secret.Data["password"]))

}

// TestDeleteSecret test TestDeleteSecret
func TestDeleteSecret(t *testing.T) {
	secretName := "test-secret"
	secretNamespace := "flux-system"
	secretData := map[string][]byte{
		"username": []byte("test-username"),
		"password": []byte("test-password"),
	}

	clientset := fake.NewSimpleClientset()
	err := CreateSecret(secretName, secretNamespace, secretData, clientset)
	assert.NoError(t, err, "error creating secret: %v", err)

	err = DeleteSecret(secretName, secretNamespace, clientset)
	assert.NoError(t, err, "error deleting secret: %v", err)

	_, err = GetSecret(secretName, secretNamespace, clientset)
	assert.Error(t, err, "an error was expected")

}

// TestGetKubernetesClient test TestGetKubernetesClient
func TestGetKubernetesClient(t *testing.T) {
	kubeConfigFileContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: http://example.com
  name: my-cluster
contexts:
- context:
    cluster: my-cluster
    user: my-user
  name: my-context
current-context: my-context
users:
- name: my-user
  user:
    username: test
    password: test
`
	fakeKubeconfigfile := filepath.Join(os.TempDir(), fmt.Sprintf("test-kubeconfig-%s.yaml", random.RandomString(6)))
	file, err := os.Create(fakeKubeconfigfile)
	assert.NoError(t, err, "error creating file")

	defer file.Close()
	defer os.Remove(fakeKubeconfigfile)

	_, err = file.WriteString(kubeConfigFileContent)
	assert.NoError(t, err, "error creating to file")

	client, err := GetKubernetesClient(fakeKubeconfigfile)
	assert.NoError(t, err, "error getting Kubernetes client")
	assert.NotNil(t, client, "expected a non-nil clientset")
}
