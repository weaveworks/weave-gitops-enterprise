package server

import (
	"github.com/go-logr/logr"
	prom2 "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"net/http"
)

type MetricsServerConf struct {
	Enabled bool
	Address string
	Log     logr.Logger
}

// NewMetricsServer creates and starts a prometheus metrics server in /metrics path
// with the gatherers and configuration given as argument.
func NewMetricsServer(conf MetricsServerConf, gatherers prom2.Gatherers) *http.Server {
	log := conf.Log.WithName("metrics-server")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{}))
	metricsServer := &http.Server{
		Addr:    conf.Address,
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

// WithMetrics instruments http server with a prometheus metrics filter to generate
// golden signal metrics using https://github.com/slok/go-http-metrics
func WithMetrics(h http.Handler) http.Handler {
	recorder := prometheus.NewRecorder(prometheus.Config{})

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: recorder,
	})

	return std.Handler("", metricsMiddleware, h)
}
