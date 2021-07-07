package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func templatesCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "templates",
		Short:   "Interact with CAPI templates",
		Example: `mccp templates`,
	}

	cmd.AddCommand(
		templatesListCmd(client),
		templatesRenderCmd(client),
	)

	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", os.Getenv("CAPI_TEMPLATES_API_URL"), "The CAPI templates HTTP API endpoint")

	return cmd
}

var (
	endpoint string
)
