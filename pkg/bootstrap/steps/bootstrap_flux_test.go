package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapFlux(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantInput []StepInput
	}{
		{
			name: "should not ask for key password if user introduced empty flag",
			config: MakeTestConfig(t, Config{
				PrivateKeyPassword:        "",
				PrivateKeyPasswordChanged: true,
			}),
			wantInput: []StepInput{
				getKeyPath,
				getGitUsername,
				getGitPassword,
			},
		},
		{
			name: "should ask for key password if user not introduced",
			config: MakeTestConfig(t, Config{
				PrivateKeyPassword:        "",
				PrivateKeyPasswordChanged: false,
			}),
			wantInput: []StepInput{
				getKeyPath,
				getKeyPassword,
				getGitUsername,
				getGitPassword,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBootstrapFlux(tt.config)
			opts := cmpopts.IgnoreFields(StepInput{}, "Enabled")
			if diff := cmp.Diff(tt.wantInput, got.Input, opts); diff != "" {
				t.Fatalf("unpected step inputs:\n%s", diff)
			}
		})
	}
}

func TestConfigureFluxCreds(t *testing.T) {

	tests := []struct {
		name     string
		input    []StepInput
		config   *Config
		askSSH   bool
		askHTTPS bool
		err      bool
	}{
		{
			name: "test with valid ssh scheme with flux installed",
			input: []StepInput{
				{
					Name:  inPrivateKeyPath,
					Value: "testkey",
				},
				{
					Name:  inPrivateKeyPassword,
					Value: "testpassword",
				},
			},
			config: &Config{
				FluxConfig: FluxConfig{
					IsInstalled: true,
				},
				GitRepository: GitRepositoryConfig{
					Scheme: sshScheme,
				},
			},
			askSSH:   true,
			askHTTPS: false,
			err:      false,
		},
		{
			name: "test with valid https scheme with flux installed",
			input: []StepInput{
				{
					Name:  inGitUserName,
					Value: "testgitusername",
				},
				{
					Name:  inGitPassword,
					Value: "testgittoken",
				},
			},
			config: &Config{
				FluxConfig: FluxConfig{
					IsInstalled: true,
				},
				GitRepository: GitRepositoryConfig{
					Scheme: httpsScheme,
				},
			},
			err:      false,
			askSSH:   false,
			askHTTPS: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MakeTestConfig(t, *tt.config)

			_, err := configureFluxCreds(tt.input, &config)
			if err != nil {
				if tt.err {
					assert.Error(t, err, "expected error")
					return
				}
				t.Fatalf("unexpected error occurred: %v", err)
			}
			askhttps := canAskForHTTPSGitConfig(tt.input, tt.config)
			ashssh := canAskForSSHGitConfig(tt.input, tt.config)
			assert.Equal(t, tt.askSSH, ashssh, "wrong method selection")
			assert.Equal(t, tt.askHTTPS, askhttps, "wrong method selection")
		})
	}
}
