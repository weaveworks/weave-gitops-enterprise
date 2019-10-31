package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/permissions"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/workspaces"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/gitops/workspaces", workspaces.List).Methods("GET")
	r.HandleFunc("/gitops/workspaces", workspaces.Create).Methods("POST")

	r.HandleFunc("/gitops/permissions", permissions.Create).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
	}

	logrus.Info("Server listening...")
	logrus.Fatal(srv.ListenAndServe())
}
