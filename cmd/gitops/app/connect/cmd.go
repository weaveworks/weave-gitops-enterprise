package connect

import (
	"errors"

	"github.com/spf13/cobra"
	connect "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/connect/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect clusters",
		Example: `
# Connect remote cluster
gitops connect cluster`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			names := []string{
				"endpoint",
				"password",
				"username",
			}
			flags := cmd.InheritedFlags()
			for _, name := range names {
				err := flags.SetAnnotation(name, cobra.BashCompOneRequiredFlag, []string{"false"})
				if err != nil {
					return errors.New("error setting annotation")
				}
			}
			return nil
		},
	}

	cmd.AddCommand(connect.ConnectCommand(opts, client))

	return cmd
}
