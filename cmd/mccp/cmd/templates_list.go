package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/templates"
	"k8s.io/cli-runtime/pkg/printers"
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

	cmd.PersistentFlags().StringVar(&templatesListCmdFlags.Provider, "provider", "", "Filters templates by provider name")

	return cmd
}

type templatesListFlags struct {
	Provider string
}

var templatesListCmdFlags templatesListFlags

func getTemplatesListCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(endpoint, client, os.Stdout)
		if err != nil {
			return err
		}
		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		return templates.ListTemplates(templatesListCmdFlags.Provider, r, w)
	}
}
