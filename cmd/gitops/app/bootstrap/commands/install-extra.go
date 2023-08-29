package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

// UpdateHelmReleaseValues add the extra HelmRelease values.
func UpdateHelmReleaseValues(controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(wgeHelmReleaseName, wgeDefaultNamespace)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case domain.PolicyAgentValuesName:
		values.PolicyAgent = controllerValues
	case domain.OIDCValuesName:
		values.Config.OIDC = controllerValues
	case domain.CAPIValuesName:
		values.Config.CAPI = controllerValues
		values.Global.CapiEnabled = true
	case domain.TerraformValuesName:
		values.EnableTerraformUI = true
	}

	version, err := utils.GetCurrentVersionForHelmRelease(wgeHelmReleaseName, wgeDefaultNamespace)
	if err != nil {
		return err
	}

	helmRelease, err := constructWGEhelmRelease(values, version)
	if err != nil {
		return err
	}

	pathInRepo, err := utils.CloneRepo()
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

	if err := utils.ReconcileFlux(wgeHelmReleaseName); err != nil {
		return err
	}

	return nil
}
