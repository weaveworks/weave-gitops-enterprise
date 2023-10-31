package steps

import (
	"fmt"
)

const (
	gitAuthTypeMsg      = "please select your git authentication method"
	gitAuthTypeStepName = "git authentication"

	gitRepoBranchMsg = "please enter your git repository branch (default: main)"
	gitRepoPathMsg   = "please enter your path for your cluster (default: clusters/my-cluster)"
)

var (
	gitAuthTypes = []string{
		sshAuthType,
		httpsAuthType,
	}

	getGitAuthType = StepInput{
		Name:    gitAuthType,
		Type:    multiSelectionChoice,
		Msg:     gitAuthTypeMsg,
		Values:  gitAuthTypes,
		Enabled: canAskForGitConfig,
	}
)

func NewSelectGitAuthType(config Config) BootstrapStep {
	inputs := []StepInput{}

	switch config.AuthType {
	case sshAuthType:
		break
	case httpsAuthType:
		break
	default:
		inputs = append(inputs, getGitAuthType)
	}

	return BootstrapStep{
		Name:  gitAuthTypeStepName,
		Input: inputs,
		Step:  selectGitAuthType,
	}
}

func selectGitAuthType(input []StepInput, c *Config) ([]StepOutput, error) {
	if !canAskForGitConfig(input, c) {
		return []StepOutput{}, nil
	}
	for _, param := range input {
		if param.Name == gitAuthType {
			gitType, ok := param.Value.(string)
			if !ok {
				return []StepOutput{}, fmt.Errorf("unexpected error occurred. %s is not found", gitAuthType)
			}
			c.GitAuthType = gitType
		}
	}
	if c.GitAuthType == "" {
		return []StepOutput{}, fmt.Errorf("unexpected error occurred. %s is not found", gitAuthType)
	}
	c.Logger.Successf("git auth type: %s", c.AuthType)

	return []StepOutput{}, nil
}

// if fluxInstallation is false, then can ask for git config
func canAskForGitConfig(input []StepInput, c *Config) bool {
	return !c.FluxInstallated
}
