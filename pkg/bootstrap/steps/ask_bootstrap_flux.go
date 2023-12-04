package steps

import "fmt"

const (
	bootstrapFluxMsg = "do you want to bootstrap flux using the generic way"
)

var (
	bootstrapFLuxQuestion = StepInput{
		Name:         inBootstrapFlux,
		Type:         confirmInput,
		Msg:          bootstrapFluxMsg,
		Enabled:      canAskForFluxBootstrap,
		DefaultValue: confirmNo,
	}
)

// NewAskBootstrapFluxStep step for asking if user want to install flux using generic method
func NewAskBootstrapFluxStep(config Config) BootstrapStep {
	return BootstrapStep{
		Name: "bootstrap flux",
		Input: []StepInput{
			bootstrapFLuxQuestion,
		},
		Step: askBootstrapFlux,
	}
}

func askBootstrapFlux(input []StepInput, c *Config) ([]StepOutput, error) {
	if !canAskForFluxBootstrap(input, c) {
		return []StepOutput{}, nil
	}
	if c.BootstrapFlux && c.Silent {
		c.Logger.Generatef("bootstrapping flux in the generic way")
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
	return []StepOutput{}, nil
}

// canAskForGitConfig if fluxInstallation is false, then can ask for git config
func canAskForFluxBootstrap(input []StepInput, c *Config) bool {
	return !c.FluxInstallated
}
