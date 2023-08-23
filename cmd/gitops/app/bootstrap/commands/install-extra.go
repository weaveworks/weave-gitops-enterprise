package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

func InstallController(controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(WGE_HELMRELEASE_NAME, WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return utils.CheckIfError(err)
	}

	switch controllerValuesName {
	case domain.POLICY_AGENT_VALUES_NAME:
		values.PolicyAgent = controllerValues
	}

	version, err := utils.GetCurrentVersionForHelmRelease(WGE_HELMRELEASE_NAME, WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return utils.CheckIfError(err)
	}

	helmRelease, err := ConstructWGEhelmRelease(values, version)
	if err != nil {
		return utils.CheckIfError(err)
	}

	pathInRepo, err := utils.CloneRepo()
	if err != nil {
		return utils.CheckIfError(err)
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			utils.Warning("cleanup failed!")
		}
	}()

	err = utils.CreateFileToRepo(WGE_HELMRELEASE_FILENAME, helmRelease, pathInRepo, WGE_HELMRELEASE_COMMITMSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.ReconcileFlux(WGE_HELMRELEASE_NAME)
	if err != nil {
		return utils.CheckIfError(err)
	}

	return nil
}
