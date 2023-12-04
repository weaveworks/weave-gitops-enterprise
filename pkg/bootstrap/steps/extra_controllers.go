package steps

import (
	"fmt"
)

const (
	extraComponentsMsg = "do you want to install extra Components from the following on your cluster"
)

const (
	policyAgentController = "policy-agent"
	tfController          = "tf-controller"
	capiController        = "capi"
)

// NewInstallExtraComponents start installing extra Components
func NewInstallExtraComponents(config Config) BootstrapStep {
	inputs := []StepInput{}
	controllersValues := []string{
		"",
		policyAgentController,
		tfController,
		capiController,
	}

	installExtraComponentsStep := StepInput{
		Name:         inExtraComponents,
		Type:         multiSelectionChoice,
		Msg:          extraComponentsMsg,
		Values:       controllersValues,
		DefaultValue: "",
	}

	if len(config.ExtraComponents) < 1 && !config.Silent {
		inputs = append(inputs, installExtraComponentsStep)
	}

	return BootstrapStep{
		Name:  "install extra components",
		Input: inputs,
		Step:  installExtraComponents,
	}
}

func installExtraComponents(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == inExtraComponents {
			extraComponents, ok := param.Value.(string)
			if ok {
				c.ExtraComponents = append(c.ExtraComponents, extraComponents)
			}
		}
	}
	for _, controller := range c.ExtraComponents {
		switch controller {
		case policyAgentController:
			agentStep := NewInstallPolicyAgentStep(*c)
			_, err := agentStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install policy agent: %v", err)
			}
		case tfController:
			tfControllerStep := NewInstallTFControllerStep(*c)
			_, err := tfControllerStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install tf controller: %v", err)
			}
		case capiController:
			capiStep := NewInstallCapiControllerStep(*c)
			_, err := capiStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install capi controller: %v", err)
			}
		default:
			c.Logger.Warningf("unsupported or empty controller, selected: %s", controller)
			return []StepOutput{}, nil
		}
	}

	return []StepOutput{}, nil
}
