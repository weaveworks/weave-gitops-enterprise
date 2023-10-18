package monitoring

import (
	"fmt"

	"net/http"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/profiling"
)

// Options configuration options for the monitoring server
type Options struct {
	// Enabled controls whether monitoring server should be enabled
	Enabled bool
	// ServerAddress indicate the monitoring server binding address
	ServerAddress string
	// Log upstream logger to use for monitoring events
	Log logr.Logger
	// MetricsOptions configuration options for metrics endpoints
	MetricsOptions metrics.Options
	// ProfilingOptions configuration options for profiling
	ProfilingOptions profiling.Options
}

// NewSever creates and starts a management server for all endpoints that we need to expose internally. For example metrics or profiling.
func NewServer(opts Options) (*http.Server, error) {
	if opts.ServerAddress == "" {
		return nil, fmt.Errorf("cannot create server for empty address")
	}

	log := opts.Log.WithName("monitoring-server")
	pprofMux := http.NewServeMux()

	if opts.MetricsOptions.Enabled {
		metricsPath, metricsHandler := metrics.NewDefaultPrometheusHandler()
		pprofMux.Handle(metricsPath, metricsHandler)
		log.Info("added metrics handler", "path", metricsPath)
	}

	if opts.ProfilingOptions.Enabled {
		pprofPath, pprofHandler := profiling.NewDefaultPprofHandler()
		pprofMux.Handle(pprofPath, pprofHandler)
		log.Info("added profiling handler", "path", pprofPath)
	}

	server := &http.Server{
		Addr:    opts.ServerAddress,
		Handler: pprofMux,
	}

	go func() {
		log.Info("starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Error(err, "could not start metrics server")
		}
	}()

	return server, nil
}
