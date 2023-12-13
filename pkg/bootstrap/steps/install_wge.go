package steps

import (
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/exp/slices"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	wgeInstallMsg   = "installing v%s ... It may take a few minutes."
	versionStepName = "select WGE version"
	wgeExistsMsg    = "Weave GitOps Enterprise is already installed in namespace %s"
	versionMsg      = "select one of the following"
)

const (
	wgeHelmRepoCommitMsg              = "Add WGE HelmRepository YAML file"
	wgeHelmReleaseCommitMsg           = "Add WGE HelmRelease YAML file"
	wgeChartName                      = "mccp"
	wgeHelmRepositoryName             = "weave-gitops-enterprise-charts"
	WgeHelmReleaseName                = "weave-gitops-enterprise"
	WGEDefaultNamespace               = "flux-system"
	WGEDefaultRepoName                = "flux-system"
	wgeHelmrepoFileName               = "wge-hrepo.yaml"
	wgeHelmReleaseFileName            = "wge-hrelease.yaml"
	wgeChartUrl                       = "https://charts.dev.wkp.weave.works/releases/charts-v3"
	clusterControllerFullOverrideName = "cluster"
	clusterControllerImageName        = "docker.io/weaveworks/cluster-controller"
	clusterControllerImageTag         = "v1.5.2"
	gitopssetsEnabledGenerators       = "GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config"
	gitopssetsBindAddress             = "127.0.0.1:8080"
	gitopssetsHealthBindAddress       = ":8081"
)

var getVersionInput = StepInput{
	Name:         inWGEVersion,
	Type:         multiSelectionChoice,
	Msg:          versionMsg,
	Valuesfn:     getWgeVersions,
	DefaultValue: "",
}

// NewCheckWGEInstallationConfig handles the WGE installation configuration
func NewCheckWGEInstallationConfig(client client.Client) (bool, error) {

	_, err := utils.GetHelmReleaseProperty(client, WgeHelmReleaseName, WGEDefaultNamespace, utils.HelmVersionProperty)
	if err != nil {
		return false, err
	}

	return true, nil
}

// NewInstallWGEStep step to install Weave GitOps Enterprise
func NewInstallWGEStep(config Config) BootstrapStep {

	// check if WGE is already installed
	if config.IsExistingWgeInstallation {
		config.Logger.Waitingf(wgeExistsMsg, WGEDefaultNamespace)
		return BootstrapStep{
			Name:  "Existing WGE installation found",
			Input: []StepInput{},
			Step:  doNothingStep,
		}
	}

	inputs := []StepInput{}

	// validate value by user
	if config.WGEVersion != "" {
		versions, err := getWgeVersions(inputs, &config)
		if err != nil {
			config.Logger.Failuref("couldn't get WGE helm chart: %v", err)
			return BootstrapStep{
				Name:  "Helm chart not found",
				Input: []StepInput{},
				Step:  doNothingStep,
			}
		}
		if versions, ok := versions.([]string); !ok || !slices.Contains(versions, config.WGEVersion) {
			config.Logger.Failuref("invalid version: %v. available versions: %s", config.WGEVersion, versions)
			return BootstrapStep{
				Name:  "invalid WGE version",
				Input: []StepInput{},
				Step:  doNothingStep,
			}
		}
	}

	if config.WGEVersion == "" {
		inputs = append(inputs, getVersionInput)
	}

	return BootstrapStep{
		Name:  "Install Weave GitOps Enterprise",
		Input: inputs,
		Step:  installWge,
	}
}

