package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetRepoPath tests the GetRepoPath function
func TestGetRepoPath(t *testing.T) {
	fakeClient := utils.CreateFakeClient(t, &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "clusters/production",
		}})

	expectedRepoPath := "clusters/production"

	repoPath, err := getRepoPath(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}

func TestGetGitAuthMethod(t *testing.T) {
	tests := []struct {
		name               string
		authType           string
		privateKeyPath     string
		privateKeyPassword string
		gitUsername        string
		gitToken           string
		authName           string
		err                bool
	}{
		{
			name:        "test with valid https",
			authType:    httpsAuth,
			gitUsername: "testuser",
			gitToken:    "testtoken",
			err:         false,
			authName:    "http-basic-auth",
		},
		{
			name:               "test with valid ssh",
			authType:           sshAuth,
			privateKeyPath:     "/tmp/pk-222",
			privateKeyPassword: "",
			err:                false,
			authName:           "ssh-public-keys",
		},
		{
			name:               "test with unsupported",
			authType:           "unsupported",
			privateKeyPath:     "/tmp/pk-222",
			privateKeyPassword: "",
			gitUsername:        "testuser",
			gitToken:           "testtoken",
			err:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.privateKeyPath != "" {
				file, err := createSSHPrivateKey(tt.privateKeyPath)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				defer file.Close()
			}
			authMethod, err := getGitAuthMethod(tt.authType, tt.privateKeyPath, tt.privateKeyPassword, tt.gitUsername, tt.gitToken)
			if err != nil {
				if tt.err {
					assert.Error(t, err, "expected error")
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, tt.authName, authMethod.Name(), "unexpected auth method")
		})
	}
}

func createSSHPrivateKey(privateKeyPath string) (*os.File, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Save private key to a file
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		defer privateKeyFile.Close()
		return nil, err
	}

	err = pem.Encode(privateKeyFile, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return privateKeyFile, nil
}
