//go:build integration
// +build integration

package server_test

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	api "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/test"
	l "github.com/weaveworks/weave-gitops/core/logger"
	"go.uber.org/zap/zapcore"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"
)

// Test case to ensure that we can debug issues via log events
// https://github.com/weaveworks/weave-gitops-enterprise/issues/2691
func TestServerIntegrationTest_Debug(t *testing.T) {
	g := NewGomegaWithT(t)
	testLog := testr.New(t)
	ctx := context.Background()

	//appLog, err := l.New("debug", false)

	appLog, loggerPath := createDebugLogger(t)
	//appLog, r, w := createDebugLogger(t)

	// setup app server
	c, err := makeGRPCServer(t, cfg, appLog, testLog)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		objects        []client.Object
		queryRequest   api.QueryRequest
		expectedEvents []string
	}{
		{
			name:         "can trace query server creation",
			objects:      []client.Object{},
			queryRequest: api.QueryRequest{},
			expectedEvents: []string{
				"collectors started",
				"query server created",
			},
		},
		//{
		//	name: "can trace new helm release object",
		//	objects: []client.Object{
		//		testutils.NewHelmRelease("createdOrUpdatedHelmRelease", "any"),
		//	},
		//	expectedEvents: []string{"debug message"},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given a new event
			test.Create(ctx, t, cfg, tt.objects...)
			//when processed
			query, err := c.DoQuery(context.Background(), &tt.queryRequest)
			g.Expect(err).To(BeNil())
			g.Expect(len(query.Objects)).To(BeIdenticalTo(10))
			//then processing events are found
			g.Expect(assertLogs(t, loggerPath, tt.expectedEvents)).To(Succeed())
		})
	}
}

func assertLogs(t *testing.T, loggerPath string, events []string) error {
	//get logs
	logs, err := os.ReadFile(loggerPath)
	if err != nil {
		return fmt.Errorf("cannot read logger file: %w", err)

	}

	logLines := strings.Split(string(logs), "\n")
	for _, event := range events {
		found := false
		for _, logLine := range logLines {
			if strings.Contains(logLine, event) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("event not found: %s", event)
		}
	}
	return nil
}

func createDebugLogger(t *testing.T) (logr.Logger, string) {
	g := NewGomegaWithT(t)

	file, err := os.CreateTemp(os.TempDir(), "query-server-log")
	g.Expect(err).ShouldNot(HaveOccurred())

	name := file.Name()
	g.Expect(err).ShouldNot(HaveOccurred())

	level, err := zapcore.ParseLevel("debug")
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
		l.WithOutAndErrPaths(name, name),
	)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		os.Remove(file.Name())
	})

	return log, name
}
