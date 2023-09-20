package profiles

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	admissionModeMsg             = "Do you want to enable admission mode"
	mutationMsg                  = "Do you want to enable mutation"
	auditMsg                     = "Do you want to enable audit mode"
	failurePolicyMsg             = "Choose your failure policy"
	policyAgentGettingStartedMsg = "please follow the getting started guide to continue: https://docs.gitops.weave.works/enterprise/getting-started/policy-agent/"
	policyAgentInstallInfoMsg    = "Installing Policy agent ..."
	policyAgentInstallConfirmMsg = "Policy Agent is installed successfully"
)

func PolicyAgentCommand(opts *config.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "policy-agent",
		Short: "Add Weave Policy Agent",
		Example: `
	# Add Weave Policy Agent
	gitops add controllers policy-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return InstallPolicyAgent(opts)
		},
	}
}

// InstallPolicyAgent start installing policy agent helm chart
func InstallPolicyAgent(opts *config.Options) error {
	var enableAdmission, enableAudit, enableMutate bool
	utils.Warning(policyAgentGettingStartedMsg)

	enableAdmissionResult := utils.GetConfirmInput(admissionModeMsg)

	if enableAdmissionResult == "y" {
		enableAdmission = true
	} else {
		enableAdmission = false
	}

	enableMutationResult := utils.GetConfirmInput(mutationMsg)

	if enableMutationResult == "y" {
		enableMutate = true
	} else {
		enableMutate = false
	}

	enableAuditResult := utils.GetConfirmInput(auditMsg)

	if enableAuditResult == "y" {
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

	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return err
	}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	err = commands.UpdateHelmReleaseValues(cl, domain.PolicyAgentValuesName, values)
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
