package app

import (
	"context"
	"net/http"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "capi-server",
		Long: `The capi-server servers and handles REST operations for CAPI templates.
		CAPI templates are stored in the cluster as a ConfigMap indexed by their name.`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			StartServer()
			return nil
		},
	}
	return cmd
}


func NewServer(ctx context.Context) *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/templates", templates.List(context.Background())).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
	}

	return srv
}

func StartServer() {
	s := NewServer(context.Background())
	log.Info("Starting capi-server...")
	s.ListenAndServe()
}
