package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const RepoCleanupMsg = "Cleaning up repo ..."

func UpdateHelmReleaseValues(controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(WGEHelmReleaseName, WGEDefaultNamespace)
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

	version, err := utils.GetCurrentVersionForHelmRelease(WGEHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	helmRelease, err := ConstructWGEhelmRelease(values, version)
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
			utils.Warning(RepoCleanupMsg)
		}
	}()

	err = utils.CreateFileToRepo(WGEHelmReleaseFileName, helmRelease, pathInRepo, WGEHelmReleaseCommitMsg)
	if err != nil {
		return err
	}

	err = utils.ReconcileFlux(WGEHelmReleaseName)
	if err != nil {
		return err
	}

	return nil
}
