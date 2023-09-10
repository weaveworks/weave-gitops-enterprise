package commands

import (
	"testing"

	"github.com/alecthomas/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIsAdminCredsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		secret    *v1.Secret
		available bool
		err       bool
	}{
		{
			name:      "secret doesn't exist",
			secret:    &v1.Secret{},
			available: false,
			err:       false,
		},
		{
			name: "secret exist",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: wgeDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte("test-username"),
					"password": []byte("test-password"),
				},
			},
			available: true,
			err:       false,
		},
		{
			name: "failed to get secret",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "@s", Namespace: wgeDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte("test-username"),
					"password": []byte("test-password"),
				},
			},
			available: false,
			err:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(tt.secret)

			available, err := isAdminCredsAvailable(clientset)
			assert.Equal(t, tt.available, available, "error verifying admin password")
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error verifying admin password, error: %v", err)
			}
		})
	}
}
