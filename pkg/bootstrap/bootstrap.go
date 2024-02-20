package bootstrap

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap workflow
func Bootstrap(config steps.Config) error {

	adminCredentials, err := steps.NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.ModesConfig)
	if err != nil {
		return fmt.Errorf("cannot create ask admin creds step: %v", err)
	}

	repositoryConfig := steps.NewGitRepositoryConfigStep(config.GitRepository)

	checkUiDomain, err := steps.NewCheckUIDomainStep(config.ModesConfig)
	if err != nil {
		return fmt.Errorf("cannot create check ui: %v", err)
	}

	installWge, err := steps.NewInstallWGEStep(config.WgeConfig, config.Logger)
	if err != nil {
		return fmt.Errorf("cannot create install WGE: %v", err)
	}

	componentsExtra := steps.NewInstallExtraComponentsStep(config.ComponentsExtra, config.ModesConfig.Silent)

	// TODO have a single workflow source of truth and documented in https://docs.gitops.weave.works/docs/0.33.0/enterprise/getting-started/install-enterprise/
	var workflow = []steps.BootstrapStep{
		steps.NewAskBootstrapFluxStep(config),
		repositoryConfig,
		steps.NewBootstrapFlux(config),
		adminCredentials,
		installWge,
		steps.NewInstallOIDCStep(config),
		steps.NewOIDCConfigStep(config),
		componentsExtra,
		checkUiDomain,
	}

	return execute(config, workflow)
}

func execute(config steps.Config, worfklow []steps.BootstrapStep) error {
	var allOutputs []steps.StepOutput

	for _, step := range worfklow {
		config.Logger.Waitingf(step.Name)
		stepOutputs, err := step.Execute(&config)
		if err != nil {
			return fmt.Errorf("error on step %s: %v", step.Name, err)
		}
		allOutputs = append(allOutputs, stepOutputs...)
	}

	if config.ModesConfig.Export {
		config.Logger.Actionf("export manifests")
		for _, output := range allOutputs {
			err := output.Export(config.OutWriter)
			if err != nil {
				return fmt.Errorf("error exporting output %s: %v", output.Name, err)
			}
		}
	}

	return nil
}
