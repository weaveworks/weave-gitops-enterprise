package steps

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	// ssh authentication
	privateKeyMsg         = "private key path and password\nDisclaimer: private key will be used to push WGE resources into the default repository only. It won't be stored or used anywhere else for any reason."
	privateKeyPathMsg     = "private key path"
	privateKeyPasswordMsg = "private key password"

	// https authentication
	gitUserNameMsg = "please enter your git username"
	gitPasswordMsg = "please enter your git authentication password/token with valid creds"
)
const (
	bootstrapFluxStepName = "git credentials"
	defaultBranch         = "main"
	defaultPath           = "clusters/my-cluster"
)

var (
	privateKeyDefaultPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
)

var (
	getKeyPath = StepInput{
		Name:         inPrivateKeyPath,
		Type:         stringInput,
		Msg:          privateKeyPathMsg,
		DefaultValue: privateKeyDefaultPath,
		Enabled:      canAskForSSHGitConfig,
	}
	getKeyPassword = StepInput{
		Name:         inPrivateKeyPassword,
		Type:         passwordInput,
		Msg:          privateKeyPasswordMsg,
		DefaultValue: "",
		Enabled:      canAskForSSHGitConfig,
	}

	getGitUsername = StepInput{
		Name:         inGitUserName,
		Type:         stringInput,
		Msg:          gitUserNameMsg,
		DefaultValue: "",
		Enabled:      canAskForHTTPSGitConfig,
	}
	getGitPassword = StepInput{
		Name:         inGitPassword,
		Type:         passwordInput,
		Msg:          gitPasswordMsg,
		DefaultValue: "",
		Enabled:      canAskForHTTPSGitConfig,
		Required:     true,
	}
)

// NewBootstrapFlux step to bootstrap flux and configuring git creds
func NewBootstrapFlux(config Config) BootstrapStep {
	// create steps
	inputs := []StepInput{}

	if config.PrivateKeyPath == "" {
		inputs = append(inputs, getKeyPath)
	}
	if config.PrivateKeyPassword == "" {
		inputs = append(inputs, getKeyPassword)
	}

	if config.GitUsername == "" {
		inputs = append(inputs, getGitUsername)
	}
	if config.GitToken == "" {
		inputs = append(inputs, getGitPassword)
	}

	return BootstrapStep{
		Name:  bootstrapFluxStepName,
		Input: inputs,
		Step:  configureFluxCreds,
	}
}

func configureFluxCreds(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		// process ssh
		if param.Name == inPrivateKeyPath {
			privateKeyPath, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPath = privateKeyPath
			}
		}
		if param.Name == inPrivateKeyPassword {
			privateKeyPassword, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPassword = privateKeyPassword
			}
		}

		// process https
		if param.Name == inGitUserName {
			username, ok := param.Value.(string)
			if ok {
				c.GitUsername = username
			}
		}
		if param.Name == inGitPassword {
			token, ok := param.Value.(string)
			if ok {
				c.GitToken = token
			}
		}
	}
	if !canAskForFluxBootstrap(input, c) {
		return []StepOutput{}, nil
	}

	c.Logger.Waitingf("bootstrapping flux ...")
	err := bootstrapFlux(c)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to bootstrap flux: %v", err)
	}

	return []StepOutput{}, nil
}

func bootstrapFlux(c *Config) error {
	var runner runner.CLIRunner
	var out []byte
	var err error

	switch c.GitRepository.Scheme {
	case sshScheme:
		out, err = runner.Run("flux",
			"bootstrap",
			"git",
			"--url", c.RepoURL,
			"--branch", c.Branch,
			"--path", c.RepoPath,
			"--private-key-file", c.PrivateKeyPath,
			"--password", c.PrivateKeyPassword,
			"-s",
		)
	case httpsScheme:
		out, err = runner.Run("flux",
			"bootstrap",
			"git",
			"--url", c.RepoURL,
			"--branch", c.Branch,
			"--path", c.RepoPath,
			"--username", c.GitUsername,
			"--password", c.GitToken,
			"--token-auth", "true",
			"-s",
		)
	}
	if err != nil {
		return fmt.Errorf("%v:%v", err, string(out))
	}

	c.Logger.Successf("successfully bootstrapped flux!")
	return nil
}

// canAskForSSHGitConfig check when ask for gitconfig when ssh scheme is enabled
func canAskForSSHGitConfig(input []StepInput, c *Config) bool {
	return c.GitRepository.Scheme == sshScheme
}

// canAskForHTTPSGitConfig check when ask for gitconfig when https scheme is enabled
func canAskForHTTPSGitConfig(input []StepInput, c *Config) bool {
	return c.GitRepository.Scheme == httpsScheme

}
