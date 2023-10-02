package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	oidcValuesName = "oidc"
)

// updateHelmReleaseValues add the extra HelmRelease values.
func updateHelmReleaseValues(cl client.Client, controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(cl, WGEHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case oidcValuesName:
		values.Config.OIDC = controllerValues
	}

	version, err := utils.GetHelmReleaseProperty(cl, WGEHelmReleaseName, WGEDefaultNamespace, "version")
	if err != nil {
		return err
	}

	helmRelease, err := constructWGEhelmRelease(values, version)
	if err != nil {
		return err
	}

	pathInRepo, err := utils.CloneRepo(cl, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			// TODO: handle error
		}
	}()

	if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, helmRelease, pathInRepo, wgeHelmReleaseCommitMsg); err != nil {
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
