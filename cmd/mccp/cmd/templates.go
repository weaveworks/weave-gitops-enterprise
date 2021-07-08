package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func templatesCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "templates",
		Short:         "Interact with CAPI templates",
		Example:       `mccp templates`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		templatesListCmd(client),
		templatesRenderCmd(client),
	)

	return cmd
}
