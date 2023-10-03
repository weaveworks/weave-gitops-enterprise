package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	v1 "k8s.io/api/core/v1"
)

type BootstrapStep struct {
	Name   string
	Input  []StepInput
	Output []StepOutput
	Step   func(input []StepInput, c *Config) ([]StepOutput, error)
}

type StepInput struct {
	Name         string
	Msg          string
	Type         string
	DefaultValue string
	Value        any
}

type StepOutput struct {
	Name  string
	Type  string
	Value any
}

func (s BootstrapStep) Execute(c *Config) error {
	err := defaultInputStep(s.Input, c)
	if err != nil {
		return fmt.Errorf("cannot read input: %v", err)
	}

	outputs, err := s.Step(s.Input, c)
	if err != nil {
		return fmt.Errorf("cannot execute s: %v", err)
	}

	err = defaultOutputStep(outputs, c)
	if err != nil {
		return fmt.Errorf("cannot execute s: %v", err)
	}
	return nil
}

func defaultInputStep(params []StepInput, c *Config) error {
	fmt.Println("default input step")
	for _, param := range params {
		// handle global behaviour
		// for example silent

		// handle particular behaviours
		switch param.Type {
		case "string":
			if param.Value == "" {
				paramValue, err := utils.GetStringInput(param.Msg, param.DefaultValue)
				if err != nil {
					return err
				}
				param.Value = paramValue
			}
		case "secret":
			//read kubernetes secret
		default:
			return fmt.Errorf("not supported")
		}
	}
	return nil
}

// TODO here you handle
func defaultOutputStep(params []StepOutput, c *Config) error {
	fmt.Println("default input step")
	for _, param := range params {
		switch param.Type {
		case "secret":
			_ = param.Value.(v1.Secret)
			//TODO incomplete just to get the point
			if err := utils.CreateSecret(c.KubernetesClient, "secret", "", nil); err != nil {
				return err
			}
			//read kubernetes secret
		default:
			return fmt.Errorf("not supported")
		}
	}
	return nil
}
