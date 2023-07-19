package shard

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

type profileCommandFlags struct {
	namespace string
}

var flags profileCommandFlags

func GetCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "shard metrics",
		Aliases:      []string{"shard"},
		Short:        "Get shard metrics",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,

		SilenceErrors: true,
		Example: `
	# Get shard metrics
	gitops get shard metrics source-controller-shardset -n  flux-system
	`,
		PreRunE: getShardMetricsCmdPreRunE(&opts.Endpoint),
		RunE:    getShardMetricsCmdRunE(opts, client),
	}

	cmd.Flags().StringVar(&flags.namespace, "namespace", "flux-system", "Namespace of the repository")

	return cmd
}

func getShardMetricsCmdPreRunE(endpoint *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func getShardMetricsCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}
