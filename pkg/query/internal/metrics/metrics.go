package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const queryServiceSubSystem = "query"

const RunQueryAction = "RunQuery"

const (
	FailedLabel  = "error"
	SuccessLabel = "success"
)

var QueryServiceLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Subsystem: queryServiceSubSystem,
	Name:      "latency_seconds",
	Help:      "query service latency",
	Buckets:   prometheus.LinearBuckets(0.01, 0.01, 10),
}, []string{"action", "status"})

func QueryServiceSetLatency(action string, status string, duration time.Duration) {
	QueryServiceLatencyHistogram.WithLabelValues(action, status).Observe(duration.Seconds())
}

func init() {
	prometheus.MustRegister(QueryServiceLatencyHistogram)
}

func AssertMetrics(t *testing.T, ts *httptest.Server, expMetrics []string) {
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}

	metrics := string(b)

	for _, expMetric := range expMetrics {
		if !strings.Contains(metrics, expMetric) {
			t.Errorf("Expected metric not found: %s", expMetric)
		}
	}
}
