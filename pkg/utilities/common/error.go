package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type errorView struct {
	Error string `json:"message"`
}

// WriteError writes the error into the request
func WriteError(w http.ResponseWriter, appError error, code int) {
	log.Error(appError)
	payload, err := json.Marshal(errorView{Error: appError.Error()})
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(payload))
}
