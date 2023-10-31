package steps

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	// ssh authentication
	gitSSHRepoURLMsg      = "please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)"
	privateKeyMsg         = "private key path and password\nDisclaimer: private key will be used to push WGE resources into the default repository only. It won't be stored or used anywhere else for any reason."
	privateKeyPathMsg     = "private key path"
	privateKeyPasswordMsg = "private key password"
)
const (
	sshAuthStepName = "git SSH config"
	defaultBranch   = "main"
	defaultPath     = "clusters/my-cluster"
)

var (
	privateKeyDefaultPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
)

var (
	getSSHRepoURL = StepInput{
		Name:         sshRepoURL,
		Type:         stringInput,
		Msg:          gitSSHRepoURLMsg,
		DefaultValue: "",
		Enabled:      canAskForSSHGitConfig,
	}

	getSSHRepoBranch = StepInput{
		Name:         branch,
		Type:         stringInput,
		Msg:          gitRepoBranchMsg,
		DefaultValue: defaultBranch,
		Enabled:      canAskForSSHGitConfig,
	}

	getSSHRepoPath = StepInput{
		Name:         repoPath,
		Type:         stringInput,
		Msg:          gitRepoPathMsg,
		DefaultValue: defaultPath,
		Enabled:      canAskForSSHGitConfig,
	}

	getKeyPath = StepInput{
		Name:         PrivateKeyPath,
		Type:         stringInput,
		Msg:          privateKeyPathMsg,
		DefaultValue: privateKeyDefaultPath,
		Enabled:      canAskForSSHGitCreds,
	}

	getKeyPassword = StepInput{
		Name:         PrivateKeyPassword,
		Type:         passwordInput,
		Msg:          privateKeyPasswordMsg,
		DefaultValue: "",
		Enabled:      canAskForSSHGitCreds,
	}
)

// NewBootstrapFluxUsingSSH step to bootstrap flux and configuring git using ssh
func NewBootstrapFluxUsingSSH(config Config) BootstrapStep {
	// create steps
	inputs := []StepInput{}
	if config.SSHRepoURL == "" {
		inputs = append(inputs, getSSHRepoURL)
	}

	if config.Branch == "" {
		inputs = append(inputs, getSSHRepoBranch)
	}

	if config.RepoPath == "" {
		inputs = append(inputs, getSSHRepoPath)
	}

	if config.PrivateKeyPath == "" {
		inputs = append(inputs, getKeyPath)
	}
	if config.PrivateKeyPassword == "" {
		inputs = append(inputs, getKeyPassword)
	}

	return BootstrapStep{
		Name:  sshAuthStepName,
		Input: inputs,
		Step:  createGitSSHConfig,
	}
}

func createGitSSHConfig(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == sshRepoURL {
			repoURL, ok := param.Value.(string)
			if ok {
				c.SSHRepoURL = repoURL
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

		if param.Name == PrivateKeyPath {
			privateKeyPath, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPath = privateKeyPath
			}
		}
		if param.Name == PrivateKeyPassword {
			privateKeyPassword, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPassword = privateKeyPassword
			}
		}
	}
	if !canAskForSSHGitConfig(input, c) {
		return []StepOutput{}, nil
	}
	c.Logger.Waitingf("bootstrapping flux ...")
	err := bootstrapFluxSSH(c)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to bootstrap flux: %v", err)
	}

	return []StepOutput{}, nil
}

func bootstrapFluxSSH(c *Config) error {
	var runner runner.CLIRunner
	out, err := runner.Run("flux",
		"bootstrap",
		"git",
		"--url", c.SSHRepoURL,
		"--branch", c.Branch,
		"--path", c.RepoPath,
		"--private-key-file", c.PrivateKeyPath,
		"--gpg-passphrase", c.PrivateKeyPassword,
		"-s",
	)

	if err != nil {
		return fmt.Errorf("%v:%v", err, string(out))
	}
	c.Logger.Successf("successfully bootstrapped flux!")
	return nil
}

// canAskForSSHGitConfig check when ask for gitconfig
func canAskForSSHGitConfig(input []StepInput, c *Config) bool {
	if !c.FluxInstallated {
		return c.GitAuthType == sshAuthType
	}
	return false
}

// canAskForSSHGitCreds check when ask for gitconfig
func canAskForSSHGitCreds(input []StepInput, c *Config) bool {
	return c.GitAuthType == sshAuthType

}
