package profiling

import (
	"net/http/pprof"

	"net/http"
)

// Options structure to configure profiling behaviour. For example 'Enabled' acts a feature flag to control whether to enable profiling.
type Options struct {
	// Enabled controls whether profiling should be enabled
	Enabled bool
}

// NewDefaultPprofHandler creates a default http handler for profiling data using pprof https://pkg.go.dev/net/http/pprof.
// 'p' is the default path to expose the handler with value '/debug/pprof'
func NewDefaultPprofHandler() (p string, h http.Handler) {
	return "/debug/pprof/", http.HandlerFunc(pprof.Index)
}
