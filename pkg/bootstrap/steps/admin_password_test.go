package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewAskAdminCredsSecretStep(t *testing.T) {
	tests := []struct {
		name string

		config ClusterUserAuthConfig
		modes  ModesConfig
		silent bool

		want    BootstrapStep
		wantErr string
	}{
		{
			name: "should create step with create password input if password is not resolved",
			modes: ModesConfig{
				Silent: false,
			},
			config: ClusterUserAuthConfig{
				ExistCredentials: false,
			},
			want: BootstrapStep{
				Name: "user authentication",
				Input: []StepInput{
					createPasswordInput,
				},
			},
		},
		{
			name: "should create step without inputs if password resolved",
			modes: ModesConfig{
				Silent: false,
			},
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
			name: "should create step with update password input if credentials exist",
			modes: ModesConfig{
				Silent: false,
			},
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
			},
			want: BootstrapStep{
				Name:  "user authentication",
				Input: []StepInput{updatePasswordInput},
			},
		},
		{
			name: "should create step without inputs if non interactive",
			modes: ModesConfig{
				Silent: true,
			},
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
			name: "should fail if trying to update non interactive as updates are not supported",
			modes: ModesConfig{
				Silent: true,
			},
			config: ClusterUserAuthConfig{
				ExistCredentials: true,
				Password:         "password123",
			},
			want:    BootstrapStep{},
			wantErr: "admin login credentials already exist on the cluster. To reset admin credentials please remove secret 'cluster-user-auth' in namespace 'flux-system'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAskAdminCredsSecretStep(tt.config, tt.modes)

			if tt.wantErr != "" {
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("got error %q, want %q", msg, tt.wantErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Name, got.Name)
			if diff := cmp.Diff(tt.want.Input, got.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}

func TestAskAdminCredsSecretStep_Execute(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() (BootstrapStep, Config)
		config          Config
		wantOutput      []StepOutput
		wantErrorString string
	}{
		{
			name: "should create cluster user non-interactive",
			setup: func() (BootstrapStep, Config) {
				config := MakeTestConfig(t, Config{
					ClusterUserAuth: ClusterUserAuthConfig{
						Password: "password123",
					},
					ModesConfig: ModesConfig{
						Silent: true,
					},
				})
				step, err := NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.ModesConfig)
				assert.NoError(t, err)
				return step, config
			},
			wantOutput: []StepOutput{
				{
					Name: "cluster-user-auth",
					Type: "secret",
					Value: v1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster-user-auth",
							Namespace: "flux-system",
						},
					},
				},
			},
		},
		{
			name: "should create cluster user interactive",
			setup: func() (BootstrapStep, Config) {
				// Your predefined input strings
				inputStrings := []string{"password123\n"}
				config := MakeTestConfig(t, Config{
					InReader: &utils.MockReader{Inputs: inputStrings},
				})
				step, err := NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.ModesConfig)
				assert.NoError(t, err)
				return step, config
			},
			wantOutput: []StepOutput{
				{
					Name: "cluster-user-auth",
					Type: "secret",
					Value: v1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cluster-user-auth",
							Namespace: "flux-system",
						},
					},
				},
			},
		},
		{
			name: "should not support updates non-interactive",
			setup: func() (BootstrapStep, Config) {
				// secret exists
				secret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "cluster-user-auth", Namespace: WGEDefaultNamespace},
					Type:       "Opaque",
					Data: map[string][]byte{
						"username": []byte("wego-admin"),
						"password": []byte("test-password"),
					},
				}
				config := MakeTestConfig(t, Config{
					ModesConfig: ModesConfig{
						Silent: true,
					},
				}, secret)
				// user flags that wants to update the password
				config.ClusterUserAuth.Password = "new-password"
				step, err := NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.ModesConfig)
				assert.NoError(t, err)
				return step, config
			},
			wantErrorString: "cannot process output 'user authentication': secrets \"cluster-user-auth\" already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, config := tt.setup()
			gotOutputs, err := step.Execute(&config)
			if tt.wantErrorString != "" {
				assert.EqualError(t, err, tt.wantErrorString)
				return
			}
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.wantOutput, gotOutputs, cmpopts.IgnoreFields(v1.Secret{}, "Data")); diff != "" {
				t.Fatalf("expected output:\n%s", diff)
			}
		})
	}
}

func TestAskAdminCredsSecretStep_createCredentials(t *testing.T) {
	tests := []struct {
		name     string
		input    []StepInput
		password string
		config   Config
		output   []StepOutput
		err      bool
	}{
		{
			name:   "should error if trying to create credentials with invalid configuration",
			input:  []StepInput{},
			config: MakeTestConfig(t, Config{}),
			output: []StepOutput{},
			err:    true,
		},
		{
			name:     "should create secret for valid input password",
			password: "password",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "password",
				},
			},
			config: MakeTestConfig(t, Config{}),
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
			err: false,
		},
		{
			name:     "should create secret from input if password configuration exists",
			password: "passwordFromInput",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "passwordFromInput",
				},
			},
			config: MakeTestConfig(t, Config{
				ClusterUserAuth: ClusterUserAuthConfig{
					Password: "passwordFromConfig",
				},
			}),
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
			err: false,
		},
		{
			name:     "should create nothing if update but no new password exists",
			password: "passwordFromInput",
			input: []StepInput{
				{
					Name:  inPassword,
					Value: "",
				},
			},
			config: MakeTestConfig(t, Config{
				ClusterUserAuth: ClusterUserAuthConfig{
					ExistCredentials: true,
				},
			}),
			output: []StepOutput{},
			err:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := createCredentials(tt.input, &tt.config)
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
