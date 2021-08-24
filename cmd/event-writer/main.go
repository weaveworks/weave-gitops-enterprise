package main

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/database"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/run"
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
		panic(err)
	}
}
