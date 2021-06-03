package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mccp",
	Short: "MCCP CLI",
}

func Execute() error {
	return rootCmd.Execute()
}
