package run

import (
	"errors"

	"github.com/spf13/cobra"
)

// Cmd to start the event-writer process
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Start event-writer process",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runCommand()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runCommand() error {
	return errors.New("not implemented")
}
