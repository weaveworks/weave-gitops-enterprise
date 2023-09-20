package commands

import (
	"testing"

	"github.com/alecthomas/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// CheckEntitlementFile test CheckEntitlementFile
func TestCheckEntitlementFile(t *testing.T) {
	var (
		// This entitlement has been generated with the right private key for 1 day
		validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNjMxMzYxMjg2LCJpYXQiOjE2MzEyNzQ4ODYsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxMjc0ODg2LCJzdWIiOiJ0ZXN0QHdlYXZlLndvcmtzIn0.EKGp89DFcRKZ_kGmC8FuLVPB0wiab2KddkQKAmVNC9UH459v63tCP13eFybx9dAmMuaC77SA8rp7ukN1qZM7DA`

		// This entitlement has been generated with a different private key
		invalidEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNjMxMzYxNDkwLCJpYXQiOjE2MzEyNzUwOTAsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxMjc1MDkwLCJzdWIiOiJ0ZXN0QHdlYXZlLndvcmtzIn0.E3Kfg4YzDOYJsTN9lD6B4uoW29tE0IB9X7lOpirSTwcZ7vVHk5PUXznYdiPIi9aSgLGAPIQL3YkAM4lyft3BDg`
	)

	tests := []struct {
		name   string
		secret *v1.Secret
		valid  bool
	}{
		{
			name:   "secret does not exist",
			secret: &v1.Secret{},
			valid:  false,
		},
		{
			name: "invalid entitlement",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: entitlementSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"entitlement": []byte(invalidEntitlement),
					"username":    []byte("test-username"),
					"password":    []byte("test-password"),
				},
			},
			valid: false,
		},
		{
			name: "valid entitlement",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: entitlementSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"entitlement": []byte(validEntitlement),
					"username":    []byte("test-username"),
					"password":    []byte("test-password"),
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(tt.secret)

			err := verifyEntitlementFile(clientset)
			valid := true
			if err != nil {
				valid = false
			}
			assert.Equal(t, tt.valid, valid, "error verifying entitlement")
		})
	}

}
