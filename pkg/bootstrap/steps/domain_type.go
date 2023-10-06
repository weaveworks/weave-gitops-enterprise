package steps

import (
	"errors"
)

const (
	domainStepName = "Dashboard access"
	domainMsg      = "Please select the domain to be used"
)
const (
	domainTypeLocalhost   = "localhost"
	domainTypeExternalDNS = "external DNS"
)

var (
	domainTypes = []string{
		domainTypeLocalhost,
		domainTypeExternalDNS,
	}
)

var getDomainType = StepInput{
	Name:         "domainType",
	Type:         multiSelectionChoice,
	Msg:          domainMsg,
	Values:       domainTypes,
	DefaultValue: "",
}

func NewSelectDomainType(config Config) BootstrapStep {
	inputs := []StepInput{}

	if config.DomainType == "" {
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
		if param.Name == "domainType" {
			domainType, ok := param.Value.(string)
			if !ok {
				return []StepOutput{}, errors.New("unexpected error occured. domain type not found")
			}
			c.DomainType = domainType
		}
	}
	c.Logger.Successf("dashboard access domain: %s", c.DomainType)
	return []StepOutput{}, nil
}
