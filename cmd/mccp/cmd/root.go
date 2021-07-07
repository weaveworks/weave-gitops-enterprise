package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func RootCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "mccp",
		Short: "MCCP CLI",
	}

	cmd.AddCommand(
		templatesCmd(client),
	)

	return cmd
}

func Execute() error {
	return RootCmd(resty.New()).Execute()
}
