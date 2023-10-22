package steps

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckEntitlementFile test CheckEntitlementFile
func TestCheckEntitlementFile(t *testing.T) {
	var (
		expiredEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNjMxMzYxMjg2LCJpYXQiOjE2MzEyNzQ4ODYsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxMjc0ODg2LCJzdWIiOiJ0ZXN0QHdlYXZlLndvcmtzIn0.EKGp89DFcRKZ_kGmC8FuLVPB0wiab2KddkQKAmVNC9UH459v63tCP13eFybx9dAmMuaC77SA8rp7ukN1qZM7DA`
		invalidEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNjMxMzYxNDkwLCJpYXQiOjE2MzEyNzUwOTAsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxMjc1MDkwLCJzdWIiOiJ0ZXN0QHdlYXZlLndvcmtzIn0.E3Kfg4YzDOYJsTN9lD6B4uoW29tE0IB9X7lOpirSTwcZ7vVHk5PUXznYdiPIi9aSgLGAPIQL3YkAM4lyft3BDg`
	)

	tests := []struct {
		name   string
		secret *v1.Secret
		err    bool
	}{
		{
			name:   "secret does not exist",
			secret: &v1.Secret{},
			err:    true,
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
			err: true,
		},
		{
			name: "expired entitlement",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: entitlementSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"entitlement": []byte(expiredEntitlement),
					"username":    []byte("test-username"),
					"password":    []byte("test-password"),
				},
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := MakeTestConfig(t, Config{}, tt.secret)
			if err != nil {
				t.Fatalf("error creating config: %v", err)
			}
			_, err = checkEntitlementSecret([]StepInput{}, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error validating entitlement: %v", err)
			}
		})
	}

}
