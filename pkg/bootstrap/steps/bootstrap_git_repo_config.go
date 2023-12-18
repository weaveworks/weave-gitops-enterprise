package steps

import (
	"fmt"
	"net/url"
	"strings"
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

	getRepoPathIn = StepInput{
		Name:         inRepoPath,
		Type:         stringInput,
		Msg:          gitRepoPathMsg,
		DefaultValue: defaultPath,
		Enabled:      canAskForFluxBootstrap,
	}
)

// GitRepositoryConfig contains the configuration for the configuration repo
type GitRepositoryConfig struct {
	// Url is the git repository url
	Url string
	// Branch is the git repository branch
	Branch string
	// Path is the git repository path
	Path string
	// Scheme is the git repository url scheme
	Scheme string
}

// NewGitRepositoryConfig creates new Git repository configuration from valid input parameters.
func NewGitRepositoryConfig(url string, branch string, path string, fluxConfig FluxConfig) (GitRepositoryConfig, error) {
	var scheme string
	var err error
	var normalisedUrl string

	// using flux config as we dont support updates
	if fluxConfig.IsInstalled {
		return GitRepositoryConfig{
			Url:    fluxConfig.Url,
			Scheme: fluxConfig.Scheme,
			Branch: fluxConfig.Branch,
			Path:   fluxConfig.Path,
		}, nil
	}

	if url != "" {
		normalisedUrl, scheme, err = normaliseUrl(url)
		if err != nil {
			return GitRepositoryConfig{}, fmt.Errorf("error parsing repo scheme: %v", err)
		}
	}

	return GitRepositoryConfig{
		Url:    normalisedUrl,
		Branch: branch,
		Path:   path,
		Scheme: scheme,
	}, nil

}

// normaliseUrl normalises the given url to meet standard URL syntax. The main motivation to have this function
// is to support Git server URLs in "shorter scp-like syntax for the SSH protocol" as described in https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols
// and followed by popular Git server providers like GitHub (git@github.com:weaveworks/weave-gitops.git) and GitLab (i.e. git@gitlab.com:gitlab-org/gitlab-foss.git).
// Returns the normalisedUrl, as well the scheme and an error if any.
func normaliseUrl(repoURL string) (normalisedUrl string, scheme string, err error) {
	// transform in case of ssh like git@github.com:username/repository.git
	if strings.Contains(repoURL, "@") && !strings.Contains(repoURL, "://") {
		repoURL = "ssh://" + strings.Replace(repoURL, ":", "/", 1)
	}

	repositoryURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("error parsing repository URL: %v", err)
	}

	switch repositoryURL.Scheme {
	case sshScheme:
		return repositoryURL.String(), sshScheme, nil
	case httpsScheme:
		return repositoryURL.String(), httpsScheme, nil
	default:
		return "", "", fmt.Errorf("invalid repository scheme: %s", repositoryURL.Scheme)
	}
}

// NewGitRepositoryConfig step to configure the flux git repository
func NewGitRepositoryConfigStep(config GitRepositoryConfig) BootstrapStep {
	// create steps
	inputs := []StepInput{}
	if config.Url == "" {
		inputs = append(inputs, getRepoURL)
	}

	if config.Branch == "" {
		inputs = append(inputs, getRepoBranch)
	}

	if config.Path == "" {
		inputs = append(inputs, getRepoPathIn)
	}

	return BootstrapStep{
		Name:  repoConfigStepName,
		Input: inputs,
		Step:  createGitRepositoryConfig,
	}
}

func createGitRepositoryConfig(input []StepInput, c *Config) ([]StepOutput, error) {

	if c.FluxConfig.IsInstalled {
		return []StepOutput{}, nil
	}

	var repoURL = c.GitRepository.Url
	var repoBranch = c.GitRepository.Branch
	var repoPath = c.GitRepository.Path

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

	repoConfig, err := NewGitRepositoryConfig(repoURL, repoBranch, repoPath, c.FluxConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating git repository configuration: %v", err)
	}
	c.GitRepository = repoConfig
	c.Logger.Actionf("configured repo: %s", c.GitRepository.Url)
	return []StepOutput{}, nil
}