// InstallWge installs weave gitops enterprise chart.
func installWge(input []StepInput, c *Config) ([]StepOutput, error) {

	c.Logger.Actionf(wgeInstallMsg, c.WGEVersion)

	wgeHelmRepository, err := constructWgeHelmRepository()
	if err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("rendered HelmRepository file")

	gitOpsSetsValues := map[string]interface{}{
		"enabled": true,
		"controllerManager": map[string]interface{}{
			"manager": map[string]interface{}{
				"args": []string{
					fmt.Sprintf("--health-probe-bind-address=%s", gitopssetsHealthBindAddress),
					fmt.Sprintf("--metrics-bind-address=%s", gitopssetsBindAddress),
					"--leader-elect",
					fmt.Sprintf("--enabled-generators=%s", gitopssetsEnabledGenerators),
				},
			},
		},
	}

	clusterControllerValues := clusterController{
		Enabled:          true,
		FullNameOverride: clusterControllerFullOverrideName,
		ControllerManager: clusterControllerManager{
			Manager: clusterControllerManagerManager{
				Image: clusterControllerImage{
					Repository: clusterControllerImageName,
					Tag:        clusterControllerImageTag,
				},
			},
		}}

	wgeValues := valuesFile{
		Service: defaultServiceValues(),
		Ingress: defaultIngressValues(),
		TLS: map[string]interface{}{
			"enabled": false,
		},
		GitOpsSets:        gitOpsSetsValues,
		EnablePipelines:   true,
		ClusterController: clusterControllerValues,
	}

	wgeHelmRelease, err := constructWGEhelmRelease(wgeValues, c.WGEVersion)
	if err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("rendered HelmRelease file")

	helmrepoFile := fileContent{
		Name:      wgeHelmrepoFileName,
		Content:   wgeHelmRepository,
		CommitMsg: wgeHelmRepoCommitMsg,
	}
	helmreleaseFile := fileContent{
		Name:      wgeHelmReleaseFileName,
		Content:   wgeHelmRelease,
		CommitMsg: wgeHelmReleaseCommitMsg,
	}

	return []StepOutput{
		{
			Name:  wgeHelmrepoFileName,
			Type:  typeFile,
			Value: helmrepoFile,
		},
		{
			Name:  wgeHelmReleaseFileName,
			Type:  typeFile,
			Value: helmreleaseFile,
		},
	}, nil
}

func constructWgeHelmRepository() (string, error) {
	wgeHelmRepo := sourcev1beta2.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      wgeHelmRepositoryName,
			Namespace: WGEDefaultNamespace,
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			URL: wgeChartUrl,
			Interval: v1.Duration{
				Duration: time.Minute,
			},
			SecretRef: &meta.LocalObjectReference{
				Name: entitlementSecretName,
			},
		},
	}

	return utils.CreateHelmRepositoryYamlString(wgeHelmRepo)
}

func defaultServiceValues() map[string]interface{} {
	serviceValues := map[string]interface{}{
		"type": "ClusterIP",
		"port": map[string]interface{}{
			"https": 8000,
		},
		"targetPort": map[string]interface{}{
			"https": 8000,
		},
		"nodePorts": map[string]interface{}{
			"http":  "",
			"https": "",
			"tcp":   map[string]interface{}{},
			"udp":   map[string]interface{}{},
		},
		"clusterIP":                "",
		"externalIPs":              []string{},
		"loadBalancerIP":           "",
		"loadBalancerSourceRanges": []string{},
		"externalTrafficPolicy":    "",
		"healthCheckNodePort":      0,
		"annotations":              map[string]string{},
	}

	return serviceValues

}
func defaultIngressValues() map[string]interface{} {
	ingressValues := map[string]interface{}{
		"enabled":   false,
		"className": "",
		"service": map[string]interface{}{
			"name": "clusters-service",
			"port": 8000,
		},
		"annotations": map[string]string{},
		"hosts": []map[string]interface{}{
			{
				"host": "",
				"paths": []map[string]string{
					{
						"path":     "/",
						"pathType": "ImplementationSpecific",
					},
				},
			},
		},
		"tls": []map[string]string{},
	}

	return ingressValues
}

func constructWGEhelmRelease(valuesFile valuesFile, chartVersion string) (string, error) {
	valuesBytes, err := json.Marshal(valuesFile)
	if err != nil {
		return "", err
	}

	wgeHelmRelease := helmv2.HelmRelease{
		ObjectMeta: v1.ObjectMeta{
			Name:      WgeHelmReleaseName,
			Namespace: WGEDefaultNamespace,
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             wgeChartName,
					ReconcileStrategy: sourcev1beta2.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      wgeHelmRepositoryName,
						Namespace: WGEDefaultNamespace,
					},
					Version: chartVersion,
				},
			},
			Install: &helmv2.Install{
				CRDs: helmv2.CreateReplace,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: helmv2.CreateReplace,
			},
			Interval: v1.Duration{
				Duration: time.Hour,
			},
			Values: &apiextensionsv1.JSON{Raw: valuesBytes},
		},
	}

	return utils.CreateHelmReleaseYamlString(wgeHelmRelease)
}
