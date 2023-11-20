package steps

import (
	"fmt"
	"io"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	v1 "k8s.io/api/core/v1"
)

// BootstrapStep struct that defines the contract of a bootstrapping step.
// It is abstracted to have a generic way to handle them, so we could achieve easier
// extensibility, consistency and maintainability.
type BootstrapStep struct {
	Name  string
	Input []StepInput
	Step  func(input []StepInput, c *Config) ([]StepOutput, error)
	Stdin io.ReadCloser
}

// StepInput represents an input a step requires to execute it. for example user needs to introduce a string or a password.
type StepInput struct {
	// Name of the input to be used as id and debug logging.
	Name string
	// Msg overview message about the input.
	Msg string
	// StepInformation extended information about the input
	StepInformation string
	// Type of the input.
	Type string
	// Value is the value of the input introduced via configuration or the user.
	Value any
	// DefaultValue is the value that will be used or suggested to the user depending on the mode.
	DefaultValue any
	// IsUpdate indicates whether using this input would translate in updating a value on the system.
	IsUpdate bool
	// UpdateMsg is the message to be displayed to the user when the input is an update.
	UpdateMsg string

	// Value is the value of the input introduced via configuration or the user.
	Values []string
	// Valuesfn function to resolve potential values
	Valuesfn func(input []StepInput, c *Config) (interface{}, error)
	// Deprecated
	// Required: indicates whether the input is required or not. @deprecated
	Required bool
	// Deprecated
	// Required: indicates whether the input is required or not. @deprecated
	Enabled func(input []StepInput, c *Config) bool
}

// StepOutput represents an output generated out of the execution of a step.
// An example could be a helm release manifest for weave gitops.
type StepOutput struct {
	Name  string
	Type  string
	Value any
}

// Execute contains the business logic for executing an step.
func (s BootstrapStep) Execute(c *Config) ([]StepOutput, error) {
	inputValues, err := defaultInputStep(s.Input, c, s.Stdin)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("cannot process input '%s': %v", s.Name, err)
	}

	outputs, err := s.Step(inputValues, c)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("cannot execute '%s': %v", s.Name, err)
	}

	err = defaultOutputStep(outputs, c)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("cannot process output '%s': %v", s.Name, err)
	}
	return outputs, nil
}

// defaultInputStep default input processing
func defaultInputStep(inputs []StepInput, c *Config, stdin io.ReadCloser) ([]StepInput, error) {
	processedInputs := []StepInput{}
	for _, input := range inputs {
		// we ignore inputs that requires update but the user does not want overwrite
		if input.IsUpdate {
			if !(utils.GetConfirmInput(input.UpdateMsg, stdin) == "y") {
				continue
			}
		}

		// we ignore inputs that user has already introduced value (via flag)
		if input.Value != nil {
			continue
		}

		// we ask the user for input in any other condition
		switch input.Type {
		case stringInput:
			// verify the input is enabled by executing the function
			if input.Enabled != nil && !input.Enabled(nil, c) {
				continue
			}

			if input.StepInformation != "" {
				c.Logger.Warningf(input.StepInformation)
			}

			if input.Value == nil {
				paramValue, err := utils.GetStringInput(input.Msg, input.DefaultValue.(string), stdin)
				if err != nil {
					return []StepInput{}, err
				}
				input.Value = paramValue
			}

		case passwordInput:
			// verify the input is enabled by executing the function
			if input.Enabled != nil && !input.Enabled(inputs, c) {
				continue
			}

			if input.StepInformation != "" {
				c.Logger.Warningf(input.StepInformation)
			}

			if input.Value == nil {
				paramValue, err := utils.GetPasswordInput(input.Msg, input.Required, stdin)
				if err != nil {
					return []StepInput{}, err
				}
				input.Value = paramValue
			}
		case confirmInput:
			// verify the input is enabled by executing the function
			if input.Enabled != nil && !input.Enabled(inputs, c) {
				continue
			}

			if input.StepInformation != "" {
				c.Logger.Warningf(input.StepInformation)
			}
			// if silent mode is enabled, select yes
			if c.Silent {
				input.Value = confirmYes
			}

			// get the value from user otherwise
			if input.Value == nil {
				input.Value = utils.GetConfirmInput(input.Msg, os.Stdin)
			}
		case multiSelectionChoice:
			if input.Enabled != nil && !input.Enabled(inputs, c) {
				continue
			}
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
		default:
			return []StepInput{}, fmt.Errorf("input not supported: %s", input.Name)
		}
		processedInputs = append(processedInputs, input)
	}
	return processedInputs, nil
}

func defaultOutputStep(params []StepOutput, c *Config) error {
	for _, param := range params {
		switch param.Type {
		case typeSecret:
			secret, ok := param.Value.(v1.Secret)
			if !ok {
				panic("unexpected internal error casting secret")
			}
			name := secret.ObjectMeta.Name
			namespace := secret.ObjectMeta.Namespace
			data := secret.Data
			c.Logger.Actionf("creating secret: %s/%s", namespace, name)
			if err := utils.CreateSecret(c.KubernetesClient, name, namespace, data); err != nil {
				return err
			}
			c.Logger.Successf("created secret %s/%s", secret.Namespace, secret.Name)
		case typeFile:
			c.Logger.Actionf("write file to repo: %s", param.Name)
			file, ok := param.Value.(fileContent)
			if !ok {
				panic("unexpected internal error casting file")
			}
			c.Logger.Actionf("cloning flux git repo: %s/%s", WGEDefaultNamespace, WGEDefaultRepoName)
			pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace, c.GitScheme, c.PrivateKeyPath, c.PrivateKeyPassword, c.GitUsername, c.GitToken)
			if err != nil {
				return fmt.Errorf("cannot clone repo: %v", err)
			}
			defer func() {
				err = utils.CleanupRepo()
				if err != nil {
					c.Logger.Failuref("failed to cleanup repo!")
				}
			}()
			c.Logger.Successf("cloned flux git repo: %s/%s", WGEDefaultRepoName, WGEDefaultRepoName)

			err = utils.CreateFileToRepo(file.Name, file.Content, pathInRepo, file.CommitMsg, c.GitScheme, c.PrivateKeyPath, c.PrivateKeyPassword, c.GitUsername, c.GitToken)
			if err != nil {
				return err
			}
			c.Logger.Successf("file committed to repo: %s", file.Name)

			c.Logger.Waitingf("reconciling changes")
			if err := utils.ReconcileFlux(); err != nil {
				return err
			}
			c.Logger.Successf("changes are reconciled successfully!")
		default:
			return fmt.Errorf("unsupported param type: %s", param.Type)
		}
	}
	return nil
}
