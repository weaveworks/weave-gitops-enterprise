package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	cleanerSubsystem = "objects_cleaner"

	// cleaner actions
	removeObjectsAction = "RemoveObjects"

	FailedLabel  = "error"
	SuccessLabel = "success"
)

var cleanerWatcher = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Subsystem: cleanerSubsystem,
	Name:      "status",
	Help:      "cleaner status",
}, []string{"status"})

var CleanerLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Subsystem: cleanerSubsystem,
	Name:      "latency_seconds",
	Help:      "cleaner latency",
	Buckets:   prometheus.LinearBuckets(0.01, 0.01, 10),
}, []string{"action", "status"})

var CleanerInflightRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Subsystem: cleanerSubsystem,
	Name:      "inflight_requests",
	Help:      "number of cleaner in-flight requests.",
}, []string{"action"})

func init() {
	prometheus.MustRegister(cleanerWatcher)
	prometheus.MustRegister(CleanerLatencyHistogram)
	prometheus.MustRegister(CleanerInflightRequests)
}

// CleanerWatcherDecrease decreases cleaner_watcher metric for status
func CleanerWatcherDecrease(status string) {
	cleanerWatcher.WithLabelValues(status).Dec()
}

// CleanerWatcherIncrease increases cleaner_watcher metric for status
func CleanerWatcherIncrease(status string) {
	cleanerWatcher.WithLabelValues(status).Inc()
}

func CleanerSetLatency(status string, duration time.Duration) {
	CleanerLatencyHistogram.WithLabelValues(removeObjectsAction, status).Observe(duration.Seconds())
}

func CleanerAddInflightRequests(number float64) {
	CleanerInflightRequests.WithLabelValues(removeObjectsAction).Add(number)
}
