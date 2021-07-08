package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/wks/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/clusters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/formatter"

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
		r, err := adapters.NewHttpClient(endpoint, client)
		if err != nil {
			return err
		}
		w := formatter.NewTableWriter()
		defer w.Flush()

		return clusters.ListClusters(r, w)
	}
}
