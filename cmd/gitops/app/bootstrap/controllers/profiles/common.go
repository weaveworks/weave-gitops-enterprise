package profiles

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

func InstallController(controllerValuesName string, controllerValues map[string]interface{}) error {
	values, err := utils.GetCurrentValuesForHelmRelease(commands.WGE_HELMRELEASE_NAME, commands.WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return utils.CheckIfError(err)
	}

	switch controllerValuesName {
	case POLICY_AGENT_VALUES_NAME:
		values.PolicyAgent = controllerValues
	}

	version, err := utils.GetCurrentVersionForHelmRelease(commands.WGE_HELMRELEASE_NAME, commands.WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return utils.CheckIfError(err)
	}

	helmRelease, err := commands.ConstructWGEhelmRelease(values, version)
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

	err = utils.CreateFileToRepo(commands.WGE_HELMRELEASE_FILENAME, helmRelease, pathInRepo, commands.WGE_HELMRELEASE_COMMITMSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.ReconcileFlux(commands.WGE_HELMRELEASE_NAME)
	if err != nil {
		return utils.CheckIfError(err)
	}

	return nil
}
