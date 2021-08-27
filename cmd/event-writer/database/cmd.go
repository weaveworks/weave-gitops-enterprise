package database

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/database/create"
)

// Cmd group for database operations
var Cmd = &cobra.Command{
	Use:   "database",
	Short: "MCCP database operations",
}

func init() {
	Cmd.AddCommand(create.Cmd)
}
