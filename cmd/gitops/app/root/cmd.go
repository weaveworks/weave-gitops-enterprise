package root

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/add"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/connect"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/create"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/delete"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/disconnect"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/generate"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/update"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/upgrade"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/check"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/docs"
	"github.com/weaveworks/weave-gitops/cmd/gitops/set"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/analytics"
	analyticsconfig "github.com/weaveworks/weave-gitops/pkg/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/client-go/rest"
)

const defaultNamespace = "flux-system"

var options = &config.Options{}

// Only want AutomaticEnv to be called once!
func init() {
	// Setup flag to env mapping:
	//   config-repo => GITOPS_CONFIG_REPO
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("WEAVE_GITOPS")

	viper.AutomaticEnv()
}

func Command(client *adapters.HTTPClient) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:           "gitops",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Weave GitOps",
		Long:          "Command line utility for managing Kubernetes applications via GitOps.",
		Example: `
  # Get help for gitops add cluster command
  gitops add cluster -h
  gitops help add cluster

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Sync flag values and env vars.
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				log.Fatalf("Error binding viper to flags: %v", err)
			}

			ns, _ := cmd.Flags().GetString("namespace")

			if ns == "" {
				return
			}

			if nserr := utils.ValidateNamespace(ns); nserr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", nserr)
				os.Exit(1)
			}
			if options.OverrideInCluster {
				kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			}

			err = cmd.Flags().Set("username", viper.GetString("username"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			err = cmd.Flags().Set("password", viper.GetString("password"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			gitopsConfig, err := analyticsconfig.GetConfig(false)
			if err != nil {
				seed := time.Now().UnixNano()

				gitopsConfig = &analyticsconfig.GitopsCLIConfig{
					UserID:    analyticsconfig.GenerateUserID(10, seed),
					Analytics: true,
				}

				_ = analyticsconfig.SaveConfig(gitopsConfig)
			}

			if gitopsConfig.Analytics {
				_ = analytics.TrackCommand(cmd, gitopsConfig.UserID)
			}
		},
	}

	rootCmd.PersistentFlags().StringP("namespace", "n", defaultNamespace, "The namespace scope for this operation")
	rootCmd.PersistentFlags().StringVarP(&options.Endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint can be set with `WEAVE_GITOPS_ENTERPRISE_API_URL` environment variable")
	rootCmd.PersistentFlags().StringVarP(&options.Username, "username", "u", "", "The Weave GitOps Enterprise username for authentication can be set with `WEAVE_GITOPS_USERNAME` environment variable")
	rootCmd.PersistentFlags().StringVarP(&options.Password, "password", "p", "", "The Weave GitOps Enterprise password for authentication can be set with `WEAVE_GITOPS_PASSWORD` environment variable")
	rootCmd.PersistentFlags().BoolVar(&options.OverrideInCluster, "override-in-cluster", false, "override running in cluster check")
	rootCmd.PersistentFlags().StringToStringVar(&options.GitHostTypes, "git-host-types", map[string]string{}, "Specify which custom domains are running what (github, gitlab or bitbucket-server)")
	rootCmd.PersistentFlags().BoolVar(&options.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	rootCmd.PersistentFlags().StringVar(&options.Kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("override-in-cluster"))
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("git-host-types"))

	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(get.Command(options, client))
	rootCmd.AddCommand(add.Command(options, client))
	rootCmd.AddCommand(create.Command())
	rootCmd.AddCommand(update.Command(options, client))
	rootCmd.AddCommand(delete.Command(options, client))
	rootCmd.AddCommand(upgrade.Cmd)
	rootCmd.AddCommand(docs.Cmd)
	rootCmd.AddCommand(check.GetCommand(options))
	rootCmd.AddCommand(set.SetCommand(options))
	rootCmd.AddCommand(generate.Command())
	rootCmd.AddCommand(bootstrap.Command(options))
	rootCmd.AddCommand(connect.Command(options))
	rootCmd.AddCommand(disconnect.Command(options))

	return rootCmd
}
