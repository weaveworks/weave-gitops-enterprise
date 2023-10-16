package profiling

import (
	"net/http/pprof"

	"net/http"

	"github.com/go-logr/logr"
)

type Options struct {
	Enabled       bool
	ServerAddress string
	Log           logr.Logger
}

// NewPprofServer creates and starts a pprof server for profiling purposes. It is intended to be used in a different internal port than the main server
func NewPprofServer(opts Options) *http.Server {
	log := opts.Log.WithName("pprof-server")
	pprofMux := http.NewServeMux()
	pprofMux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
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

	return server
}
