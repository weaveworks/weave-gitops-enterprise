package steps

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) Config {
	fakeClient := utils.CreateFakeClient(t, objects...)
	cliLogger := utils.CreateLogger()
	return Config{
		KubernetesClient:        fakeClient,
		Logger:                  cliLogger,
		WGEVersion:              config.WGEVersion,
		ClusterUserAuth:         config.ClusterUserAuth,
		DomainType:              config.DomainType,
		UserDomain:              config.UserDomain,
		GitScheme:               config.GitScheme,
		FluxInstallated:         config.FluxInstallated,
		PrivateKeyPath:          config.PrivateKeyPath,
		PrivateKeyPassword:      config.PrivateKeyPassword,
		GitUsername:             config.GitUsername,
		GitToken:                config.GitToken,
		RepoURL:                 config.RepoURL,
		Branch:                  config.Branch,
		RepoPath:                config.RepoPath,
		AuthType:                config.AuthType,
		InstallOIDC:             config.InstallOIDC,
		DiscoveryURL:            config.DiscoveryURL,
		IssuerURL:               config.IssuerURL,
		ClientID:                config.ClientID,
		ClientSecret:            config.ClientSecret,
		RedirectURL:             config.RedirectURL,
		PromptedForDiscoveryURL: config.PromptedForDiscoveryURL,
		Silent:                  config.Silent,
		ExtraComponents:         config.ExtraComponents,
	}
}

func createWGEHelmReleaseFakeObject(version string) (helmv2.HelmRelease, error) {
	values := valuesFile{
		TLS: map[string]interface{}{
			"enabled": false,
		},
		GitOpsSets: map[string]interface{}{
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
		},
		EnablePipelines: true,
		ClusterController: clusterController{
			Enabled:          true,
			FullNameOverride: clusterControllerFullOverrideName,
			ControllerManager: clusterControllerManager{
				Manager: clusterControllerManagerManager{
					Image: clusterControllerImage{
						Repository: clusterControllerImageName,
						Tag:        clusterControllerImageTag,
					},
				},
			}},
	}

	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return helmv2.HelmRelease{}, err
	}

	wgeHRObject := helmv2.HelmRelease{
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
					Version: version,
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
	return wgeHRObject, nil
}
