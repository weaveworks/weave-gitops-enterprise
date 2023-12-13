package steps

import "fmt"

const (
	bootstrapFluxMsg = "do you want to bootstrap flux using the generic way"
)

var (
	bootstrapFluxQuestion = StepInput{
		Name:         inBootstrapFlux,
		Type:         confirmInput,
		Msg:          bootstrapFluxMsg,
		Enabled:      canAskForFluxBootstrap,
		DefaultValue: confirmNo,
	}
)

// NewAskBootstrapFluxStep step for asking if user want to install flux using generic method
func NewAskBootstrapFluxStep(config Config) BootstrapStep {
	if config.BootstrapFlux {
		bootstrapFluxQuestion.DefaultValue = confirmYes
	}
	return BootstrapStep{
		Name: "bootstrap flux",
		Input: []StepInput{
			bootstrapFluxQuestion,
		},
		Step: askBootstrapFlux,
	}
}

func askBootstrapFlux(input []StepInput, c *Config) ([]StepOutput, error) {
	if !canAskForFluxBootstrap(input, c) {
		return []StepOutput{}, nil
	}
	if c.BootstrapFlux && c.ModesConfig.Silent {
		c.Logger.Actionf("bootstrapping flux in the generic way")
		return []StepOutput{}, nil
	}
	for _, param := range input {
		if param.Name == inBootstrapFlux {
			fluxBootstrapRes, ok := param.Value.(string)
			if ok {
				if fluxBootstrapRes != "y" {
					return []StepOutput{}, fmt.Errorf("flux error: %s", fluxFatalErrorMsg)
				}

			}
		}
	}

	if c.ModesConfig.Export {
		return []StepOutput{}, fmt.Errorf("cannot execute with export mode")
	}

	return []StepOutput{}, nil
}

// canAskForGitConfig if fluxInstallation is false, then can ask for git config
func canAskForFluxBootstrap(input []StepInput, c *Config) bool {
	return !c.FluxInstalled
}
