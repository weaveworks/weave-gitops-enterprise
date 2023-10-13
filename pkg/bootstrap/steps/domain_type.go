package steps

import (
	"errors"
)

const (
	domainStepName = "dashboard access"
	domainMsg      = "select dashboard access domain type"
)
const (
	domainTypeLocalhost   = "localhost"
	domainTypeExternalDNS = "externalDNS"
)

var (
	domainTypes = []string{
		domainTypeLocalhost,
		domainTypeExternalDNS,
	}
)

var getDomainType = StepInput{
	Name:         domainType,
	Type:         multiSelectionChoice,
	Msg:          domainMsg,
	Values:       domainTypes,
	DefaultValue: "",
}

func NewSelectDomainType(config Config) BootstrapStep {
	inputs := []StepInput{}

	switch config.DomainType {
	case domainTypeLocalhost:
		break
	case domainTypeExternalDNS:
		break
	default:
		inputs = append(inputs, getDomainType)
	}

	return BootstrapStep{
		Name:  domainStepName,
		Input: inputs,
		Step:  selectDomainType,
	}
}

func selectDomainType(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == domainType {
			domainType, ok := param.Value.(string)
			if !ok {
				return []StepOutput{}, errors.New("unexpected error occurred. domainType is not found")
			}
			c.DomainType = domainType
		}
	}
	if c.DomainType == "" {
		return []StepOutput{}, errors.New("unexpected error occurred. domainType is not found")
	}
	c.Logger.Successf("dashboard access domain: %s", c.DomainType)
	return []StepOutput{}, nil
}
