package profiles

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	ADMISSION_MODE_MSG = "Do you want to enable admission mode"
	MUTATION_MSG       = "Do you want to enable mutation"
	AUDIT_MSG          = "Do you want to enable audit mode"
	FAILURE_POLICY_MSG = "Choose your failure policy"
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
	utils.Warning("For more information about the configurations please refer to the docs https://github.com/weaveworks/policy-agent/blob/dev/docs/README.md")

	enableAdmissionResult, err := utils.GetConfirmInput(ADMISSION_MODE_MSG)
	if err != nil {
		return err
	}

	if strings.Compare(enableAdmissionResult, "y") == 0 {
		enableAdmission = true
	} else {
		enableAdmission = false
	}

	enableMutationResult, err := utils.GetConfirmInput(MUTATION_MSG)
	if err != nil {
		return err
	}

	if strings.Compare(enableMutationResult, "y") == 0 {
		enableMutate = true
	} else {
		enableMutate = false
	}

	enableAuditResult, err := utils.GetConfirmInput(AUDIT_MSG)
	if err != nil {
		return err
	}

	if strings.Compare(enableAuditResult, "y") == 0 {
		enableAudit = true
	} else {
		enableAudit = false
	}

	failurePolicies := []string{
		"Fail", "Ignore",
	}

	failurePolicyResult, err := utils.GetSelectInput(FAILURE_POLICY_MSG, failurePolicies)
	if err != nil {
		return err
	}

	values := constructPolicyAgentValues(enableAdmission, enableMutate, enableAudit, failurePolicyResult)

	utils.Warning("Installing policy agent ...")
	err = commands.InstallController(domain.POLICY_AGENT_VALUES_NAME, values)
	if err != nil {
		return err
	}

	utils.Info("Policy Agent is installed successfully")
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
