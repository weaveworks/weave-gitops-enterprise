package common

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type errorView struct {
	Error string `json:"message"`
}

// WriteError writes the error into the request
func WriteError(w http.ResponseWriter, appError error, code int) {
	log.Error(appError)
	json, err := json.Marshal(errorView{Error: appError.Error()})
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
	}
	http.Error(w, string(json), code)
}
