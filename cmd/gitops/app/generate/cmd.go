package generate

import (
	"github.com/spf13/cobra"
	gitopssets "github.com/weaveworks/gitopssets-controller/pkg/cmd"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "generate",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Generate one or more resources",
	}

	generateGitOpsSetCmd := gitopssets.NewGenerateCommand("gitopsset")
	cmd.AddCommand(generateGitOpsSetCmd)

	return cmd
}
