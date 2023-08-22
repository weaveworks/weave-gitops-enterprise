package profiles

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	POLICY_AGENT_HELMREPO_NAME    = "policy-agent"
	POLICY_AGENT_CHART_URL        = "https://weaveworks.github.io/policy-agent/"
	POLICY_AGENT_HELMRELEASE_NAME = "policy-agent"
	AGENT_VERSION                 = "2.5.0"
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

	enableAdmissionResult, err := admissionPrompt.Run()
	if err != nil {
		return utils.CheckIfError(err)
	}
	if strings.Compare(enableAdmissionResult, "y") == 0 {
		enableAdmissionResult = "true"
	} else {
		enableAdmissionResult = "false"
	}

	mutationPrompt := promptui.Prompt{
		Label:     "Do you want to enable mutation",
		IsConfirm: true,
	}

	enableMutationResult, err := mutationPrompt.Run()
	if err != nil {
		return utils.CheckIfError(err)
	}
	if strings.Compare(enableMutationResult, "y") == 0 {
		enableMutationResult = "true"
	} else {
		enableMutationResult = "false"
	}

	auditPrompt := promptui.Prompt{
		Label:     "Do you want to enable audit mode",
		IsConfirm: true,
	}

	enableAuditResult, err := auditPrompt.Run()
	if err != nil {
		return utils.CheckIfError(err)
	}
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

	failurePolicyResult, err := utils.GetPromptSelect(FailurePolicyPrompt, FailurePolicies)
	if err != nil {
		return utils.CheckIfError(err)
	}

	fmt.Println("Installing policy agent ...")

	pathInRepo, err := utils.CloneRepo()
	if err != nil {
		return utils.CheckIfError(err)
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			fmt.Println("cleanup failed!")
		}
	}()

	policyAgentHelmRepo := fmt.Sprintf(`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: %s
  namespace: flux-system
spec:
  interval: 1m0s
  url: %s
`, POLICY_AGENT_HELMREPO_NAME, POLICY_AGENT_CHART_URL)

	err = utils.CreateFileToRepo("policy-agent-helmrepo.yaml", policyAgentHelmRepo, pathInRepo, "create policy agent helmrepository")
	if err != nil {
		return utils.CheckIfError(err)
	}

	policyAgentHelmRelease := fmt.Sprintf(`apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: %s
  namespace: flux-system
spec:
  chart:
    spec:
      chart: %s
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: %s
        namespace: flux-system
      version: %s
  interval: 10m0s
  targetNamespace: policy-system
  install:
    createNamespace: true
  values:
    caCertificate: ""
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
`, POLICY_AGENT_HELMRELEASE_NAME, POLICY_AGENT_HELMRELEASE_NAME, POLICY_AGENT_HELMREPO_NAME, AGENT_VERSION, enableAdmissionResult, enableMutationResult, enableAuditResult, failurePolicyResult)

	err = utils.CreateFileToRepo("policy-agent-helmrelease.yaml", policyAgentHelmRelease, pathInRepo, "create policy agent helmrelease")
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.ReconcileFlux()
	if err != nil {
		return utils.CheckIfError(err)
	}

	fmt.Println("✔ Policy Agent is installed successfully")
	return nil
}
