package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func upgradeCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "upgrade",
		Short:         "Upgrade to WGE",
		Example:       `mccp upgrade`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}
