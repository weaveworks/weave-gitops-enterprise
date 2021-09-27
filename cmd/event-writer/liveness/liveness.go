package liveness

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
)

// StartLivenessProcess starts an http server to respond to the k8s liveness probes
func StartLivenessProcess() {
	started := time.Now()
	http.HandleFunc("/started", startedHandler(started))
	http.HandleFunc("/healthz", healthzHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func isDBReady() bool {
	if utils.DB != nil && utils.HasAllTables(utils.DB) {
		return true
	}
	return false
}

func startedHandler(started time.Time) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		data := (time.Since(started)).String()
		_, err := w.Write([]byte(data))
		if err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}

func healthzHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if isDBReady() {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("database is not ready")))
		}
	}
}
