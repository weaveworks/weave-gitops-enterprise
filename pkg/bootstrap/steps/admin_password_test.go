package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		secret   *v1.Secret
		input    []StepInput
		password string
		output   []StepOutput
		err      bool
	}{
		{
			name:     "secret doesn't exist",
			secret:   &v1.Secret{},
			password: "password",
			input: []StepInput{
				{
					Name:  UserName,
					Value: "wego-admin",
				},
				{
					Name:  Password,
					Value: "password",
				},
				{
					Name:  existingCreds,
					Value: false,
				},
			},
			output: []StepOutput{
				{
					Name: adminSecretName,
					Type: typeSecret,
					Value: v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      adminSecretName,
							Namespace: WGEDefaultNamespace,
						},
						Data: map[string][]byte{
							"username": []byte("wego-admin"),
						},
					},
				},
			},
			err: false,
		},
		{
			name: "secret exist and user refuse to continue",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte("test-username"),
					"password": []byte("test-password"),
				},
			},
			password: "password",
			input: []StepInput{
				{
					Name:  UserName,
					Value: "wego-admin",
				},
				{
					Name:  Password,
					Value: "password",
				},
				{
					Name:  existingCreds,
					Value: "n",
				},
			},
			err: true,
		},
		{
			name: "secret exist and user continue",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte("test-username"),
					"password": []byte("test-password"),
				},
			},
			password: "password",
			input: []StepInput{
				{
					Name:  UserName,
					Value: "wego-admin",
				},
				{
					Name:  Password,
					Value: "password",
				},
				{
					Name:  existingCreds,
					Value: "y",
				},
			},
			output: []StepOutput{
				{
					Name: adminSecretName,
					Type: typeSecret,
					Value: v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      adminSecretName,
							Namespace: WGEDefaultNamespace,
						},
						Data: map[string][]byte{
							"username": []byte("wego-admin"),
						},
					},
				},
			},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := MakeTestConfig(t, Config{}, tt.secret)
			if err != nil {
				t.Fatalf("error creating config: %v", err)
			}

			out, err := createCredentials(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error creating creds: %v", err)
			}

			for i, item := range out {
				assert.Equal(t, item.Name, tt.output[i].Name)
				assert.Equal(t, item.Type, tt.output[i].Type)
				outSecret, ok := item.Value.(v1.Secret)
				if !ok {
					t.Fatalf("failed getting result secret data")
				}
				inSecret, ok := tt.output[i].Value.(v1.Secret)
				if !ok {
					t.Fatalf("failed getting output secret data")
				}
				assert.Equal(t, outSecret.Name, inSecret.Name, "mismatch name")
				assert.Equal(t, outSecret.Namespace, inSecret.Namespace, "mismatch namespace")
				assert.Equal(t, outSecret.Data["username"], inSecret.Data["username"], "mismatch username")
				assert.NoError(t, bcrypt.CompareHashAndPassword(outSecret.Data["password"], []byte(tt.password)), "mismatch password")
			}
		})
	}
}
