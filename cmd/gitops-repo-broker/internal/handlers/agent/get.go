package agent

import (
	"fmt"
	"net/http"
)

// Get returns a YAMLStream given a token
func Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")

	token := r.URL.Query().Get("token")
	stream := token

	if token != "" {
		w.Write([]byte(stream))
	} else {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(stream))
}
