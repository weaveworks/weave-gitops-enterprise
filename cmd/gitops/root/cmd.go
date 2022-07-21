package root

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/create"
)

func init() {
	// Setup flag to env mapping:
	//   config-repo => GITOPS_CONFIG_REPO
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("WEAVE_GITOPS")

	viper.AutomaticEnv()
}

func RootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:           "gitops",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Weave GitOps Enterprise",
		Long:          "Command line utility for managing Kubernetes applications via GitOps.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Sync flag values and env vars.
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				log.Fatalf("Error binding viper to flags: %v", err)
			}
		},
	}

	rootCmd.AddCommand(create.CreateCommand())

	return rootCmd
}
