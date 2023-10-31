package steps

import "fmt"

const (
	bootstrapFluxMsg = "do you want to bootstrap flux using the generic way"
)

var (
	bootstrapFLuxQuestion = StepInput{
		Name:    bootstrapFlux,
		Type:    confirmInput,
		Msg:     bootstrapFluxMsg,
		Enabled: canAskForGitConfig,
	}
)

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
	if !canAskForGitConfig(input, c) {
		return []StepOutput{}, nil
	}
	for _, param := range input {
		if param.Name == bootstrapFlux {
			fluxBootstrapRes, ok := param.Value.(string)
			if ok {
				if fluxBootstrapRes != "y" {
					return []StepOutput{}, fmt.Errorf("flux bootstrapped error: %s", fluxRecoverMsg)
				}

			}
		}
	}
	return []StepOutput{}, nil
}
