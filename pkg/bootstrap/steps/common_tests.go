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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeGitClient struct {
}

func (f fakeGitClient) CloneRepo(kubeClient client.Client, repoName string, namespace string, authType string, privateKeyPath string, privateKeyPassword string, username string, token string) (string, error) {
	return "/", nil
}

func (f fakeGitClient) CreateFileToRepo(filename, filecontent, path, commitmsg, authType, privateKeyPath, privateKeyPassword, username, token string) error {
	return nil
}

type fakeFluxClient struct {
}

func (f fakeFluxClient) ReconcileFlux() error {
	return nil
}

func (f fakeFluxClient) ReconcileHelmRelease(hrName string) error {
	return nil
}

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) Config {
	fakeClient := utils.CreateFakeClient(t, objects...)

	cliLogger := utils.CreateLogger()

	return Config{
		KubernetesClient:        fakeClient,
		GitClient:               fakeGitClient{},
		FluxClient:              fakeFluxClient{},
		Logger:                  cliLogger,
		WGEVersion:              config.WGEVersion,
		ClusterUserAuth:         config.ClusterUserAuth,
		GitRepository:           config.GitRepository,
		FluxInstallated:         config.FluxInstallated,
		PrivateKeyPath:          config.PrivateKeyPath,
		PrivateKeyPassword:      config.PrivateKeyPassword,
		GitUsername:             config.GitUsername,
		GitToken:                config.GitToken,
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
