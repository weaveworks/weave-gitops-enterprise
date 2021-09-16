package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/upgrade"
)

func upgradeCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "upgrade",
		Short:         "Upgrade to WGE",
		Example:       `mccp upgrade`,
		RunE:          upgradeCmdRun(client),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}

func upgradeCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return upgrade.Upgrade(os.Stdout)
	}
}
