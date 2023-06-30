package metrics

import (
	"io"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/metrics"
)

func TestRecorder(t *testing.T) {
	g := NewWithT(t)

	metrics.NewPrometheusServer(metrics.Options{
		ServerAddress: "localhost:8080",
	})

	t.Run("can update cluster watcher metrics", func(t *testing.T) {
		//when metrics recorded
		ClusterWatcherIncrease("starting")
		expMetrics := []string{
			`# HELP collector_cluster_watcher number of active cluster watchers by watcher status`,
			`# TYPE collector_cluster_watcher gauge`,
			`collector_cluster_watcher{status="starting"} 1`,
		}
		assertMetrics(g, expMetrics)

		ClusterWatcherDecrease("starting")
		assertMetrics(g, []string{
			`collector_cluster_watcher{status="starting"} 0`,
		})

	})

}

func assertMetrics(g *WithT, expMetrics []string) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
	g.Expect(err).NotTo(HaveOccurred())
	resp, err := http.DefaultClient.Do(req)
	g.Expect(err).NotTo(HaveOccurred())
	b, err := io.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())
	metrics := string(b)

	for _, expMetric := range expMetrics {
		//Contains expected value
		g.Expect(metrics).To(ContainSubstring(expMetric))
	}
}
