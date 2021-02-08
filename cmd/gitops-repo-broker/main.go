package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/agent"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/api"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/branches"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/clusters"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/clusters/upgrades"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/clusters/version"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/workspaces"
	"github.com/weaveworks/wks/common/database/utils"
	"github.com/weaveworks/wks/pkg/utilities/healthcheck"
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
	privKeyFile                  string
	gitURL                       string
	gitBranch                    string
	gitPath                      string
	httpReadTimeout              time.Duration
	httpWriteTimeout             time.Duration
	agentTemplateNatsURL         string
	agentTemplateAlertmanagerURL string
	dbURI                        string
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

	cmd.Flags().StringVar(&globalParams.agentTemplateAlertmanagerURL, "--agent-template-alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Value used to populate the alertmanager URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.agentTemplateNatsURL, "--agent-template-nats-url", "nats://nats-client.wkp-mccp:4222", "Value used to populate the nats URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.dbURI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
}

func runServer(params paramSet) error {
	privKey, err := ioutil.ReadFile(params.privKeyFile)
	if err != nil {
		return err
	}

	started := time.Now()
	db, err := utils.Open(params.dbURI)
	if err != nil {
		return err
	}
	// Get hold of internal *sql.DB and ensure it's closed at the end
	sqldb, err := db.DB()
	if err != nil {
		return err
	}
	defer sqldb.Close()

	r := mux.NewRouter()

	// These endpoints assume WKS single cluster (no multi-cluster support)
	r.HandleFunc("/gitops/cluster/upgrades", upgrades.List).Methods("GET")
	r.HandleFunc("/gitops/cluster/version", version.Get(params.gitURL, params.gitBranch, privKey)).Methods("GET")
	r.HandleFunc("/gitops/cluster/version", version.Update(params.gitURL, params.gitBranch, privKey)).Methods("PUT")

	// These endpoints assume EKSCluster CRDs being present in git
	r.HandleFunc("/gitops/clusters/{namespace}/{name}", clusters.Get).Methods("GET")
	r.HandleFunc("/gitops/clusters/{namespace}/{name}", clusters.Update(params.gitURL, params.gitBranch, privKey)).Methods("POST")
	r.HandleFunc("/gitops/clusters", clusters.List).Methods("GET")

	r.HandleFunc("/gitops/repo/branches", branches.List(params.gitURL, params.privKeyFile)).Methods("GET")

	r.HandleFunc("/gitops/workspaces", workspaces.List).Methods("GET")
	r.HandleFunc("/gitops/workspaces", workspaces.MakeCreateHandler(
		params.gitURL, params.gitBranch, privKey, params.gitPath)).Methods("POST")

	r.HandleFunc("/gitops/api/agent.yaml", agent.NewGetHandler(
		params.agentTemplateNatsURL, params.agentTemplateAlertmanagerURL)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.NewGetClusters(db, json.MarshalIndent)).Methods("GET")

	r.HandleFunc("/gitops/started", healthcheck.Started(started))
	r.HandleFunc("/gitops/healthz", healthcheck.Healthz(started))
	r.HandleFunc("/gitops/redirect", healthcheck.Redirect)

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
