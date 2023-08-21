package controllers

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PolicyAgentCommand = &cobra.Command{
	Use:   "policy-agent",
	Short: "Bootstraps Weave Policy Agent",
	Example: `
# Bootstrap Weave Policy Agent
gitops bootstrap controllers policy-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return BootstrapPolicyAgent()
	},
}

func BootstrapPolicyAgent() error {
	fmt.Println("installing policy agent ...")
	return nil
}
