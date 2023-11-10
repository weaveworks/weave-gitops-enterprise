package utils

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetSecret test TestGetSecret
func TestGetSecret(t *testing.T) {
	secretName := "test-secret"
	secretNamespace := "flux-system"
	invalidSecretName := "invalid-secret"

	fakeClient := utils.CreateFakeClient(t, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Type:       "Opaque",
		Data: map[string][]byte{
			"username": []byte("test-username"),
			"password": []byte("test-password"),
		},
	})

	secret, err := GetSecret(fakeClient, invalidSecretName, secretNamespace)
	assert.Error(t, err, "error fetching secret: %v", err)
	assert.Nil(t, secret, "error fetching secret: %v", err)

	secret, err = GetSecret(fakeClient, secretName, secretNamespace)

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

	fakeClient := utils.CreateFakeClient(t)

	err := CreateSecret(fakeClient, secretName, secretNamespace, secretData)
	assert.NoError(t, err, "error creating secret: %v", err)

	secret, err := GetSecret(fakeClient, secretName, secretNamespace)
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

	fakeClient := utils.CreateFakeClient(t)
	err := CreateSecret(fakeClient, secretName, secretNamespace, secretData)
	assert.NoError(t, err, "error creating secret: %v", err)

	err = DeleteSecret(fakeClient, secretName, secretNamespace)
	assert.NoError(t, err, "error deleting secret: %v", err)

	_, err = GetSecret(fakeClient, secretName, secretNamespace)
	assert.Error(t, err, "an error was expected")

}
