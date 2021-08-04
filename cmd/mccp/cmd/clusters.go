package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func clustersCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "clusters",
		Short:         "Interact with Kubernetes clusters",
		Example:       `mccp clusters`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		clustersListCmd(client),
		clustersGetCmd(client),
		clustersDeleteCmd(client),
	)

	return cmd
}
