//go:build integration
// +build integration

package objectscollector_test

import (
	"fmt"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	l "github.com/weaveworks/weave-gitops/core/logger"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/internal/entesting"
	"github.com/weaveworks/weave-gitops-enterprise/test"
)

func TestCollector_IntegrationTest(t *testing.T) {
	//ctx := context.Background()
	debugLogger, r, w := createDebugLogger(t)
	k8sEnv := test.StartTestEnv(t)
	g := NewGomegaWithT(t)

	//setup data
	clusterName := "anyCluster"
	_, _, _ = entesting.MakeGRPCServer(t, k8sEnv.Rest, k8sEnv, debugLogger)

	tests := []struct {
		name           string
		objects        []client.Object
		expectedEvents []string
	}{
		{
			name: "can trace new helm release object",
			objects: []client.Object{
				testutils.NewHelmRelease("createdOrUpdatedHelmRelease", clusterName),
			},
			expectedEvents: []string{"debug message"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given a new event
			//test.Create(ctx, t, k8sEnv.Rest, tt.objects...)
			//when processed

			//then processing events are found
			g.Expect(assertLogs(t, r, w, tt.expectedEvents)).To(Succeed())
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
