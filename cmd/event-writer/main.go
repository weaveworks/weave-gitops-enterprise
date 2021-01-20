package main

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/event-writer/database"
	"github.com/weaveworks/wks/cmd/event-writer/run"
	"github.com/weaveworks/wks/pkg/cmdutil"
)

var cmd = &cobra.Command{
	Use:           "event-writer",
	Short:         "Writer of messages received from the MCCP NATS server to a relational database.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func main() {
	cmd.AddCommand(database.Cmd)
	cmd.AddCommand(run.Cmd)

	if err := cmd.Execute(); err != nil {
		cmdutil.ErrorExit("Error", err)
	}
}
