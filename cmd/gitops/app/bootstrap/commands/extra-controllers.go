package commands

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

func CheckExtraControllers(version string) {

	var extraControllers []string = []string{
		"None",
		"policy-agent",
		"pipeline-controller",
		"gitopssets-controller",
	}

	extraControllerPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Do you want another controller to be installed on your cluster",
		DefaultValue: "",
	}

	controllerName := utils.GetPromptSelect(extraControllerPrompt, extraControllers)
	if strings.Compare(controllerName, "None") == 0 {
		return
	}

	switch controllerName {
	case "policy-agent":
		profiles.BootstrapPolicyAgent()
		return
	case "pipeline-controller":
		fmt.Println("not implemented yet!")
	case "gitopssets-controller":
		fmt.Println("not implemented yet!")
	}
}
