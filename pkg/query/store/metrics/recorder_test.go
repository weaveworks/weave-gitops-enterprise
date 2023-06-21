package metrics

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/metrics"
)

func TestRecorder(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)

	// Create a new Recorder instance
	subsystem := "test"
	recorder := NewRecorder(false, subsystem)

	metrics.NewPrometheusServer(metrics.Options{
		ServerAddress: "localhost:8080",
	}, prometheus.Gatherers{
		prometheus.DefaultGatherer,
	})

	t.Run("can set inflight requests", func(t *testing.T) {
		// Set inflight requests
		recorder.SetStoreLatency("add", Success, time.Duration(time.Duration.Seconds(1)))
		recorder.InflightRequests("add", 1)

		// Retrieve the metrics
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
		g.Expect(err).NotTo(HaveOccurred())
		resp, err := http.DefaultClient.Do(req)
		g.Expect(err).NotTo(HaveOccurred())
		b, err := io.ReadAll(resp.Body)
		g.Expect(err).NotTo(HaveOccurred())
		metrics := string(b)
		log.Info("metrics: %s", metrics)

		expMetrics := []string{
			`test_inflight_requests_total{action="add"}`,
			`test_latency_seconds_bucket{action="add"`,
		}

		for _, expMetric := range expMetrics {
			//Contains expected value
			g.Expect(metrics).To(ContainSubstring(expMetric))
		}
	})

}
