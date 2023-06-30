package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const collectorSubsystem = "collector"

var clusterWatcher = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Subsystem: collectorSubsystem,
	Name:      "cluster_watcher",
	Help:      "number of active cluster watchers by watcher status",
}, []string{"status"})

func init() {
	prometheus.MustRegister(clusterWatcher)
}

// ClusterWatcherDecrease decreases collector_cluster_watcher for status
func ClusterWatcherDecrease(status string) {
	clusterWatcher.WithLabelValues(status).Dec()
}

// ClusterWatcherIncrease increases collector_cluster_watcher metric for status
func ClusterWatcherIncrease(status string) {
	clusterWatcher.WithLabelValues(status).Inc()
}
