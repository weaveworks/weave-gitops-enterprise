//go:build integration
// +build integration

package bootstrap_test

import (
	"testing"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"

	. "github.com/onsi/gomega"
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

// TestBootstrapCmd is an integration test for bootstrapping command.
// It uses envtest to simulate a cluster.
func TestBootstrapCmd(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	tests := []struct {
		name             string
		flags            []string
		expectedErrorStr string
	}{
		{
			name:             "should fail without entitlements",
			flags:            []string{},
			expectedErrorStr: "entitlement file is not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := adapters.NewHTTPClient()

			cmd := root.Command(client)
			bootstrapCmdArgs := []string{"bootstrap"}
			bootstrapCmdArgs = append(bootstrapCmdArgs, tt.flags...)
			cmd.SetArgs(bootstrapCmdArgs)

			err := cmd.Execute()

			if err != nil {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrorStr))
			} else {
				g.Expect(err).To(BeNil())
			}

		})
	}
}
