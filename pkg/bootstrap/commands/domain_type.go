package commands

import "errors"

const (
	domainMsg = "Please select the domain to be used"
)
const (
	domainTypelocalhost   = "localhost"
	domainTypeExternalDNS = "external DNS"
)

var (
	domainTypes = []string{
		domainTypelocalhost,
		domainTypeExternalDNS,
	}
)

var SelectDomainType = BootstrapStep{
	Name: "select domain type",
	Input: []StepInput{
		{
			Name:         "domainType",
			Type:         multiSelectionChoice,
			Msg:          domainMsg,
			Values:       domainTypes,
			DefaultValue: "",
		},
	},
	Step: selectDomainType,
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
	return []StepOutput{}, nil
}
