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

	tests := []struct {
		name              string
		objects           []client.Object
		queryRequest      api.QueryRequest
		principal         string
		logLevel          string
		expectedEvents    []string
		nonExpectedEvents []string
	}{
		{
			name:              "can follow write path with debug level: happy path",
			objects:           []client.Object{},
			queryRequest:      api.QueryRequest{},
			logLevel:          "debug",
			principal:         "wego-admin",
			nonExpectedEvents: []string{},
			expectedEvents: []string{
				"objects collector started",
				"role collector started",
				"collectors started", //potential duplicate
				"watcher started",    //potential duplicate
				"watching cluster",
				"object transaction received",
				"storing object",
				"objects stored",
				"rolebinding stored",
				"role stored",
				"object transactions processed",
			},
		},
		{
			name:           "cannot follow write path without debug level",
			objects:        []client.Object{},
			queryRequest:   api.QueryRequest{},
			principal:      "wego-admin",
			logLevel:       "info",
			expectedEvents: []string{},
			nonExpectedEvents: []string{
				"objects collector started",
				"role collector started",
				"object transaction received",
				"storing object",
				"objects stored",
				"rolebinding stored",
				"role stored",
				"object transactions processed",
			},
		},
		{
			name:              "can follow read path with debug level: happy path",
			objects:           []client.Object{},
			principal:         "unauthorised-user",
			queryRequest:      api.QueryRequest{},
			logLevel:          "debug",
			nonExpectedEvents: []string{},
			expectedEvents: []string{
				"access checker created",
				"query service created",
				"query server created",
				"query received",
				"objects retrieved",
				"unauthorised access",
				"query processed",
			},
		},
		{
			name:           "cannot follow read path without debug level",
			objects:        []client.Object{},
			principal:      "unauthorised-user",
			queryRequest:   api.QueryRequest{},
			logLevel:       "info",
			expectedEvents: []string{},
			nonExpectedEvents: []string{
				"query received",
				"objects retrieved",
				"unauthorised access",
				"query processed",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			appLog, loggerPath := newLoggerWithLevel(t, tt.logLevel)
			// setup app server
			c, err := makeQueryServer(t, cfg, tt.principal, appLog, testLog)
			g.Expect(err).To(BeNil())

			//given a new event
			test.Create(ctx, t, cfg, tt.objects...)
			//when processed
			_, _ = c.DoQuery(context.Background(), &tt.queryRequest)
			g.Expect(err).To(BeNil())
			//g.Expect(len(query.Objects)).To(BeIdenticalTo(10))
			//then processing events are found
			g.Expect(assertLogs(loggerPath, tt.expectedEvents, tt.nonExpectedEvents)).To(Succeed())
		})
	}
}

func assertLogs(loggerPath string, expectedEvents []string, nonExpectedEvents []string) error {
	//get logs
	logs, err := os.ReadFile(loggerPath)
	if err != nil {
		return fmt.Errorf("cannot read logger file: %w", err)

	}
	logss := string(logs)
	logLines := strings.Split(logss, "\n")
	for _, event := range expectedEvents {
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

	for _, nonExpectedEvent := range nonExpectedEvents {
		if strings.Contains(logss, nonExpectedEvent) {
			return fmt.Errorf("non expected event found:%s", nonExpectedEvent)
		}
	}

	return nil
}

func newLoggerWithLevel(t *testing.T, logLevel string) (logr.Logger, string) {
	g := NewGomegaWithT(t)

	file, err := os.CreateTemp(os.TempDir(), "query-server-log")
	g.Expect(err).ShouldNot(HaveOccurred())

	name := file.Name()
	g.Expect(err).ShouldNot(HaveOccurred())

	level, err := zapcore.ParseLevel(logLevel)
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
		l.WithOutAndErrPaths(name, name),
	)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	return log, name
}
