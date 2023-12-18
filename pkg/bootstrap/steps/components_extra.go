package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentsExtraMsg = "do you want to install extra Components from the following on your cluster"
)

const (
	none                  = "none"
	policyAgentController = "policy-agent"
	tfController          = "tf-controller"
)

var ComponentsExtra = []string{
	none,
	policyAgentController,
	tfController,
}

// ComponentsExtraConfig contains the configuration for the extra components
type ComponentsExtraConfig struct {
	Requested []string
	Existing  []string
}

// NewInstallExtraComponentsConfig handles the extra components configurations
func NewInstallExtraComponentsConfig(components []string, client client.Client, fluxInstalled bool) (ComponentsExtraConfig, error) {
	// validate requested components against pre-defined ComponentsExtra
	config := ComponentsExtraConfig{
		Requested: components,
	}

	for _, component := range config.Requested {
		if !slices.Contains(ComponentsExtra, component) {
			return ComponentsExtraConfig{}, fmt.Errorf("unsupported component selected: %s", component)
		}
	}

	if fluxInstalled {
		// check existing components
		for _, component := range ComponentsExtra {
			version, err := utils.GetHelmReleaseProperty(client, component, WGEDefaultNamespace, utils.HelmVersionProperty)
			if err == nil && version != "" {
				config.Existing = append(config.Existing, component)
			}
		}
	}

	return config, nil
}

// NewInstallExtraComponents contains the extra components installation step
func NewInstallExtraComponentsStep(config ComponentsExtraConfig, silent bool) BootstrapStep {
	inputs := []StepInput{}

	installExtraComponentsStep := StepInput{
		Name:         inComponentsExtra,
		Type:         multiSelectionChoice,
		Msg:          componentsExtraMsg,
		Values:       ComponentsExtra,
		DefaultValue: none,
	}

	if len(config.Requested) < 1 && !silent {
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
		if param.Name == inComponentsExtra {
			componentsExtra, ok := param.Value.(string)
			if ok {
				c.ComponentsExtra.Requested = append(c.ComponentsExtra.Requested, componentsExtra)
			}
		}
	}
	for _, controller := range c.ComponentsExtra.Requested {
		switch controller {
		case none:
			return []StepOutput{}, nil
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
		default:
			return []StepOutput{}, fmt.Errorf("unsupported component selected: %s", controller)
		}
	}

	return []StepOutput{}, nil
}
