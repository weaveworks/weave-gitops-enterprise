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

type StepInput struct {
	Name            string
	Msg             string
	StepInformation string
	Type            string
	DefaultValue    any
	Value           any
	Values          []string
	Valuesfn        func(input []StepInput, c *Config) (interface{}, error)
}

type StepOutput struct {
	Name  string
	Type  string
	Value any
}

func (s BootstrapStep) Execute(c *Config, flagsInput map[string]string) error {
	inputValues, err := defaultInputStep(s.Input, c, flagsInput)
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

func defaultInputStep(params []StepInput, c *Config, flagsInput map[string]string) ([]StepInput, error) {
	processedParams := []StepInput{}
	for _, param := range params {
		// handle global behaviour
		// for example silent

		// handle particular behaviours
		switch param.Type {
		case stringInput:
			// verify the input is enabled by executing the function
			enable := true
			if param.Valuesfn != nil {
				res, _ := param.Valuesfn(params, c)
				enable = res.(bool)
				if enable && param.StepInformation != "" {
					c.Logger.Warningf(param.StepInformation)
				}
			}
			// get the value from flags
			val, ok := flagsInput[param.Name]
			if ok && val != "" {
				param.Value = val
				enable = false
			}
			// get the value from user otherwise
			if param.Value == nil && enable {
				paramValue, err := utils.GetStringInput(param.Msg, param.DefaultValue.(string))
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
			}
			// fill the new params
			processedParams = append(processedParams, param)
		case passwordInput:
			// verify the input is enabled by executing the function
			enable := true
			if param.Valuesfn != nil {
				res, _ := param.Valuesfn(params, c)
				enable = res.(bool)
				if enable && param.StepInformation != "" {
					c.Logger.Warningf(param.StepInformation)
				}
			}
			// get the value from flags
			val, ok := flagsInput[param.Name]
			if ok && val != "" {
				param.Value = val
				enable = false
			}
			// get the value from user otherwise
			if param.Value == nil && enable {
				paramValue, err := utils.GetPasswordInput(param.Msg)
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
			}
			processedParams = append(processedParams, param)
		case confirmInput:
			// verify the input is enabled by executing the function
			enable := true
			if param.Valuesfn != nil {
				res, _ := param.Valuesfn(params, c)
				enable = res.(bool)
				if enable && param.StepInformation != "" {
					c.Logger.Warningf(param.StepInformation)
				}
			}
			// get the value from flags
			val, ok := flagsInput[param.Name]
			if ok && val != "" {
				param.Value = val
				enable = false
			}
			// get the value from user otherwise
			if param.Value == nil && enable {
				param.Value = utils.GetConfirmInput(param.Msg)
			}
			processedParams = append(processedParams, param)
		case multiSelectionChoice:
			// process the values from the function
			var values []string = param.Values
			if param.Valuesfn != nil {
				res, err := param.Valuesfn(params, c)
				if err != nil {
					return []StepInput{}, err
				}
				values = res.([]string)
			}
			// get the value from flags
			val, ok := flagsInput[param.Name]
			if ok && val != "" {
				param.Value = val
			}
			// get the values from user
			if param.Value == nil {
				paramValue, err := utils.GetSelectInput(param.Msg, values)
				if err != nil {
					return []StepInput{}, err
				}
				param.Value = paramValue
			}
			processedParams = append(processedParams, param)
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
		case typeSecret:
			secret, ok := param.Value.(v1.Secret)
			if !ok {
				return errors.New("unexpected error casting secret")
			}
			name := secret.ObjectMeta.Name
			namespace := secret.ObjectMeta.Namespace
			data := secret.Data
			if err := utils.CreateSecret(c.KubernetesClient, name, namespace, data); err != nil {
				return err
			}
		case typeFile:
			file, ok := param.Value.(fileContent)
			if !ok {
				return errors.New("unexpected error casting file")
			}
			pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace)
			if err != nil {
				return err
			}

			defer func() {
				err = utils.CleanupRepo()
				if err != nil {
					c.Logger.Failuref("failed to cleanup repo!")
				}
			}()

			err = utils.CreateFileToRepo(file.Name, file.Content, pathInRepo, file.CommitMsg)
			if err != nil {
				return err
			}

			if err := utils.ReconcileFlux(); err != nil {
				return err
			}

		case typePortforward:
			portforward, ok := param.Value.(func() error)
			if !ok {
				return errors.New("unexpected error for function casting")
			}
			err := portforward()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not supported")
		}
	}
	return nil
}
