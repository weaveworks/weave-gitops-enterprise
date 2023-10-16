package profiling

import (
	"net/http/pprof"

	"net/http"
)

type Options struct {
	Enabled bool
}

func NewDefaultPprofHandler() (string, http.Handler) {
	return "/debug/pprof/", http.HandlerFunc(pprof.Index)
}
