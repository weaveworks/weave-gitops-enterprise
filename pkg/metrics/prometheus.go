package metrics

import (
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
	Enabled bool
}

func NewDefaultPrometheusHandler() (string, http.Handler) {
	return "/metrics", promhttp.HandlerFor(DefaultGatherers, promhttp.HandlerOpts{})
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
