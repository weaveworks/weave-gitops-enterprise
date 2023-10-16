package metrics

import (
	"github.com/go-logr/logr"
	prom "github.com/prometheus/client_golang/prometheus"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

// DefaultGatherers are the prometheus gatherers to serve metrics from
var DefaultGatherers = prom.Gatherers{
	prom.DefaultGatherer,
	clustersmngr.Registry,
}

type Options struct {
	Enabled       bool
	ServerAddress string
	Log           logr.Logger
}

// NewPrometheusServer creates and starts a prometheus metrics server in /metrics path
// with the gatherers and configuration given as argument.
func NewPrometheusServer(opts Options) *http.Server {
	log := opts.Log.WithName("metrics-server")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(DefaultGatherers, promhttp.HandlerOpts{}))
	metricsServer := &http.Server{
		Addr:    opts.ServerAddress,
		Handler: metricsMux,
	}

	go func() {
		log.Info("starting metrics server", "address", metricsServer.Addr)
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Error(err, "could not start metrics server")
		}
	}()

	return metricsServer
}

// WithHttpMetrics instruments http server with a prometheus metrics filter to generate
// golden signal metrics using https://github.com/slok/go-http-metrics
func WithHttpMetrics(h http.Handler) http.Handler {
	recorder := prometheus.NewRecorder(prometheus.Config{})

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: recorder,
	})

	return std.Handler("", metricsMiddleware, h)
}
