package steps

import (
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionMsg                 = "do you want to enable the Agent's admission controller"
	policyAgentInstallInfoMsg    = "installing Policy Agent ..."
	policyAgentInstallConfirmMsg = "Policy Agent is installed successfully"
)

const (
	agentChartURL             = "https://weaveworks.github.io/policy-agent/"
	agentHelmRepoName         = "policy-agent"
	agentHelmReleaseName      = "policy-agent"
	agentNamespace            = "policy-system"
	agentHelmRepoFileName     = "policy-agent-helmrepo.yaml"
	agentHelmReleaseFileName  = "policy-agent-helmrelease.yaml"
	agentHelmRepoCommitMsg    = "Add Policy Agent HelmRepository YAML file"
	agentHelmReleaseCommitMsg = "Add Policy Agent HelmRelease YAML file"
	agentVersion              = "2.5.0"
)

var enableAdmission = StepInput{
	Name:         inEnableAdmission,
	Type:         confirmInput,
	Msg:          admissionMsg,
	DefaultValue: confirmNo,
}

// NewInstallPolicyAgentStep ask for continue installing OIDC
func NewInstallPolicyAgentStep(config Config) BootstrapStep {
	config.Logger.Warningf("please note that the Policy Agent requires cert-manager to be installed!")
	return BootstrapStep{
		Name:  "install Policy Agent",
		Input: []StepInput{enableAdmission},
		Step:  installPolicyAgent,
	}
}

// installPolicyAgent start installing policy agent helm chart
func installPolicyAgent(input []StepInput, c *Config) ([]StepOutput, error) {
	enableAdmission := false

	for _, param := range input {
		if param.Name == inEnableAdmission {
			enable, ok := param.Value.(string)
			if ok && enable == confirmYes {
				enableAdmission = true
			}
		}
	}
	c.Logger.Actionf(policyAgentInstallInfoMsg)
	c.Logger.Actionf("rendering Policy Agent HelmRepository file")
	agentHelmRepoObject := sourcev1beta2.HelmRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: sourcev1beta2.GroupVersion.Identifier(),
			Kind:       sourcev1beta2.HelmRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      agentHelmRepoName,
			Namespace: WGEDefaultNamespace,
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			URL: agentChartURL,
			Interval: v1.Duration{
				Duration: time.Minute,
			},
		},
	}
	agentHelmRepoFile, err := utils.CreateHelmRepositoryYamlString(agentHelmRepoObject)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to render Policy Agent HelmRepository: %v", err)
	}
	c.Logger.Actionf("rendered Policy Agent HelmRepository file")

	c.Logger.Actionf("rendering Policy Agent HelmRelease file")

	values := map[string]interface{}{
		"config": map[string]interface{}{
			"admission": map[string]interface{}{
				"enabled": enableAdmission,
				"sinks": map[string]interface{}{
					"k8sEventsSink": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			"audit": map[string]interface{}{
				"enabled": true,
				"sinks": map[string]interface{}{
					"k8sEventsSink": map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
		"excludeNamespaces": []string{
			"kube-system",
			"flux-system",
		},
		"failurePolicy":  "Fail",
		"useCertManager": true,
	}

	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to render Policy Agent HelmRepository values: %v", err)
	}

	agentHelmReleaseObject := helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      agentHelmReleaseName,
			Namespace: WGEDefaultNamespace,
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: agentHelmRepoName,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:       sourcev1beta2.HelmRepositoryKind,
						Name:       agentHelmRepoName,
						Namespace:  WGEDefaultNamespace,
						APIVersion: sourcev1beta2.GroupVersion.Identifier(),
					},
					Version: agentVersion,
				},
			},
			Install: &helmv2.Install{
				CRDs:            helmv2.CreateReplace,
				CreateNamespace: true,
			},
			Interval: v1.Duration{
				Duration: time.Minute * 10,
			},
			TargetNamespace: agentNamespace,
			Values:          &apiextensionsv1.JSON{Raw: valuesBytes},
		},
	}

	agentHelmReleaseFile, err := utils.CreateHelmReleaseYamlString(agentHelmReleaseObject)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to render Policy Agent HelmRelease: %v", err)
	}
	c.Logger.Actionf("rendered Policy Agent HelmRelease file")

	helmrepoFile := fileContent{
		Name:      agentHelmRepoFileName,
		Content:   agentHelmRepoFile,
		CommitMsg: agentHelmRepoCommitMsg,
	}
	helmreleaseFile := fileContent{
		Name:      agentHelmReleaseFileName,
		Content:   agentHelmReleaseFile,
		CommitMsg: agentHelmReleaseCommitMsg,
	}

	c.Logger.Successf(policyAgentInstallConfirmMsg)
	return []StepOutput{
		{
			Name:  agentHelmRepoFileName,
			Type:  typeFile,
			Value: helmrepoFile,
		},
		{
			Name:  agentHelmReleaseFileName,
			Type:  typeFile,
			Value: helmreleaseFile,
		},
	}, nil
}
