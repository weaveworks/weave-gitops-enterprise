//go:build integration
// +build integration

package bootstrap_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"

	. "github.com/onsi/gomega"
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

// TestQueryServer is an integration test for exercising the integration of the
// query system that includes both collecting from a cluster (using envtest) and doing queries via grpc.
// It is also used in the context of logging events per https://github.com/weaveworks/weave-gitops-enterprise/issues/2691
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
			name:  "can execute bootstrap command",
			flags: []string{},
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
			assert.NoError(t, err)

			if err != nil {
				g.Expect(err).To(ContainSubstring(tt.expectedErrorStr))

			} else {
				g.Expect(err).To(BeNil())
			}

		})
	}
}
