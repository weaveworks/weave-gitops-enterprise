package profiles

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	POLICY_AGENT_HELMREPO_NAME         = "policy-agent"
	POLICY_AGENT_CHART_URL             = "https://weaveworks.github.io/policy-agent/"
	POLICY_AGENT_HELMRELEASE_NAME      = "policy-agent"
	AGENT_VERSION                      = "2.5.0"
	DEFAULT_NAMESPACE                  = "flux-system"
	POLICY_AGENT_TARGET_NAMESPACE      = "policy-system"
	POLICY_AGENT_HELMREPO_FILENAME     = "policy-agent-helmrepo.yaml"
	POLICY_AGENT_HELMRELEASE_FILENAME  = "policy-agent-helmrelease.yaml"
	POLICY_AGENT_HELMREPO_COMMITMSG    = "Add Policy Agent HelmRepository YAML file"
	POLICY_AGENT_HELMRELEASE_COMMITMSG = "Add Policy Agent HelmRelease YAML file"
	ADMISSION_MODE_MSG                 = "Do you want to enable admission mode"
	MUTATION_MSG                       = "Do you want to enable mutation"
	AUDIT_MSG                          = "Do you want to enable audit mode"
	FAILURE_POLICY_MSG                 = "Choose your failure policy"
)

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

// BootstrapPolicyAgent start installing policy agent helm chart
func BootstrapPolicyAgent() error {
	var enableAdmission, enableAudit, enableMutate bool
	utils.Warning("For more information about the configurations please refer to the docs https://github.com/weaveworks/policy-agent/blob/dev/docs/README.md")

	enableAdmissionResult, err := utils.GetConfirmInput(ADMISSION_MODE_MSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if strings.Compare(enableAdmissionResult, "y") == 0 {
		enableAdmission = true
	} else {
		enableAdmission = false
	}

	enableMutationResult, err := utils.GetConfirmInput(MUTATION_MSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if strings.Compare(enableMutationResult, "y") == 0 {
		enableMutate = true
	} else {
		enableMutate = false
	}

	enableAuditResult, err := utils.GetConfirmInput(AUDIT_MSG)
	if err != nil {
		return utils.CheckIfError(err)
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
		return utils.CheckIfError(err)
	}

	utils.Warning("Installing policy agent ...")

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

	policyAgentHelmRepo, err := constructPolicyAgentHelmRepository()
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.CreateFileToRepo(POLICY_AGENT_HELMREPO_FILENAME, policyAgentHelmRepo, pathInRepo, POLICY_AGENT_HELMREPO_COMMITMSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	policyAgentHelmRelease, err := constructPolicyAgentHelmRelease(enableAdmission, enableMutate, enableAudit, failurePolicyResult)
	if err != nil {
		return utils.CheckIfError(err)
	}
	err = utils.CreateFileToRepo(POLICY_AGENT_HELMRELEASE_FILENAME, policyAgentHelmRelease, pathInRepo, POLICY_AGENT_HELMRELEASE_FILENAME)
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.ReconcileFlux()
	if err != nil {
		return utils.CheckIfError(err)
	}

	utils.Info("✔ Policy Agent is installed successfully")
	return nil
}

func constructPolicyAgentHelmRepository() (string, error) {
	agentHelmRepo := sourcev1.HelmRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: sourcev1.GroupVersion.Identifier(),
			Kind:       sourcev1.HelmRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              POLICY_AGENT_HELMRELEASE_NAME,
			Namespace:         DEFAULT_NAMESPACE,
			CreationTimestamp: v1.Now(),
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: POLICY_AGENT_CHART_URL,
			Interval: v1.Duration{
				Duration: time.Minute,
			},
		},
	}

	agentHelmRepoBytes, err := k8syaml.Marshal(agentHelmRepo)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	return string(agentHelmRepoBytes), nil
}

func constructPolicyAgentHelmRelease(enableAdmission bool, enableMutate bool, enableAudit bool, failurePolicy string) (string, error) {
	values := constructPolicyAgentValues(enableAdmission, enableMutate, enableAudit, failurePolicy)

	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	policyAgentHelmRelease := helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              POLICY_AGENT_HELMRELEASE_NAME,
			Namespace:         DEFAULT_NAMESPACE,
			CreationTimestamp: v1.Now(),
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: POLICY_AGENT_HELMREPO_NAME,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:       sourcev1.HelmRepositoryKind,
						Name:       POLICY_AGENT_HELMREPO_NAME,
						Namespace:  DEFAULT_NAMESPACE,
						APIVersion: sourcev1.GroupVersion.Identifier(),
					},
					Version: AGENT_VERSION,
				},
			},
			Install: &helmv2.Install{
				CRDs:            helmv2.CreateReplace,
				CreateNamespace: true,
			},
			Interval: v1.Duration{
				Duration: time.Minute * 10,
			},
			TargetNamespace: POLICY_AGENT_TARGET_NAMESPACE,
			Values:          &apiextensionsv1.JSON{Raw: valuesBytes},
		},
	}

	policyAgentHelmReleaseBytes, err := k8syaml.Marshal(policyAgentHelmRelease)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	return string(policyAgentHelmReleaseBytes), nil
}

func constructPolicyAgentValues(enableAdmission bool, enableMutate bool, enableAudit bool, failurePolicy string) map[string]interface{} {
	values := map[string]interface{}{
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
