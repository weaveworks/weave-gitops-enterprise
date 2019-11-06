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
	Use:   "gitops-repo-broker",
	Short: "HTTP server for playing w/ git",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runServer(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	privKeyFile      string
	gitURL           string
	gitBranch        string
	gitPath          string
	httpReadTimeout  time.Duration
	httpWriteTimeout time.Duration
}

var globalParams paramSet

func init() {
	cmd.Flags().StringVar(&globalParams.privKeyFile, "git-private-key-file", "", "Path to a SSH private key that is authorized for pull/push from/to the git repo specified by --git-url")
	cobra.MarkFlagRequired(cmd.Flags(), "private-key-file")

	cmd.Flags().StringVar(&globalParams.gitURL, "git-url", "", "Remote URL of the GitOps repository. Only the SSH protocol is supported. No HTTP/HTTPS.")
	cobra.MarkFlagRequired(cmd.Flags(), "git-url")

	cmd.Flags().StringVar(&globalParams.gitBranch, "git-branch", "master", "Branch that will be used by GitOps")
	cobra.MarkFlagRequired(cmd.Flags(), "git-branch")

	cmd.Flags().StringVar(&globalParams.gitPath, "git-path", "/", "Subdirectory of the GitOps repository where configuration as code can be found.")
	cmd.Flags().DurationVar(&globalParams.httpReadTimeout, "http-read-timeout", 30*time.Second, "ReadTimeout is the maximum duration for reading the entire request, including the body.")
	cmd.Flags().DurationVar(&globalParams.httpWriteTimeout, "http-write-timeout", 30*time.Second, "WriteTimeout is the maximum duration before timing out writes of the response.")
}

func runServer(params paramSet) error {
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
		WriteTimeout: params.httpWriteTimeout,
		ReadTimeout:  params.httpReadTimeout,
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
