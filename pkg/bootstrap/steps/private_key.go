package steps

import (
	"fmt"
	"os"
)

const (
	privateKeyMsg         = "Private key path and password\nDisclaimer: private key will be used to push WGE resources into the default repository only. It won't be stored or used anywhere else for any reason."
	privateKeyPathMsg     = "Private key path"
	privateKeyPasswordMsg = "Private key password"
)

var (
	privateKeyDefaultPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
)

var getKeyPath = StepInput{
	Name:         PrivateKeyPath,
	Type:         stringInput,
	Msg:          privateKeyPathMsg,
	DefaultValue: privateKeyDefaultPath,
}

var getKeyPassword = StepInput{
	Name:         PrivateKeyPassword,
	Type:         passwordInput,
	Msg:          privateKeyPasswordMsg,
	DefaultValue: "",
}

func NewAskPrivateKeyStep(config Config) BootstrapStep {
	inputs := []StepInput{}

	if config.PrivateKeyPath == "" {
		inputs = append(inputs, getKeyPath)
		inputs = append(inputs, getKeyPassword)
	}
	return BootstrapStep{
		Name:  privateKeyMsg,
		Input: inputs,
		Step:  configurePrivateKey,
	}
}

func configurePrivateKey(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == PrivateKeyPath {
			privateKeyPath, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPath = privateKeyPath
			}
		}
		if param.Name == PrivateKeyPassword {
			privateKeyPassword, ok := param.Value.(string)
			if ok {
				c.PrivateKeyPassword = privateKeyPassword
			}
		}
	}
	return []StepOutput{}, nil
}
