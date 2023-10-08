package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	externalDNSWarningMsg = `Please make sure to have the external DNS service installed in your cluster, or you have a domain that points to your cluster.
For more information about external DNS, please refer to: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html
`
	clusterDomainMsg = "Please enter your cluster domain"

	wgeInstallMsg          = "All set installing WGE v%s, This may take a few minutes"
	installSuccessMsg      = "WGE v%s is installed successfully\nYou can visit the UI at https://%s/"
	localInstallSuccessMsg = "WGE v%s is installed successfully\nYou can visit the UI at http://localhost:8000/"
)

const (
	wgeHelmRepoCommitMsg              = "Add WGE HelmRepository YAML file"
	wgeHelmReleaseCommitMsg           = "Add WGE HelmRelease YAML file"
	wgeChartName                      = "mccp"
	wgeHelmRepositoryName             = "weave-gitops-enterprise-charts"
	WGEHelmReleaseName                = "weave-gitops-enterprise"
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

var InstallWGEStep = BootstrapStep{
	Name: "install wge",
	Input: []StepInput{
		{
			Name:         UserDomain,
			Type:         stringInput,
			Msg:          clusterDomainMsg,
			DefaultValue: "",
			Valuesfn:     isUserDomainEnabled,
		},
	},
	Step:   installWge,
	Output: []StepOutput{},
}

// InstallWge installs weave gitops enterprise chart.
func installWge(input []StepInput, c *Config) ([]StepOutput, error) {
	c.UserDomain = domainTypelocalhost
	if c.DomainType == domainTypeExternalDNS {
		for _, param := range input {
			if param.Name == UserDomain {
				userDomain, ok := param.Value.(string)
				if !ok {
					return []StepOutput{}, errors.New("unexpected error occured. user domain not found")
				}
				c.UserDomain = userDomain
			}
		}
	}

	c.Logger.Waitingf(wgeInstallMsg, c.WGEVersion)

	wgehelmRepo, err := constructWgeHelmRepository()
	if err != nil {
		return []StepOutput{}, err
	}

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

	clusterController := clusterController{
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

	values := valuesFile{
		Ingress: constructIngressValues(c.UserDomain),
		TLS: map[string]interface{}{
			"enabled": false,
		},
		GitOpsSets:        gitOpsSetsValues,
		EnablePipelines:   true,
		ClusterController: clusterController,
	}

	wgeHelmRelease, err := constructWGEhelmRelease(values, c.WGEVersion)
	if err != nil {
		return []StepOutput{}, err
	}

	helmrepoFile := fileContent{
		Name:      wgeHelmrepoFileName,
		Content:   wgehelmRepo,
		CommitMsg: wgeHelmRepoCommitMsg,
	}
	helmreleaseFile := fileContent{
		Name:      wgeHelmReleaseFileName,
		Content:   wgeHelmRelease,
		CommitMsg: wgeHelmReleaseCommitMsg,
	}

	return []StepOutput{
		{
			Name:  "helmrepo file",
			Type:  typeFile,
			Value: helmrepoFile,
		},
		{
			Name:  "helmrelease file",
			Type:  typeFile,
			Value: helmreleaseFile,
		},
	}, nil
}

func constructWgeHelmRepository() (string, error) {
	wgeHelmRepo := sourcev1.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      wgeHelmRepositoryName,
			Namespace: WGEDefaultNamespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
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

func constructIngressValues(userDomain string) map[string]interface{} {
	ingressValues := map[string]interface{}{
		"annotations": map[string]string{
			"external-dns.alpha.kubernetes.io/hostname": userDomain,
		},
		"className": "public-nginx",
		"enabled":   true,
		"hosts": []map[string]interface{}{
			{
				"host": userDomain,
				"paths": []map[string]string{
					{
						"path":     "/",
						"pathType": "ImplementationSpecific",
					},
				},
			},
		},
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
			Name:      WGEHelmReleaseName,
			Namespace: WGEDefaultNamespace,
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             wgeChartName,
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
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

func isUserDomainEnabled(input []StepInput, c *Config) (interface{}, error) {
	if c.DomainType == domainTypeExternalDNS {
		c.Logger.L().Info(externalDNSWarningMsg)
		return true, nil
	}
	return false, nil
}
