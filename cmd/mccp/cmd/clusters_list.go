package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/clusters"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/spf13/cobra"
)

func clustersListCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "list",
		Short:         "List Kubernetes clusters",
		Example:       `mccp clusters list`,
		RunE:          getClustersListCmdRun(client),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}

func getClustersListCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		r, err := adapters.NewHttpClient(endpoint, client, os.Stdout)
		if err != nil {
			return err
		}
		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		return clusters.ListClusters(r, w)
	}
}
