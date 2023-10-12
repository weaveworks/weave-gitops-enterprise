package app

import "github.com/spf13/cobra"

// DisinheritAPIFlags turns off the required flag for CLI options related to accessing the Weave GitOps API.
//
// Currently, our top-level Command defines some top-level flags that are useful for most commands
// including endpoint,username, and passowrd
// but not all commands require access to the API, and so this can be used to disable top-level flags.
func DisinheritAPIFlags(cmd *cobra.Command, args []string) error {
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
}
