package steps

import (
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	v1 "k8s.io/api/core/v1"
)

// BootstrapStep struct that defines the contract of a bootstrapping step.
// It is abstracted to have a generic way to handle them, so we could achieve easier
// extensibility, consistency and maintainability.
type BootstrapStep struct {
	Name   string
	Input  []StepInput
	Output []StepOutput
	Step   func(input []StepInput, c *Config) ([]StepOutput, error)
}

// StepInput represents an input an step requires to execute it. for example the
// user needs to introduce an string or a password.
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

// StepOutput represents an output generated out of the execution of a step.
// An example could be a helm release manifest for weave gitops.
type StepOutput struct {
	Name  string
	Type  string
	Value any
}

func (s BootstrapStep) Execute(c *Config) error {
	inputValues, err := defaultInputStep(s.Input, c)
	if err != nil {
		return fmt.Errorf("cannot process input '%s': %v", s.Name, err)
	}

	outputs, err := s.Step(inputValues, c)
	if err != nil {
		return fmt.Errorf("cannot execute '%s': %v", s.Name, err)
	}

	err = defaultOutputStep(outputs, c)
	if err != nil {
		return fmt.Errorf("cannot process output '%s': %v", s.Name, err)
	}
	return nil
}

func defaultInputStep(inputs []StepInput, c *Config) ([]StepInput, error) {
	processedInputs := []StepInput{}
	for _, input := range inputs {
		switch input.Type {
		case stringInput:
			// verify the input is enabled by executing the function
			enable := true
			if input.Valuesfn != nil {
				res, _ := input.Valuesfn(inputs, c)
				enable = res.(bool)
				if enable && input.StepInformation != "" {
					c.Logger.Warningf(input.StepInformation)
				}
			}
			// get the value from user otherwise
			if input.Value == nil && enable {
				paramValue, err := utils.GetStringInput(input.Msg, input.DefaultValue.(string))
				if err != nil {
					return []StepInput{}, err
				}
				input.Value = paramValue
			}
			// fill the new inputs
			processedInputs = append(processedInputs, input)
		case passwordInput:
			// verify the input is enabled by executing the function
			enable := true
			if input.Valuesfn != nil {
				res, _ := input.Valuesfn(inputs, c)
				enable = res.(bool)
				if enable && input.StepInformation != "" {
					c.Logger.Warningf(input.StepInformation)
				}
			}
			// get the value from user otherwise
			if input.Value == nil && enable {
				paramValue, err := utils.GetPasswordInput(input.Msg)
				if err != nil {
					return []StepInput{}, err
				}
				input.Value = paramValue
			}
			processedInputs = append(processedInputs, input)
		case confirmInput:
			// verify the input is enabled by executing the function
			enable := true
			if input.Valuesfn != nil {
				res, _ := input.Valuesfn(inputs, c)
				enable = res.(bool)
				if enable && input.StepInformation != "" {
					c.Logger.Warningf(input.StepInformation)
				}
			}
			// get the value from user otherwise
			if input.Value == nil && enable {
				input.Value = utils.GetConfirmInput(input.Msg)
			}
			processedInputs = append(processedInputs, input)
		case multiSelectionChoice:
			// process the values from the function
			var values []string = input.Values
			if input.Valuesfn != nil {
				res, err := input.Valuesfn(inputs, c)
				if err != nil {
					return []StepInput{}, err
				}
				values = res.([]string)
			}
			// get the values from user
			if input.Value == nil {
				paramValue, err := utils.GetSelectInput(input.Msg, values)
				if err != nil {
					return []StepInput{}, err
				}
				input.Value = paramValue
			}
			processedInputs = append(processedInputs, input)
		default:
			return []StepInput{}, fmt.Errorf("input not supported: %s", input.Name)
		}
	}
	return processedInputs, nil
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
			c.Logger.Actionf("Creating secret: '%s/%s'", namespace, name)
			if err := utils.CreateSecret(c.KubernetesClient, name, namespace, data); err != nil {
				return err
			}
			c.Logger.Successf("Created secret '%s/%s'", secret.Namespace, secret.Name)
		case typeFile:
			c.Logger.Actionf("Writing file to repo: '%s'", param.Name)
			file, ok := param.Value.(fileContent)
			if !ok {
				return errors.New("unexpected error casting file")
			}
			c.Logger.Actionf("Cloning flux git repo: '%s/%s'", WGEDefaultNamespace, WGEDefaultRepoName)
			pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace, c.PrivateKeyPath, c.PrivateKeyPassword)
			if err != nil {
				return fmt.Errorf("cannot clone repo: %v", err)
			}
			defer func() {
				err = utils.CleanupRepo()
				if err != nil {
					c.Logger.Failuref("failed to cleanup repo!")
				}
			}()
			c.Logger.Successf("Cloned flux git repo: '%s/%s'", WGEDefaultRepoName, WGEDefaultRepoName)

			err = utils.CreateFileToRepo(file.Name, file.Content, pathInRepo, file.CommitMsg, c.PrivateKeyPath, c.PrivateKeyPassword)
			if err != nil {
				return err
			}
			c.Logger.Successf("File '%s' is written to repo: '%s'", file.Name, WGEDefaultRepoName)

			c.Logger.Waitingf("Reconciling changes")
			if err := utils.ReconcileFlux(); err != nil {
				return err
			}
			c.Logger.Successf("Changes are reconciled successfully!")
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
