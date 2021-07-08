package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/formatter"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

func templatesListCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "list",
		Short:         "List CAPI templates",
		Example:       `mccp templates list`,
		RunE:          getTemplatesListCmdRun(client),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}

func getTemplatesListCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(endpoint, client)
		if err != nil {
			return err
		}
		w := formatter.NewTableWriter()
		defer w.Flush()

		return templates.ListTemplates(r, w)
	}
}
