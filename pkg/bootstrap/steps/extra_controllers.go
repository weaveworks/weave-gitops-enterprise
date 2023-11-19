package steps

import (
	"fmt"
)

const (
	extraControllersMsg = "do you want to install extra controllers from the following on your cluster"
)

const (
	defaultController     = "none"
	policyAgentController = "policy-agent"
	tfController          = "tf-controller"
	capiController        = "capi"
	allControllers        = "all of above"
)

// NewInstallExtraControllers start installing extra controllers
func NewInstallExtraControllers(config Config) BootstrapStep {
	inputs := []StepInput{}
	controllersValues := []string{
		defaultController,
		policyAgentController,
		tfController,
		capiController,
		allControllers,
	}

	installExtraControllersStep := StepInput{
		Name:         inExtraControllers,
		Type:         multiSelectionChoice,
		Msg:          extraControllersMsg,
		Values:       controllersValues,
		DefaultValue: controllersValues[0],
	}

	if len(config.ExtraControllers) < 1 {
		inputs = append(inputs, installExtraControllersStep)
	}

	return BootstrapStep{
		Name:  "install extra controllers",
		Input: inputs,
		Step:  installExtraControllers,
	}
}

func installExtraControllers(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == inExtraControllers {
			extraControllers, ok := param.Value.(string)
			if ok {
				c.ExtraControllers = append(c.ExtraControllers, extraControllers)
			}
		}
	}
	for _, controller := range c.ExtraControllers {
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
		case allControllers:
			agentStep := NewInstallPolicyAgentStep(*c)
			_, err := agentStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install policy agent: %v", err)
			}
			tfControllerStep := NewInstallTFControllerStep(*c)
			_, err = tfControllerStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install tf controller: %v", err)
			}
			capiStep := NewInstallCapiControllerStep(*c)
			_, err = capiStep.Execute(c)
			if err != nil {
				return []StepOutput{}, fmt.Errorf("can't install capi controller: %v", err)
			}
		default:
			c.Logger.Successf("skipping installing controllers, selected: %s", controller)
			return []StepOutput{}, nil
		}
	}

	return []StepOutput{}, nil
}
