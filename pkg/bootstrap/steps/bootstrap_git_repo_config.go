package steps

import (
	"fmt"
	"net/url"
)

const (
	// repo configurations
	gitRepoURLMsg    = "please enter your flux git https or ssh repository url"
	gitRepoBranchMsg = "please enter your flux git repository branch (default: main)"
	gitRepoPathMsg   = "please enter your flux path for your cluster (default: clusters/my-cluster)"
)

const (
	repoConfigStepName = "flux repository configuration"
)

var (
	getRepoURL = StepInput{
		Name:         inRepoURL,
		Type:         stringInput,
		Msg:          gitRepoURLMsg,
		DefaultValue: "",
		Enabled:      canAskForFluxBootstrap,
	}

	getRepoBranch = StepInput{
		Name:         inBranch,
		Type:         stringInput,
		Msg:          gitRepoBranchMsg,
		DefaultValue: defaultBranch,
		Enabled:      canAskForFluxBootstrap,
	}

	getRepoPath = StepInput{
		Name:         inRepoPath,
		Type:         stringInput,
		Msg:          gitRepoPathMsg,
		DefaultValue: defaultPath,
		Enabled:      canAskForFluxBootstrap,
	}
)

type GitRepositoryConfig struct {
	Url    string
	Branch string
	Path   string
	Scheme string
}

// NewGitRepositoryConfig creates new configuration out of the user input and discovered state
func NewGitRepositoryConfig(url string, branch string, path string) (GitRepositoryConfig, error) {
	var scheme string
	var err error

	if url != "" {
		scheme, err = parseRepoScheme(url)
		if err != nil {
			return GitRepositoryConfig{}, fmt.Errorf("error parsing repo scheme: %v", err)
		}
	}

	return GitRepositoryConfig{
		Url:    url,
		Branch: branch,
		Path:   path,
		Scheme: scheme,
	}, nil

}

// NewGitRepositoryConfig step to configure the flux git repository
func NewGitRepositoryConfigStep(config Config) BootstrapStep {
	// create steps
	inputs := []StepInput{}
	if config.RepoURL == "" {
		inputs = append(inputs, getRepoURL)
	}

	if config.Branch == "" {
		inputs = append(inputs, getRepoBranch)
	}

	if config.RepoPath == "" {
		inputs = append(inputs, getRepoPath)
	}

	return BootstrapStep{
		Name:  repoConfigStepName,
		Input: inputs,
		Step:  createGitRepositoryConfig,
	}
}

func createGitRepositoryConfig(input []StepInput, c *Config) ([]StepOutput, error) {

	var repoURL = c.GitRepository.Url
	var repoBranch = c.GitRepository.Branch
	var repoPath = c.RepoPath

	for _, param := range input {
		if param.Name == inRepoURL {
			url, ok := param.Value.(string)
			if ok {
				repoURL = url
			}
		}
		if param.Name == inBranch {
			branch, ok := param.Value.(string)
			if ok {
				repoBranch = branch
			}
		}

		if param.Name == inRepoPath {
			path, ok := param.Value.(string)
			if ok {
				repoPath = path
			}
		}
	}

	repoConfig, err := NewGitRepositoryConfig(repoURL, repoBranch, repoPath)
	if err != nil {
		return nil, fmt.Errorf("error creating git repository configuration:%v", err)
	}
	c.GitRepository = repoConfig
	c.Logger.Actionf("configured repo: %s", c.GitRepository.Url)
	return []StepOutput{}, nil
}

func parseRepoScheme(repoURL string) (string, error) {
	repositoryURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("incorrect repository url %s:%v", repoURL, err)
	}
	var scheme string
	switch repositoryURL.Scheme {
	case "":
		return "", fmt.Errorf("repository scheme cannot be empty")
	case sshScheme:
		scheme = sshScheme
	case httpsScheme:
		scheme = httpsScheme
	default:
		return "", fmt.Errorf("unsupported repository scheme: %s", repositoryURL.Scheme)
	}
	return scheme, nil
}
