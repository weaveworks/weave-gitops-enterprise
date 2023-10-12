package bootstrap

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	. "github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const (
	autCmdName             = "auth"
	autCmdShortDescription = "Generate authentication configuration for Weave GitOps. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported."
	authCmdExamples        = `
	## Add OIDC configuration to your cluster. 
	gitops bootstrap auth --type=oidc
	`
)
const (
	authOIDC = "oidc"
)

type authConfigFlags struct {
	authType     string
	domain       string
	wgeVersion   string
	discoveryURL string
	clientID     string
	clientSecret string
}

var authFlags authConfigFlags

func AuthCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     autCmdName,
		Short:   autCmdShortDescription,
		Example: authCmdExamples,
		Run:     getAuthCmdRun(opts),
	}
	cmd.Flags().StringVarP(&authFlags.authType, "type", "t", "", "type of authentication to be configured")
	cmd.Flags().StringVarP(&authFlags.discoveryURL, "discovery-url", "", "", "OIDC discovery URL")
	cmd.Flags().StringVarP(&authFlags.clientID, "client-id", "i", "", "OIDC client ID")
	cmd.Flags().StringVarP(&authFlags.clientSecret, "client-secret", "s", "", "OIDC client secret")

	return cmd
}

func getAuthCmdRun(opts *config.Options) func(*cobra.Command, []string) error {

	err := addWGEFlags(opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return func(cmd *cobra.Command, args []string) error {
		cliLogger := logger.NewCLILogger(os.Stdout)

		c, err := steps.NewConfigBuilder().
			WithLogWriter(cliLogger).
			WithKubeconfig(opts.Kubeconfig).
			WithVersion(authFlags.wgeVersion).
			WithDomain(authFlags.domain).
			WithOIDCConfig(authFlags.discoveryURL, authFlags.clientID, authFlags.clientSecret).
			WithPromptedForDiscoveryURL(false).
			Build()

		if err != nil {
			return fmt.Errorf("cannot config bootstrap: %v", err)

		}

		//use bootstrapAuth function to bootstrap the authentication
		err = BootstrapAuth(c)
		if err != nil {
			return fmt.Errorf("cannot bootstrap: %v", err)
		}

		return nil

	}
}

func addWGEFlags(opts *config.Options) error {

	//get kubernetes client
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}

	authFlags.wgeVersion, err = utils.GetHelmReleaseProperty(kubernetesClient, steps.WGEHelmReleaseName, steps.WGEDefaultNamespace, utils.HelmVersionProperty)
	if err != nil {
		return err
	}

	authFlags.domain, err = utils.GetHelmReleaseProperty(kubernetesClient, steps.WGEHelmReleaseName, steps.WGEDefaultNamespace, utils.HelmDomainProperty)
	if err != nil {
		return err
	}

	return nil
}
