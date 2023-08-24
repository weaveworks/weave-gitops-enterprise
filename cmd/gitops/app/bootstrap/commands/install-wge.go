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
	DOMAIN_MSG                = "Please select the domain to be used"
	CLUSTER_DOMAIN_MSG        = "Please enter your cluster domain"
	WGE_CHART_NAME            = "mccp"
	WGE_HELMREPOSITORY_NAME   = "weave-gitops-enterprise-charts"
	WGE_HELMRELEASE_NAME      = "weave-gitops-enterprise"
	WGE_DEFAULT_NAMESPACE     = "flux-system"
	DOMAIN_TYPE_LOCALHOST     = "localhost (Using Portforward)"
	DOMAIN_TYPE_EXTERNALDNS   = "external DNS"
	WGE_HELMREPO_FILENAME     = "wge-hrepo.yaml"
	WGE_HELMRELEASE_FILENAME  = "wge-hrelease.yaml"
	WGE_HELMREPO_COMMITMSG    = "Add WGE HelmRepository YAML file"
	WGE_HELMRELEASE_COMMITMSG = "Add WGE HelmRelease YAML file"
	WGE_CHART_URL             = "https://charts.dev.wkp.weave.works/releases/charts-v3"
)

// InstallWge installs weave gitops enterprise chart
func InstallWge(version string) (string, error) {
	domainTypes := []string{
		DOMAIN_TYPE_LOCALHOST,
		DOMAIN_TYPE_EXTERNALDNS,
	}

	domainType, err := utils.GetSelectInput(DOMAIN_MSG, domainTypes)
	if err != nil {
		return "", err
	}

	userDomain := "localhost"

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {

		utils.Warning("\n\nPlease make sure to have the external DNS service is installed in your cluster, or you have a domain points to your cluster\nFor more information about external DNS please refer to https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html\n\n")

		userDomain, err = utils.GetStringInput(CLUSTER_DOMAIN_MSG, "")
		if err != nil {
			return "", err
		}

	}

	utils.Info("All set installing WGE v%s, This may take few minutes...\n", version)

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

	err = utils.CreateFileToRepo(WGE_HELMREPO_FILENAME, wgehelmRepo, pathInRepo, WGE_HELMREPO_COMMITMSG)
	if err != nil {
		return "", err
	}

	values := domain.ValuesFile{
		Ingress: constructIngressValues(userDomain),
		TLS: map[string]interface{}{
			"enabled": false,
		},
	}

	wgeHelmRelease, err := ConstructWGEhelmRelease(values, version)
	if err != nil {
		return "", err
	}

	err = utils.CreateFileToRepo(WGE_HELMRELEASE_FILENAME, wgeHelmRelease, pathInRepo, WGE_HELMRELEASE_COMMITMSG)
	if err != nil {
		return "", err
	}

	err = utils.ReconcileFlux(WGE_HELMRELEASE_NAME)
	if err != nil {
		return "", err
	}

	return userDomain, nil

}

func constructWgeHelmRepository() (string, error) {
	wgeHelmRepo := sourcev1.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      WGE_HELMREPOSITORY_NAME,
			Namespace: WGE_DEFAULT_NAMESPACE,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: WGE_CHART_URL,
			Interval: v1.Duration{
				Duration: time.Minute,
			},
			SecretRef: &meta.LocalObjectReference{
				Name: ENTITLEMENT_SECRET_NAME,
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
			Name:      WGE_HELMRELEASE_NAME,
			Namespace: WGE_DEFAULT_NAMESPACE,
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             WGE_CHART_NAME,
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      WGE_HELMREPOSITORY_NAME,
						Namespace: WGE_DEFAULT_NAMESPACE,
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

		utils.Info("\nWGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", wgeVersion, userDomain)
		return nil
	}

	var runner runner.CLIRunner
	out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		return fmt.Errorf("%s%s", err.Error(), string(out))
	}
	return nil
}
