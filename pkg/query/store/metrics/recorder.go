package metrics

import (
	// "fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Recorder struct {
	storeLatencyHistogram *prometheus.HistogramVec
	inflightRequests      *prometheus.GaugeVec
}

const (
	// const status that is used in metrics.
	Failed  = "error"
	Success = "success"
)

var once sync.Once

// NewRecorder creates a new recorder and registers the Prometheus metrics
func NewRecorder(register bool, subsystem string) Recorder {
	storeLatencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "latency_seconds",
		Help:      "Store latency",
		Buckets:   prometheus.LinearBuckets(0.01, 0.01, 10),
	}, []string{"action", "status"})

	inflightRequests := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "inflight_requests_total",
		Help:      "Number of in-flight requests.",
	}, []string{"action"})

	once.Do(func() {
		prometheus.MustRegister(storeLatencyHistogram)
		prometheus.MustRegister(inflightRequests)
	})

	record := Recorder{
		storeLatencyHistogram: storeLatencyHistogram,
		inflightRequests:      inflightRequests,
	}

	return record
}

func (r Recorder) SetStoreLatency(action string, status string, duration time.Duration) {
	r.storeLatencyHistogram.WithLabelValues(action, status).Observe(duration.Seconds())
}

func (r Recorder) InflightRequests(action string, number float64) {
	r.inflightRequests.WithLabelValues(action).Add(number)
}
