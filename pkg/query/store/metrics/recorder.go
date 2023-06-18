package metrics

import (
	// "fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Recorder struct {
	storeLatencyHistogram *prometheus.HistogramVec
	requestCounter        *prometheus.CounterVec
	inflightRequests      *prometheus.GaugeVec
}

// NewRecorder creates a new recorder and registers the Prometheus metrics
func NewRecorder(register bool, subsystem string) Recorder {
	storeLatencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "latency_seconds",
		Help:      "Store latency",
		Buckets:   prometheus.LinearBuckets(0.001, 0.001, 10),
	}, []string{"action"})

	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Number of requests",
	}, []string{"action", "status"})

	inflightRequests := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "inflight_requests_total",
		Help:      "Number of in-flight requests.",
	}, []string{"action"})

	_ = prometheus.Register(storeLatencyHistogram)
	_ = prometheus.Register(requestCounter)
	_ = prometheus.Register(inflightRequests)

	record := Recorder{
		storeLatencyHistogram: storeLatencyHistogram,
		requestCounter:        requestCounter,
		inflightRequests:      inflightRequests,
	}

	return record
}

func (r Recorder) SetStoreLatency(action string, duration time.Duration) {
	r.storeLatencyHistogram.WithLabelValues(action).Observe(duration.Seconds())
}

func (r Recorder) IncRequestCounter(action string, status string) {
	r.requestCounter.WithLabelValues(action, status).Inc()
}

func (r Recorder) InflightRequests(action string, number float64) {
	r.inflightRequests.WithLabelValues(action).Add(number)
}
