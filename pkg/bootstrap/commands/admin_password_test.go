package commands

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
			err:       true,
		},
		{
			name: "secret exist",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: WGEDefaultNamespace},
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
				ObjectMeta: metav1.ObjectMeta{Name: "@s", Namespace: WGEDefaultNamespace},
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
			scheme := runtime.NewScheme()
			schemeBuilder := runtime.SchemeBuilder{
				v1.AddToScheme,
			}
			err := schemeBuilder.AddToScheme(scheme)
			if err != nil {
				t.Fatal(err)
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.secret).Build()

			available, err := isAdminCredsAvailable(fakeClient)
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

func TestAskAdminCredsSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	err = AskAdminCredsSecret(fakeClient, true)
	assert.NoError(t, err, "an unexpected error occured: %w", err)

	secret, err := utils.GetSecret(adminSecretName, WGEDefaultNamespace, fakeClient)
	assert.NoError(t, err, "an unexpected error occured: %w", err)
	assert.Equal(t, defaultAdminUsername, string(secret.Data["username"]), "error verifying admin username")

	err = bcrypt.CompareHashAndPassword(secret.Data["password"], []byte(defaultAdminPassword))
	assert.NoError(t, err, "an error occured verifying password: %w", err)
}
