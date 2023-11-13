package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateCredentials(t *testing.T) {
	tests := []struct {
		name       string
		secret     *v1.Secret
		input      []StepInput
		password   string
		output     []StepOutput
		isExisting bool
		canAsk     bool
		err        bool
	}{
		{
			name:     "secret doesn't exist",
			secret:   &v1.Secret{},
			password: "password",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "password",
				},
				{
					Name:  inExistingCreds,
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
							"username": []byte(defaultAdminUsername),
						},
					},
				},
			},
			err:        false,
			isExisting: false,
			canAsk:     true,
		},
		{
			name: "secret exist and user refuse to continue",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte(defaultAdminUsername),
					"password": []byte("test-password"),
				},
			},
			password: "password",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "password",
				},
				{
					Name:  inExistingCreds,
					Value: "n",
				},
			},
			err:        true,
			isExisting: true,
			canAsk:     true,
		},
		{
			name: "secret exist and user continue",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: adminSecretName, Namespace: WGEDefaultNamespace},
				Type:       "Opaque",
				Data: map[string][]byte{
					"username": []byte(defaultAdminUsername),
					"password": []byte("test-password"),
				},
			},
			password: "password",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "password",
				},
				{
					Name:  inExistingCreds,
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
							"username": []byte(defaultAdminUsername),
						},
					},
				},
			},
			err:        false,
			isExisting: true,
			canAsk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{}, tt.secret)

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
			isExisting := isExistingAdminSecret(config.KubernetesClient)
			assert.Equal(t, tt.isExisting, isExisting, "incorrect result")

			canAsk := canAskForCreds(tt.input, &config)
			assert.Equal(t, tt.canAsk, canAsk, "incorrect result")
		})
	}
}

func TestNewAskAdminCredsSecretStep(t *testing.T) {
	tests := []struct {
		name string

		config ClusterUserAuthConfig
		silent bool

		want    BootstrapStep
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "day0/1 interactive - ask suggest default - <no silent, no existing, no values >",
			silent: false,
			config: ClusterUserAuthConfig{
				ExistCredentials: false,
			},
			want: BootstrapStep{
				Name: "user authentication",
				Input: []StepInput{
					getPasswordWithDefaultInput,
				},
			},
		},
		{
			name:   "day1 interactive - no ask use input - <no silent, no existing, values>",
			silent: false,
			config: ClusterUserAuthConfig{
				ExistCredentials: false,
				Password:         "password123",
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{},
			},
		},
		{
			name:   "day1 interactive - ask suggest previous value - <no silent, existing, no values>",
			silent: false,
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{getPasswordWithExistingAndUserInput},
			},
		},
		{
			name:   "day1 interactive - ask conflict - <no silent, existing, values>",
			silent: false,
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
				Password:         "password123",
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{getPasswordWithExistingAndUserInput},
			},
		},
		{
			name:   "day1 no-interactive - no ask use input - <silent, no existing, values>",
			silent: true,
			config: ClusterUserAuthConfig{
				ExistCredentials: false,
				Password:         "password123",
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{},
			},
		},
		{
			name:   "day1 no-interactive - no ask use existing - <silent, existing, no values>",
			silent: true,
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{},
			},
		},
		{
			name:   "day1 no-interactive - overwrite - <silent, existing, values>",
			silent: true,
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAskAdminCredsSecretStep(tt.config, tt.silent)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Name, got.Name)
			if diff := cmp.Diff(tt.want.Input, got.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}
