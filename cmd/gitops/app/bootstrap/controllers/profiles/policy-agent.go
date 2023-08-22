package profiles

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	POLICY_AGENT_HELMREPO_NAME         = "policy-agent"
	POLICY_AGENT_CHART_URL             = "https://weaveworks.github.io/policy-agent/"
	POLICY_AGENT_HELMRELEASE_NAME      = "policy-agent"
	POLICY_AGENT_VALUES_FILES_LOCATION = "/tmp/agent-values.yaml"
	TARGET_NAMESPACE                   = "flux-system"
	AGENT_VERSION                      = "2.5.0"
)

var FailurePolicies []string = []string{
	"Fail", "Ignore",
}

var PolicyAgentCommand = &cobra.Command{
	Use:   "policy-agent",
	Short: "Bootstraps Weave Policy Agent",
	Example: `
# Bootstrap Weave Policy Agent
gitops bootstrap controllers policy-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return BootstrapPolicyAgent()
	},
}

func BootstrapPolicyAgent() error {

	fmt.Println("For more information about the configurations please refer to the docs https://github.com/weaveworks/policy-agent/blob/dev/docs/README.md")

	admissionPrompt := promptui.Prompt{
		Label:     "Do you want to enable admission mode",
		IsConfirm: true,
	}

	enableAdmissionResult, _ := admissionPrompt.Run()
	if strings.Compare(enableAdmissionResult, "y") == 0 {
		enableAdmissionResult = "true"
	} else {
		enableAdmissionResult = "false"
	}

	mutationPrompt := promptui.Prompt{
		Label:     "Do you want to enable mutation",
		IsConfirm: true,
	}

	enableMutationResult, _ := mutationPrompt.Run()
	if strings.Compare(enableMutationResult, "y") == 0 {
		enableMutationResult = "true"
	} else {
		enableMutationResult = "false"
	}

	auditPrompt := promptui.Prompt{
		Label:     "Do you want to enable audit mode",
		IsConfirm: true,
	}

	enableAuditResult, _ := auditPrompt.Run()
	if strings.Compare(enableAuditResult, "y") == 0 {
		enableAuditResult = "true"
	} else {
		enableAuditResult = "false"
	}

	FailurePolicyPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Choose your failure policy",
		DefaultValue: "",
	}

	failurePolicyResult := utils.GetPromptSelect(FailurePolicyPrompt, FailurePolicies)

	fmt.Println("Installing policy agent ...")

	var runner runner.CLIRunner

	out, err := runner.Run("flux", "create", "source", "helm", POLICY_AGENT_HELMREPO_NAME, "--url", POLICY_AGENT_CHART_URL)
	if err != nil {
		fmt.Printf("An error occurred creating helmrepository\n%v\n", string(out))
		os.Exit(1)
	}

	values := fmt.Sprintf(`caCertificate: ""
certificate: ""
config:
  accountId: ""
  admission:
    enabled: %s
    sinks:
      k8sEventsSink:
        enabled: true
    mutate: %s
  audit:
    enabled: %s
    sinks:
      k8sEventsSink:
        enabled: true
  clusterId: ""
excludeNamespaces:
- kube-system
failurePolicy: %s
image: weaveworks/policy-agent
key: ""
persistence:
  enabled: false
useCertManager: true
`, enableAdmissionResult, enableMutationResult, enableAuditResult, failurePolicyResult)

	valuesFile, err := os.Create(POLICY_AGENT_VALUES_FILES_LOCATION)
	utils.CheckIfError(err)

	defer valuesFile.Close()
	_, err = valuesFile.WriteString(values)
	utils.CheckIfError(err)

	err = valuesFile.Sync()
	utils.CheckIfError(err)

	out, err = runner.Run("flux", "create", "hr", POLICY_AGENT_HELMRELEASE_NAME,
		"--source", fmt.Sprintf("HelmRepository/%s", POLICY_AGENT_HELMREPO_NAME),
		"--chart", POLICY_AGENT_HELMRELEASE_NAME,
		"--chart-version", AGENT_VERSION,
		"--interval", "10m0s",
		"--crds", "CreateReplace",
		"--values", POLICY_AGENT_VALUES_FILES_LOCATION,
		"--target-namespace", TARGET_NAMESPACE,
	)
	if err != nil {
		fmt.Printf("An error occurred installing policy agent\n%v", string(out))
	}
	fmt.Printf("✔ Policy Agent is installed successfully")
	return nil
}
