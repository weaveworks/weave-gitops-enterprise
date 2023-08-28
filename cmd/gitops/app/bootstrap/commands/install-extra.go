package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

func UpdateHelmReleaseValues(controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(WGE_HELMRELEASE_NAME, WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return err
	}

	switch controllerValuesName {
	case domain.POLICY_AGENT_VALUES_NAME:
		values.PolicyAgent = controllerValues
	case domain.OIDC_VALUES_NAME:
		values.Config.OIDC = controllerValues
	case domain.CAPI_VALUES_NAME:
		values.Config.CAPI = controllerValues
		values.Global.CapiEnabled = true
	case domain.TERRAFORM_VALUES_NAME:
		values.EnableTerraformUI = true
	}

	version, err := utils.GetCurrentVersionForHelmRelease(WGE_HELMRELEASE_NAME, WGE_DEFAULT_NAMESPACE)
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
			utils.Warning("cleanup failed!")
		}
	}()

	err = utils.CreateFileToRepo(WGE_HELMRELEASE_FILENAME, helmRelease, pathInRepo, WGE_HELMRELEASE_COMMITMSG)
	if err != nil {
		return err
	}

	err = utils.ReconcileFlux(WGE_HELMRELEASE_NAME)
	if err != nil {
		return err
	}

	return nil
}
