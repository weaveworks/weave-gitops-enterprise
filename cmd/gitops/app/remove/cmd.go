package remove

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/remove/run"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func RemoveCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove various components of Weave GitOps",
	}

	cmd.AddCommand(run.RunCommand(opts))

	return cmd
}
