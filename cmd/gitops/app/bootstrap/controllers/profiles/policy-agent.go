package profiles

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	admissionModeMsg             = "Do you want to enable admission mode"
	mutationMsg                  = "Do you want to enable mutation"
	auditMsg                     = "Do you want to enable audit mode"
	failurePolicyMsg             = "Choose your failure policy"
	policyAgentGettingStartedMsg = "Policy Agent is installed successfully, please follow the getting started guide to continue: https://docs.gitops.weave.works/enterprise/getting-started/policy-agent/"
	policyAgentInstallInfoMsg    = "Installing Policy agent ..."
	policyAgentInstallConfirmMsg = "Policy Agent is installed successfully"
)

var PolicyAgentCommand = &cobra.Command{
	Use:   "policy-agent",
	Short: "Bootstraps Weave Policy Agent",
	Example: `
# Bootstrap Weave Policy Agent
gitops bootstrap controllers policy-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallPolicyAgent()
	},
}

// InstallPolicyAgent start installing policy agent helm chart
func InstallPolicyAgent() error {
	var enableAdmission, enableAudit, enableMutate bool
	utils.Warning(policyAgentGettingStartedMsg)

	enableAdmissionResult := utils.GetConfirmInput(admissionModeMsg)

	if strings.Compare(enableAdmissionResult, "y") == 0 {
		enableAdmission = true
	} else {
		enableAdmission = false
	}

	enableMutationResult := utils.GetConfirmInput(mutationMsg)

	if strings.Compare(enableMutationResult, "y") == 0 {
		enableMutate = true
	} else {
		enableMutate = false
	}

	enableAuditResult := utils.GetConfirmInput(auditMsg)

	if strings.Compare(enableAuditResult, "y") == 0 {
		enableAudit = true
	} else {
		enableAudit = false
	}

	failurePolicies := []string{
		"Fail", "Ignore",
	}

	failurePolicyResult, err := utils.GetSelectInput(failurePolicyMsg, failurePolicies)
	if err != nil {
		return err
	}

	values := constructPolicyAgentValues(enableAdmission, enableMutate, enableAudit, failurePolicyResult)

	utils.Warning(policyAgentInstallInfoMsg)
	err = commands.UpdateHelmReleaseValues(domain.PolicyAgentValuesName, values)
	if err != nil {
		return err
	}

	utils.Info(policyAgentInstallConfirmMsg)
	return nil
}

func constructPolicyAgentValues(enableAdmission bool, enableMutate bool, enableAudit bool, failurePolicy string) map[string]interface{} {
	values := map[string]interface{}{
		"enabled": true,
		"config": map[string]interface{}{
			"admission": map[string]interface{}{
				"enabled": enableAdmission,
				"sinks": map[string]interface{}{
					"k8sEventsSink": map[string]interface{}{
						"enabled": true,
					},
				},
				"mutate": enableMutate,
			},
			"audit": map[string]interface{}{
				"enabled": enableAudit,
				"sinks": map[string]interface{}{
					"k8sEventsSink": map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
		"excludeNamespaces": []string{
			"kube-system",
		},
		"failurePolicy":  failurePolicy,
		"useCertManager": true,
	}

	return values
}
