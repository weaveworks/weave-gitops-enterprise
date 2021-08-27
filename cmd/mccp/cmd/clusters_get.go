package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/clusters"
)

func clustersGetCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "get",
		Short:         "Get CAPI cluster",
		Example:       "mccp clusters get <cluster-name>",
		RunE:          getClustersGetCmdRun(client),
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&clustersGetCmdFlags.Kubeconfig, "kubeconfig", false, "Returns the Kubeconfig of the workload cluster")

	return cmd
}

type clustersGetFlags struct {
	Kubeconfig bool
}

var clustersGetCmdFlags clustersGetFlags

func getClustersGetCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(endpoint, client)
		if err != nil {
			return err
		}

		if clustersGetCmdFlags.Kubeconfig {
			return clusters.GetClusterKubeconfig(args[0], r, os.Stdout)
		}

		return nil
	}
}
