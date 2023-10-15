package steps

import (
	"context"
	"encoding/json"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	oidcValuesName = "oidc"
)

// updateHelmReleaseValues add the extra HelmRelease values.
func updateHelmReleaseValues(c *Config, controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := getCurrentValuesForHelmRelease(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case oidcValuesName:
		values.Config.OIDC = controllerValues
	}

	version, err := utils.GetHelmReleaseProperty(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace, "version")
	if err != nil {
		fmt.Println("error getting helm release version: ", err)
		return err
	}

	helmRelease, err := constructWGEhelmRelease(values, version)
	if err != nil {
		fmt.Println("error constructing helm release: ", err)
		return err
	}

	pathInRepo, err := utils.CloneRepo(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace, c.PrivateKeyPath, c.PrivateKeyPassword)
	if err != nil {
		fmt.Println("error cloning repo: ", err)
		return err
	}

	if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, helmRelease, pathInRepo, wgeHelmReleaseCommitMsg, c.PrivateKeyPath, c.PrivateKeyPassword); err != nil {
		return err
	}

	if err := utils.ReconcileFlux(); err != nil {
		return err
	}
	if err := utils.ReconcileHelmRelease(WgeHelmReleaseName); err != nil {
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
