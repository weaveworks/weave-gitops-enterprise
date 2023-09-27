package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UpdateHelmReleaseValues add the extra HelmRelease values.
func UpdateHelmReleaseValues(cl client.Client, controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(WGEHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case domain.OIDCValuesName:
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
			utils.Warning(utils.RepoCleanupMsg)
		}
	}()

	if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, helmRelease, pathInRepo, wgeHelmReleaseCommitMsg); err != nil {
		return err
	}

	if err := utils.ReconcileFlux(); err != nil {
		return err
	}

	return nil
}
