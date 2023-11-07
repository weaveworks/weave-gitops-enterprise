package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	// https authentication
	gitHttpsRepoURLMsg = "please enter your git repository url (example: https://github.com/my-org-name/my-repo-name)"
	gitUserNameMsg     = "please enter your git username"
	gitTokenMsg        = "please enter your git authentication token with valid creds"
)

const (
	httpsAuthStepName = "git https config"
)

var (
	getHttpsRepoURL = StepInput{
		Name:         httpsRepoURL,
		Type:         stringInput,
		Msg:          gitHttpsRepoURLMsg,
		DefaultValue: "",
		Enabled:      canAskForHttpsGitConfig,
	}

	getHttpsRepoBranch = StepInput{
		Name:         branch,
		Type:         stringInput,
		Msg:          gitRepoBranchMsg,
		DefaultValue: defaultBranch,
		Enabled:      canAskForHttpsGitConfig,
	}

	getHttpsRepoPath = StepInput{
		Name:         repoPath,
		Type:         stringInput,
		Msg:          gitRepoPathMsg,
		DefaultValue: defaultPath,
		Enabled:      canAskForHttpsGitConfig,
	}

	getGitUsername = StepInput{
		Name:         gitUserName,
		Type:         stringInput,
		Msg:          gitUserNameMsg,
		DefaultValue: "",
		Enabled:      canAskForHttpsGitCreds,
	}

	getGitToken = StepInput{
		Name:         gitToken,
		Type:         passwordInput,
		Msg:          gitTokenMsg,
		DefaultValue: "",
		Enabled:      canAskForHttpsGitCreds,
		Required:     true,
	}
)

// NewBootstrapFluxUsingHTTPS step to bootstrap flux and configuring git using https
func NewBootstrapFluxUsingHTTPS(config Config) BootstrapStep {
	// create steps
	inputs := []StepInput{}
	if config.HttpsRepoURL == "" {
		inputs = append(inputs, getHttpsRepoURL)
	}

	if config.Branch == "" {
		inputs = append(inputs, getHttpsRepoBranch)
	}

	if config.RepoPath == "" {
		inputs = append(inputs, getHttpsRepoPath)
	}

	if config.GitUsername == "" {
		inputs = append(inputs, getGitUsername)
	}
	if config.GitToken == "" {
		inputs = append(inputs, getGitToken)
	}

	return BootstrapStep{
		Name:  httpsAuthStepName,
		Input: inputs,
		Step:  createGitHttpsConfig,
	}
}

func createGitHttpsConfig(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == httpsRepoURL {
			repoURL, ok := param.Value.(string)
			if ok {
				c.HttpsRepoURL = repoURL
			}
		}
		if param.Name == branch {
			repoBranch, ok := param.Value.(string)
			if ok {
				c.Branch = repoBranch
			}
		}

		if param.Name == repoPath {
			path, ok := param.Value.(string)
			if ok {
				c.RepoPath = path
			}
		}

		if param.Name == gitUserName {
			username, ok := param.Value.(string)
			if ok {
				c.GitUsername = username
			}
		}

		if param.Name == gitToken {
			token, ok := param.Value.(string)
			if ok {
				c.GitToken = token
			}
		}
	}
	if !canAskForHttpsGitConfig(input, c) {
		return []StepOutput{}, nil
	}
	c.Logger.Waitingf("bootstrapping flux ...")
	err := bootstrapFluxHttps(c)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to bootstrap flux: %v", err)
	}

	return []StepOutput{}, nil
}

func bootstrapFluxHttps(c *Config) error {
	var runner runner.CLIRunner
	out, err := runner.Run("flux",
		"bootstrap",
		"git",
		"--url", c.HttpsRepoURL,
		"--branch", c.Branch,
		"--path", c.RepoPath,
		"--username", c.GitUsername,
		"--password", c.GitToken,
		"--token-auth", "true",
		"-s",
	)

	if err != nil {
		return fmt.Errorf("%v:%v", err, string(out))
	}
	c.Logger.Successf("successfully bootstrapped flux!")
	return nil
}

// canAskForHttpsGitConfig check when ask for gitconfig
func canAskForHttpsGitConfig(input []StepInput, c *Config) bool {
	if !c.FluxInstallated {
		return c.GitAuthType == httpsAuthType
	}
	return false
}

// canAskForHttpsGitCreds check when ask for gitconfig
func canAskForHttpsGitCreds(input []StepInput, c *Config) bool {
	return c.GitAuthType == httpsAuthType
}
