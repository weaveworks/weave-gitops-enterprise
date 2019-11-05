package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/permissions"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/workspaces"
)

var cmd = &cobra.Command{
	Use:           "gitops-repo-broker",
	Short:         "HTTP server for playing w/ git",
	RunE:          runServer,
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	privKeyFile string
	gitURL      string
	gitBranch   string
	gitPath     string
}

var params paramSet

func init() {
	cmd.Flags().StringVar(&params.privKeyFile, "git-private-key-file", "", "Path to a SSH private key that is authorized for pull/push from/to the git repo specified by --git-url")
	cobra.MarkFlagRequired(cmd.Flags(), "private-key-file")

	cmd.Flags().StringVar(&params.gitURL, "git-url", "", "Remote URL of the GitOps repository. Only the SSH protocol is supported. No HTTP/HTTPS.")
	cobra.MarkFlagRequired(cmd.Flags(), "git-url")

	cmd.Flags().StringVar(&params.gitBranch, "git-branch", "master", "Branch that will be used by GitOps")
	cobra.MarkFlagRequired(cmd.Flags(), "git-branch")

	cmd.Flags().StringVar(&params.gitPath, "git-path", "/", "Subdirectory of the GitOps repository where configuration as code can be found.")
}

func runServer(cmd *cobra.Command, args []string) error {
	privKey, err := ioutil.ReadFile(params.privKeyFile)
	if err != nil {
		return nil
	}

	r := mux.NewRouter()
	r.HandleFunc("/gitops/workspaces", workspaces.MakeListHandler(
		params.gitURL,
		params.gitBranch,
		privKey,
		params.gitPath,
	)).Methods("GET")
	r.HandleFunc("/gitops/workspaces", workspaces.Create).Methods("POST")

	r.HandleFunc("/gitops/permissions", permissions.Create).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	logrus.Info("Server listening...")
	logrus.Fatal(srv.ListenAndServe())

	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
