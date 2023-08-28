package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DomainMsg               = "Please select the domain to be used"
	ClusterDomainMsg        = "Please enter your cluster domain"
	WGEHelmRepoCommitMsg    = "Add WGE HelmRepository YAML file"
	WGEHelmReleaseCommitMsg = "Add WGE HelmRelease YAML file"
	ExternalDNSWarningMsg   = `
	Please make sure to have the external DNS service installed in your cluster, 
	or you have a domain that points to your cluster.
	For more information about external DNS, please refer to:
	https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html
	`
	WGEInstallMsgFormat          = "All set installing WGE v%s, This may take a few minutes...\n"
	InstallSuccessMsgFormat      = "\nWGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n"
	LocalInstallSuccessMsgFormat = "\nWGE v%s is installed successfully\n\n✅ You can visit the UI at http://localhost:8000/\n"
)
const (
	WGEChartName                      = "mccp"
	WGEHelmRepositoryName             = "weave-gitops-enterprise-charts"
	WGEHelmReleaseName                = "weave-gitops-enterprise"
	WGEDefaultNamespace               = "flux-system"
	DomainTypelocalgost               = "localhost (Using Portforward)"
	DomainTypeExternalDNS             = "external DNS"
	WGEHelmrepoFileName               = "wge-hrepo.yaml"
	WGEHelmReleaseFileName            = "wge-hrelease.yaml"
	WGEChartUrl                       = "https://charts.dev.wkp.weave.works/releases/charts-v3"
	ClusterControllerFullOverrideName = "cluster"
	ClusterControllerImage            = "docker.io/weaveworks/cluster-controller"
	ClusterControllerImageTag         = "v1.5.2"
)

// InstallWge installs weave gitops enterprise chart
func InstallWge(version string) (string, error) {
	domainTypes := []string{
		DomainTypelocalgost,
		DomainTypeExternalDNS,
	}

	domainType, err := utils.GetSelectInput(DomainMsg, domainTypes)
	if err != nil {
		return "", err
	}

	userDomain := "localhost"

	if strings.Compare(domainType, DomainTypeExternalDNS) == 0 {

		utils.Warning(ExternalDNSWarningMsg)

		userDomain, err = utils.GetStringInput(ClusterDomainMsg, "")
		if err != nil {
			return "", err
		}

	}

	utils.Info(WGEInstallMsgFormat, version)

	pathInRepo, err := utils.CloneRepo()
	if err != nil {
		return "", err
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			utils.Warning("cleanup failed!")
		}
	}()

	wgehelmRepo, err := constructWgeHelmRepository()
	if err != nil {
		return "", err
	}

	err = utils.CreateFileToRepo(WGEHelmrepoFileName, wgehelmRepo, pathInRepo, WGEHelmRepoCommitMsg)
	if err != nil {
		return "", err
	}

	gitOpsSetsValues := map[string]interface{}{
		"enabled": true,
		"controllerManager": map[string]interface{}{
			"manager": map[string]interface{}{
				"args": []string{
					"--health-probe-bind-address=:8081",
					"--metrics-bind-address=127.0.0.1:8080",
					"--leader-elect",
					"--enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config",
				},
			},
		},
	}

	clusterController := domain.ClusterController{
		Enabled:          true,
		FullNameOverride: ClusterControllerFullOverrideName,
		ControllerManager: domain.ClusterControllerManager{
			Manager: domain.ClusterControllerManagerManager{
				Image: domain.ClusterControllerImage{
					Repository: ClusterControllerImage,
					Tag:        ClusterControllerImageTag,
				},
			},
		}}

	values := domain.ValuesFile{
		Ingress: constructIngressValues(userDomain),
		TLS: map[string]interface{}{
			"enabled": false,
		},
		GitOpsSets:        gitOpsSetsValues,
		EnablePipelines:   true,
		ClusterController: clusterController,
	}

	wgeHelmRelease, err := ConstructWGEhelmRelease(values, version)
	if err != nil {
		return "", err
	}

	err = utils.CreateFileToRepo(WGEHelmReleaseFileName, wgeHelmRelease, pathInRepo, WGEHelmReleaseCommitMsg)
	if err != nil {
		return "", err
	}

	err = utils.ReconcileFlux(WGEHelmReleaseName)
	if err != nil {
		return "", err
	}

	return userDomain, nil

}

func constructWgeHelmRepository() (string, error) {
	wgeHelmRepo := sourcev1.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      WGEHelmRepositoryName,
			Namespace: WGEDefaultNamespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: WGEChartUrl,
			Interval: v1.Duration{
				Duration: time.Minute,
			},
			SecretRef: &meta.LocalObjectReference{
				Name: EntitlementSecretName,
			},
		},
	}
	return utils.CreateHelmRepositoryYamlString(wgeHelmRepo)
}

func constructIngressValues(userDomain string) map[string]interface{} {
	ingressValues := map[string]interface{}{
		"annotations": map[string]string{
			"external-dns.alpha.kubernetes.io/hostname":                     userDomain,
			"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
			"service.beta.kubernetes.io/aws-load-balancer-type":             "nlb",
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

// ConstructWGEhelmRelease create the yaml resource for wge chart
func ConstructWGEhelmRelease(valuesFile domain.ValuesFile, chartVersion string) (string, error) {
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
					Chart:             WGEChartName,
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      WGEHelmRepositoryName,
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

func CheckUIDomain(userDomain string, wgeVersion string) error {
	if !strings.Contains(userDomain, "localhost") {

		utils.Info(InstallSuccessMsgFormat, wgeVersion, userDomain)
		return nil
	}

	utils.Info(LocalInstallSuccessMsgFormat, wgeVersion)
	var runner runner.CLIRunner
	out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		return fmt.Errorf("%s%s", err.Error(), string(out))
	}
	return nil
}
