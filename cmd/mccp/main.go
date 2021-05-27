package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/mccp/listtemplates"
	"github.com/weaveworks/wks/pkg/cmdutil"
)

var rootCmd = &cobra.Command{
	Use:   "mccp",
	Short: "MCCP CLI",
}

func main() {
	fmt.Println("Welcome to the mccp")

	rootCmd.AddCommand(listtemplates.Cmd)

	if err := rootCmd.Execute(); err != nil {
		cmdutil.ErrorExit("Error", err)
	}
}
