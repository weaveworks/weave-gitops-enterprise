package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	domainMsg             = "Please select the domain to be used"
	clusterDomainMsg      = "Please enter your cluster domain"
	externalDNSWarningMsg = `Please make sure to have the external DNS service installed in your cluster, or you have a domain that points to your cluster.
For more information about external DNS, please refer to: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html
`
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
	domainTypelocalhost               = "localhost"
	domainTypeExternalDNS             = "external DNS"
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

var (
	domainTypes = []string{
		domainTypelocalhost,
		domainTypeExternalDNS,
	}
)

// InstallWge installs weave gitops enterprise chart.
func (c *Config) InstallWge() error {
	if c.UserDomain == "" {
		c.UserDomain = domainTypelocalhost
		domainType, err := utils.GetSelectInput(domainMsg, domainTypes)
		if err != nil {
			return err
		}

		if domainType == domainTypeExternalDNS {
			c.Logger.L().Info(externalDNSWarningMsg)
			c.UserDomain, err = utils.GetStringInput(clusterDomainMsg, "")
			if err != nil {
				return err
			}
		}
	}

	c.Logger.Waitingf(wgeInstallMsg, c.WGEVersion)

	pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	defer utils.CleanupRepo()

	wgehelmRepo, err := constructWgeHelmRepository()
	if err != nil {
		return err
	}

	err = utils.CreateFileToRepo(wgeHelmrepoFileName, wgehelmRepo, pathInRepo, wgeHelmRepoCommitMsg)
	if err != nil {
		return err
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
		return err
	}

	if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, wgeHelmRelease, pathInRepo, wgeHelmReleaseCommitMsg); err != nil {
		return err
	}

	if err := utils.ReconcileFlux(); err != nil {
		return err
	}

	if err := utils.ReconcileHelmRelease(WGEHelmReleaseName); err != nil {
		return err
	}

	return nil
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

// CheckUIDomain display the message to be for external dns or localhost.
func (c *Config) CheckUIDomain() error {
	if !strings.Contains(c.UserDomain, domainTypelocalhost) {
		c.Logger.Successf(installSuccessMsg, c.WGEVersion, c.UserDomain)
		return nil
	}

	c.Logger.Successf(localInstallSuccessMsg, c.WGEVersion)

	var runner runner.CLIRunner
	_, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		// adding an error message, err is meaningless
		return errors.New("failed to make portforward 8000")
	}

	return nil
}
