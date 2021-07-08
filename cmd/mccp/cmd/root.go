package cmd

import (
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func RootCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "mccp",
		Short: "MCCP CLI",
		Args: func(*cobra.Command, []string) error {
			_, err := url.ParseRequestURI(endpoint)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", os.Getenv("MCCP_API_URL"), "The MCCP HTTP API endpoint")

	cmd.AddCommand(
		templatesCmd(client),
		clustersCmd(client),
	)

	return cmd
}

func Execute() error {
	return RootCmd(resty.New()).Execute()
}

var (
	endpoint string
)
