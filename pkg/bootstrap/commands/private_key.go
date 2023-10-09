package commands

import (
	"fmt"
	"os"
)

const (
	privateKeyPathMsg     = "Private key path"
	privateKeyPasswordMsg = "Private key password"
)

var (
	privateKeyDefaultPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
)

var AskPrivateKeyStep = BootstrapStep{
	Name: "ask private key path and password",
	Input: []StepInput{
		{
			Name:         PrivateKeyPath,
			Type:         stringInput,
			Msg:          privateKeyPathMsg,
			DefaultValue: privateKeyDefaultPath,
		},
		{
			Name:         PrivateKeyPassword,
			Type:         passwordInput,
			Msg:          privateKeyPasswordMsg,
			DefaultValue: "",
		},
	},
	Step: configurePrivateKey,
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
