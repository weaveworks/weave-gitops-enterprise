package cmd

import (
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:     "templates",
	Short:   "Interact with CAPI templates",
	Example: `mccp templates`,
}

var (
	endpoint string
)

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "The CAPI templates HTTP API endpoint")
	templatesCmd.MarkPersistentFlagRequired("endpoint")
}
