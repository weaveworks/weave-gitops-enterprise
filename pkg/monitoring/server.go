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
func NewServer(opts Options, handlers map[string]http.Handler) (*http.Server, error) {
	if !opts.Enabled {
		return nil, fmt.Errorf("cannot create disabled server")
	}
	if opts.ServerAddress == "" {
		return nil, fmt.Errorf("cannot create server for empty address")
	}

	if handlers == nil || len(handlers) == 0 {
		return nil, fmt.Errorf("cannot create server without handlers")
	}

	log := opts.Log.WithName("management-server")
	pprofMux := http.NewServeMux()

	for path, handler := range handlers {
		pprofMux.Handle(path, handler)
		log.Info("added handler", "path", path)
	}

	server := &http.Server{
		Addr:    opts.ServerAddress,
		Handler: pprofMux,
	}

	go func() {
		log.Info("starting pprof server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Error(err, "could not start metrics server")
		}
	}()

	log.Info("server created")
	return server, nil

}
