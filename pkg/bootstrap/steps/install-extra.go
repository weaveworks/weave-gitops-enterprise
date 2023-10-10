package steps

import (
	"context"
	"encoding/json"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	oidcValuesName = "oidc"
)

// updateHelmReleaseValues add the extra HelmRelease values.
func updateHelmReleaseValues(c *Config, controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := getCurrentValuesForHelmRelease(c.KubernetesClient, WGEHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case oidcValuesName:
		values.Config.OIDC = controllerValues
	}

	version, err := utils.GetHelmReleaseProperty(c.KubernetesClient, WGEHelmReleaseName, WGEDefaultNamespace, "version")
	if err != nil {
		return err
	}

	helmRelease, err := constructWGEhelmRelease(values, version)
	if err != nil {
		return err
	}

	pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace, c.PrivateKeyPath, c.PrivateKeyPassword)
	if err != nil {
		return err
	}

	defer utils.CleanupRepo()

	if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, helmRelease, pathInRepo, wgeHelmReleaseCommitMsg, c.PrivateKeyPath, c.PrivateKeyPassword); err != nil {
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

// GetCurrentValuesForHelmRelease gets the current values from a specific helmrelease.
func getCurrentValuesForHelmRelease(client k8s_client.Client, releaseName string, namespace string) (valuesFile, error) {
	helmrelease := &helmv2.HelmRelease{}
	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Namespace: namespace,
		Name:      releaseName,
	}, helmrelease); err != nil {
		return valuesFile{}, err
	}

	values := valuesFile{}
	if err := json.Unmarshal(helmrelease.Spec.Values.Raw, &values); err != nil {
		return valuesFile{}, err
	}

	return values, nil
}
