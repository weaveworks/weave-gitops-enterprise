package commands

import (
	"encoding/json"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	ENTITLEMENT_SECRET_NAME   = "weave-gitops-enterprise-credentials"
	ADMIN_SECRET_NAME         = "cluster-user-auth"
	DEFAULT_ADMIN_USERNAME    = "wego-admin"
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

func constructWGEhelmRelease(userDomain string, chartVersion string) (string, error) {

	tlsValues := map[string]interface{}{
		"enabled": false,
	}

	values := map[string]interface{}{
		"ingress": constructIngressValues(userDomain),
		"tls":     tlsValues,
	}

	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	wgeHelmRelease := helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              WGE_HELMRELEASE_NAME,
			Namespace:         WGE_DEFAULT_NAMESPACE,
			CreationTimestamp: v1.Now(),
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             WGE_CHART_NAME,
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
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

	wgeHelmReleaseBytes, err := k8syaml.Marshal(wgeHelmRelease)
	if err != nil {
		return "", utils.CheckIfError(err)
	}
	return string(wgeHelmReleaseBytes), nil
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

func constructWgeHelmRepository() (string, error) {
	wgeHelmRepo := sourcev1.HelmRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: sourcev1.GroupVersion.Identifier(),
			Kind:       sourcev1.HelmRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              WGE_HELMREPOSITORY_NAME,
			Namespace:         WGE_DEFAULT_NAMESPACE,
			CreationTimestamp: v1.Now(),
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
	wgeHelmRepoBytes, err := k8syaml.Marshal(wgeHelmRepo)
	if err != nil {
		return "", utils.CheckIfError(err)
	}
	return string(wgeHelmRepoBytes), nil
}
