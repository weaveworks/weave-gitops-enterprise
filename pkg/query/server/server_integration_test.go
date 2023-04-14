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
	"io"
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

	appLog, err := l.New("debug", false)

	//appLog, r, w := createDebugLogger(t)
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
			//g.Expect(assertLogs(t, r, w, tt.expectedEvents)).To(Succeed())
		})
	}
}

func assertLogs(t *testing.T, r *os.File, w *os.File, events []string) error {
	logs := getLogs(t, r, w)
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

func createDebugLogger(t *testing.T) (logr.Logger, *os.File, *os.File) {
	g := NewGomegaWithT(t)

	level, err := zapcore.ParseLevel("debug")
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
	)

	r, w := redirectStdout(t)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	return log, r, w

}

func redirectStdout(t *testing.T) (*os.File, *os.File) {
	g := NewGomegaWithT(t)
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	g.Expect(err).NotTo(HaveOccurred())

	os.Stdout = w

	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	return r, w
}

func getLogs(t *testing.T, r, w *os.File) []byte {
	g := NewGomegaWithT(t)
	t.Helper()

	w.Close()

	out, err := io.ReadAll(r)
	g.Expect(err).NotTo(HaveOccurred())

	return out
}
