package healthcheck

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func Started(started time.Time) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		data := (time.Since(started)).String()
		_, _ = w.Write([]byte(data))
	}
}

func Healthz(started time.Time) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		data := (time.Since(started)).String()
		_, _ = w.Write([]byte(data))
		_, _ = w.Write([]byte("ok"))
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	loc, err := url.QueryUnescape(r.URL.Query().Get("loc"))
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid redirect: %q", r.URL.Query().Get("loc")), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, loc, http.StatusFound)
}
