package commands

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	EXTRA_CONTROLLERS_MSG = "Do you want another controller to be installed on your cluster"
)

// CheckExtraControllers asks user to install extra controllers
func CheckExtraControllers(version string) error {
	var extraControllers []string = []string{
		"None",
		"policy-agent",
		"pipeline-controller",
		"gitopssets-controller",
	}

	controllerName, err := utils.GetSelectInput(EXTRA_CONTROLLERS_MSG, extraControllers)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if strings.Compare(controllerName, "None") == 0 {
		return nil
	}

	switch controllerName {
	case "policy-agent":
		return profiles.BootstrapPolicyAgent()
	case "pipeline-controller":
		fmt.Println("not implemented yet!")
	case "gitopssets-controller":
		fmt.Println("not implemented yet!")
	}

	return nil
}
