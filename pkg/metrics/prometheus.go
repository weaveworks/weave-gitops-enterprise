package metrics

import (
	prom2 "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"net/http"
)

// NewPrometheusServer creates and starts a prometheus metrics server in /metrics path
// with the gatherers and configuration given as argument.
func NewPrometheusServer(opts Options, gatherers prom2.Gatherers) *http.Server {
	log := opts.Log.WithName("metrics-server")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{}))
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
