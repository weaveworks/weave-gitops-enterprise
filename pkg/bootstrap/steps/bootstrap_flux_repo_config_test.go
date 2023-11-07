package steps

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestCreateFluxRepositoryConfig(t *testing.T) {

	tests := []struct {
		name   string
		input  []StepInput
		config *Config
		err    bool
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
				RepoURL:   "ssh://git@github.com/my-org-name/my-repo-name",
				RepoPath:  "test/test",
				Branch:    "main",
				GitScheme: sshScheme,
			},
			err: false,
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
				RepoURL:   "https://github.com/my-org-name/my-repo-name",
				RepoPath:  "test/test",
				Branch:    "main",
				GitScheme: httpsScheme,
			},
			err: false,
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
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{})

			_, err := createFluxRepositoryConfig(tt.input, &config)
			if err != nil {
				if tt.err {
					assert.Error(t, err, "expected error")
					return
				}
				t.Fatalf("unexpected error occurred: %v", err)
			}
			assert.Equal(t, tt.config.RepoURL, config.RepoURL, "wrong repo url")
			assert.Equal(t, tt.config.RepoPath, config.RepoPath, "wrong repo path")
			assert.Equal(t, tt.config.Branch, config.Branch, "wrong repo branch")
			assert.Equal(t, tt.config.GitScheme, config.GitScheme, "wrong git scheme")
		})
	}
}
