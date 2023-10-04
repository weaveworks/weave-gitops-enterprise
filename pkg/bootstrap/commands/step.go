package commands

import (
	"errors"
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

func (s BootstrapStep) WithConfig(config map[string]any) ([]StepInput, error) {
	requiredInput :=[]StepInput{}
	for _, input := range s.Input {
		// check if exists
		//a := config[input.Name]
		// if does not exist we add it to requiredInput
		// otherwise we take
	}
}

type StepInput struct {
	Name         string
	Msg          string
	Type         string
	DefaultValue any
	Value        any
	Valuesfn     func(input []StepInput, c *Config) ([]string, error)
	Values       []string
}

type StepOutput struct {
	Name  string
	Type  string
	Value any
}

func (s BootstrapStep) Execute(input []StepInput c *Config) error {
	inputValues, err := defaultInputStep(input, c)
	if err != nil {
		return fmt.Errorf("cannot read input: %v", err)
	}

	outputs, err := s.Step(inputValues, c)
	if err != nil {
		return fmt.Errorf("cannot execute step: %v", err)
	}

	err = defaultOutputStep(outputs, c)
	if err != nil {
		return fmt.Errorf("cannot execute step: %v", err)
	}
	return nil
}

func defaultInputStep(params []StepInput, c *Config) ([]StepInput, error) {
	processedParams := []StepInput{}
	for _, param := range params {
		// handle global behaviour
		// for example silent

		// handle particular behaviours
		switch param.Type {
		case stringInput:
			if param.Value == nil {
				paramValue, err := utils.GetStringInput(param.Msg, param.DefaultValue.(string))
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
				processedParams = append(processedParams, param)
			}
		case passwordInput:
			if param.Value == nil {
				paramValue, err := utils.GetPasswordInput(param.Msg)
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
				processedParams = append(processedParams, param)
			}
		case multiSelectionChoice:
			if param.Value == nil {
				var values []string
				var err error
				if param.Valuesfn != nil {
					values, err = param.Valuesfn(params, c)
					if err != nil {
						return []StepInput{}, err
					}
				}
				paramValue, err := utils.GetSelectInput(param.Msg, values)
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
				processedParams = append(processedParams, param)
			}
		default:
			return []StepInput{}, errors.New("not supported")
		}
	}
	return processedParams, nil
}

func defaultOutputStep(params []StepOutput, c *Config) error {
	for _, param := range params {
		switch param.Type {
		case successMsg:
			c.Logger.Successf(param.Value.(string))
		case "secret":
			secret := param.Value.(v1.Secret)
			name := secret.ObjectMeta.Name
			namespace := secret.ObjectMeta.Namespace
			data := secret.Data
			if err := utils.CreateSecret(c.KubernetesClient, name, namespace, data); err != nil {
				return err
			}
		default:
			return fmt.Errorf("not supported")
		}
	}
	return nil
}
