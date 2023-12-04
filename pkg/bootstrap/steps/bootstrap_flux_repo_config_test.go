package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGitRepositoryConfig(t *testing.T) {

	tests := []struct {
		name    string
		input   []StepInput
		config  *Config
		err     bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test with valid ssh repo scheme",
			input: []StepInput{
				{
					Name:  inRepoURL,
					Value: "ssh://git@github.com/my-org-name/my-repo-name",
				},
				{
					Name:  inBranch,
					Value: "main",
				},
				{
					Name:  inRepoPath,
					Value: "test/test",
				},
			},
			config: &Config{
				GitRepository: GitRepositoryConfig{
					Url:    "ssh://git@github.com/my-org-name/my-repo-name",
					Path:   "test/test",
					Branch: "main",
					Scheme: sshScheme,
				},
			},
			err:     false,
			wantErr: assert.NoError,
		},
		{
			name: "test with valid https repo scheme",
			input: []StepInput{
				{
					Name:  inRepoURL,
					Value: "https://github.com/my-org-name/my-repo-name",
				},
				{
					Name:  inBranch,
					Value: "main",
				},
				{
					Name:  inRepoPath,
					Value: "test/test",
				},
			},
			config: &Config{
				GitRepository: GitRepositoryConfig{
					Url:    "https://github.com/my-org-name/my-repo-name",
					Path:   "test/test",
					Branch: "main",
					Scheme: httpsScheme,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "test with invalid repo scheme",
			input: []StepInput{
				{
					Name:  inRepoURL,
					Value: "ssl://github.com/my-org-name/my-repo-name",
				},
				{
					Name:  inBranch,
					Value: "main",
				},
				{
					Name:  inRepoPath,
					Value: "test/test",
				},
			},
			config: &Config{
				GitRepository: GitRepositoryConfig{},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported repository scheme: ssl")
				return true
			},
		},
		{
			name: "test with empty repo scheme",
			input: []StepInput{
				{
					Name:  inRepoURL,
					Value: "git@github.com/my-org-name/my-repo-name",
				},
				{
					Name:  inBranch,
					Value: "main",
				},
				{
					Name:  inRepoPath,
					Value: "test/test",
				},
			},
			config: &Config{
				GitRepository: GitRepositoryConfig{},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "repository scheme cannot be empty")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{})
			_, err := createGitRepositoryConfig(tt.input, &config)
			if !tt.wantErr(t, err, "createGitRepositoryConfig") {
				return
			}
			assert.Equal(t, tt.config.RepoURL, config.RepoURL, "wrong repo url")
			assert.Equal(t, tt.config.RepoPath, config.RepoPath, "wrong repo path")
			assert.Equal(t, tt.config.Branch, config.Branch, "wrong repo branch")
			assert.Equal(t, tt.config.GitRepository.Scheme, config.GitRepository.Scheme, "wrong git scheme")
		})
	}
}
