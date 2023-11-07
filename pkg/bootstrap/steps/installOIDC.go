package steps

// NewInstallOIDCStep ask for continue installing OIDC
func NewInstallOIDCStep(config Config) BootstrapStep {
	installOIDCStep := StepInput{
		Name:         oidcInstalled,
		Type:         confirmInput,
		Msg:          oidcInstallMsg,
		DefaultValue: "",
		Valuesfn:     canAskOIDCPrompot,
	}

	return BootstrapStep{
		Name:  "Install OIDC",
		Input: []StepInput{installOIDCStep},
		Step:  setInstallOIDCFlag,
	}
}

func setInstallOIDCFlag(input []StepInput, c *Config) ([]StepOutput, error) {
	continueWithOIDC := confirmYes

	for _, param := range input {
		if param.Name == oidcInstalled {
			install, ok := param.Value.(string)
			if ok {
				continueWithOIDC = install
			}
			c.InstallOIDC = continueWithOIDC
		}
	}

	return []StepOutput{}, nil
}

func canAskOIDCPrompot(input []StepInput, c *Config) (interface{}, error) {
	return c.PromptedForDiscoveryURL, nil
}
