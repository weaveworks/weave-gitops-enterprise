package steps

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	v1 "k8s.io/api/core/v1"
	k8syaml "sigs.k8s.io/yaml"
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
	// SupportUpdate indicates whether the input supports being updated or not.
	SupportUpdate bool
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

func (o StepOutput) Export(writer io.Writer) error {
	switch o.Type {
	case typeSecret:
		secret, ok := o.Value.(v1.Secret)
		if !ok {
			return fmt.Errorf("unexpected internal error casting secret")
		}
		err := printResource(writer, secret)
		if err != nil {
			return fmt.Errorf("error printing resource: %v", err)
		}
	case typeFile:
		file, ok := o.Value.(fileContent)
		if !ok {
			panic("unexpected internal error casting file")
		}
		err := printByteArray(writer, []byte(file.Content))
		if err != nil {
			return fmt.Errorf("error printing string: %v", err)
		}
	default:
		return fmt.Errorf("unsupported param type: %s", o.Type)
	}
	return nil
}

func printResource(writer io.Writer, resource interface{}) error {
	resourceAsBytes, err := k8syaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error marshalling resource: %v", err)
	}
	err2 := printByteArray(writer, resourceAsBytes)
	if err2 != nil {
		return err2
	}

	return nil
}

func printByteArray(writer io.Writer, resourceAsBytes []byte) error {
	_, err := fmt.Fprintln(writer, "---")
	if err != nil {
		return fmt.Errorf("error printing resource: %v", err)
	}

	_, err = fmt.Fprintln(writer, resourceToString(resourceAsBytes))
	if err != nil {
		return fmt.Errorf("error printing resource: %v", err)
	}
	return nil
}

func resourceToString(data []byte) string {
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)
	data = bytes.TrimSpace(data)
	return string(data)
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
		// process updates
		if input.IsUpdate {
			if !input.SupportUpdate {
				// scenario a - dont support update. we show the message saying that it will use existing value.
				c.Logger.Warningf(input.UpdateMsg)
				continue
			} else if !(utils.GetConfirmInput(input.UpdateMsg, stdin) == "y") {
				// scenario b - support update but user dont want to udpate, we just leave.
				continue
			}
			// scenario c - user wants update so we ask for input
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
			// if silent mode is enabled, select the default value
			// if no default value an error will be returned
			if c.ModesConfig.Silent {
				defaultVal, ok := input.DefaultValue.(string)
				if ok {
					input.Value = defaultVal
				} else {
					return []StepInput{}, fmt.Errorf("invalid default value: %v", input.DefaultValue)
				}
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

	// if export we dont process at the level of the step but at the end of the workflow
	if c.ModesConfig.Export {
		return nil
	}

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
			pathInRepo, err := c.GitClient.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace, c.GitRepository.Scheme, c.PrivateKeyPath, c.PrivateKeyPassword, c.GitUsername, c.GitToken)
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

			err = c.GitClient.CreateFileToRepo(file.Name, file.Content, pathInRepo, file.CommitMsg, c.GitRepository.Scheme, c.PrivateKeyPath, c.PrivateKeyPassword, c.GitUsername, c.GitToken)
			if err != nil {
				return err
			}
			c.Logger.Successf("file committed to repo: %s", file.Name)

			c.Logger.Waitingf("reconciling changes")
			if err := c.FluxClient.ReconcileFlux(); err != nil {
				return err
			}
			c.Logger.Successf("changes are reconciled successfully!")
		default:
			return fmt.Errorf("unsupported param type: %s", param.Type)
		}
	}
	return nil
}

// doNothingStep is a step without logic to be used for not required steps
func doNothingStep(input []StepInput, c *Config) ([]StepOutput, error) {
	return []StepOutput{}, nil
}
