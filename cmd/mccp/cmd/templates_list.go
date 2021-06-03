package cmd

import (
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/wks/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"

	"github.com/spf13/cobra"
)

var templatesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List CAPI templates",
	Example: `mccp templates list`,
	Run:     templatesListCmdRun,
	Args: func(cmd *cobra.Command, args []string) error {
		_, err := url.ParseRequestURI(endpoint)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	templatesCmd.AddCommand(templatesListCmd)
}

func templatesListCmdRun(cmd *cobra.Command, args []string) {
	r := adapters.NewHttpTemplateRetriever(endpoint, resty.New())
	templates.ListTemplates(r, os.Stdout)
}
