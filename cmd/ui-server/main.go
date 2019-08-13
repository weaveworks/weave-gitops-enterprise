package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Serve the UI
	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", http.StripPrefix("/", fs))

	// TODO: Implement this as a part of https://github.com/weaveworks/wkp-ui/issues/175.
	http.HandleFunc("/api/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, `{"branches": ["master", "foo-branch"]}`)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}
