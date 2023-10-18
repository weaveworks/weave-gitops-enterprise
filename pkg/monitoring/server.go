package monitoring

import (
	"fmt"

	"net/http"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/profiling"
)

type Options struct {
	Enabled          bool
	ServerAddress    string
	Log              logr.Logger
	MetricsOptions   metrics.Options
	ProfilingOptions profiling.Options
}

// NewSever creates and starts a management server for all endpoints that we need to expose internally. For example metrics or profiling.
func NewServer(opts Options) (*http.Server, error) {
	if !opts.Enabled {
		return nil, fmt.Errorf("cannot create disabled server")
	}
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
