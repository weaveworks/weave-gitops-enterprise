package disconnect

import (
	"github.com/spf13/cobra"
	disconnect "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/disconnect/clusters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect clusters",
		Example: `
# Disconnect a cluster
gitops disconnect cluster`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			names := []string{
				"endpoint",
				"password",
				"username",
			}
			flags := cmd.InheritedFlags()
			for _, name := range names {
				err := flags.SetAnnotation(name, cobra.BashCompOneRequiredFlag, []string{"false"})
				return err
			}
			return nil
		},
	}

	cmd.AddCommand(disconnect.DisconnectCommand(opts))

	return cmd
}
