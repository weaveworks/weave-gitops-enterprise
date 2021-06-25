package cmd

import (
	"os"

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
	templatesCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", os.Getenv("CAPI_TEMPLATES_API_URL"), "The CAPI templates HTTP API endpoint")
}
