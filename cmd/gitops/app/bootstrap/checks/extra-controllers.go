package checks

import (
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/runner"
	"golang.org/x/exp/slices"
)

func CheckExtraControllers(version string, extraControllers []string) {

	extraControllerPrompt := promptContent{
		"",
		"Do you want another controller to be installed on your cluster",
		"",
	}

	controllerName := promptGetSelect(extraControllerPrompt, extraControllers)
	if strings.Compare(controllerName, "None") == 0 {
		return
	}

	valuesFile, err := os.OpenFile(VALUES_FILES_LOCATION, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer valuesFile.Close()
	var values string

	switch controllerName {
	case "policy-agent":
		values = `policy-agent:
  enabled: true
  config:
    admission:
      enabled: true
      sinks:
        k8sEventsSink:
          enabled: true
    audit:
      enabled: true
      sinks:
        k8sEventsSink:
          enabled: true
    excludeNamespaces:
    - kube-system
      flux-system
    accountId: ""
    clusterId: ""
`
	case "pipeline-controller":
		values = "enablePipelines: true"
	case "gitopssets-controller":
		values = `gitopssets-controller:
  enabled: true
  controllerManager:
    manager:
      args:
        - --health-probe-bind-address=:8081
        - --metrics-bind-address=127.0.0.1:8080
        - --leader-elect
        # enable the cluster generator which is not enabled by default
        - --enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config
`
	}

	if _, err = valuesFile.WriteString(values); err != nil {
		panic(err)
	}

	var runner runner.CLIRunner
	fmt.Printf("\nInstalling controller %s on your cluster ...\n", controllerName)
	out, err := runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
		"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
		"--chart", "mccp",
		"--chart-version", version,
		"--interval", "65m",
		"--crds", "CreateReplace",
		"--values", VALUES_FILES_LOCATION,
	)
	if err != nil {
		fmt.Printf("An error occurred updating helmrelease\n%v\n", string(out))
		os.Exit(1)
	}

	fmt.Printf("\n✔  controller %s is installed on your cluster\n", controllerName)
	extraControllers = slices.Delete(extraControllers, slices.Index(extraControllers, controllerName), slices.Index(extraControllers, controllerName)+1)
	CheckExtraControllers(version, extraControllers)
}
