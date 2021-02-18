package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
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

type ParamSet struct {
	PrivKeyFile                  string
	GitURL                       string
	GitBranch                    string
	GitPath                      string
	HttpReadTimeout              time.Duration
	HttpWriteTimeout             time.Duration
	AgentTemplateNatsURL         string
	AgentTemplateAlertmanagerURL string
	DbURI                        string
}

func NewServer(ctx context.Context, params ParamSet) (*http.Server, error) {
	privKey, err := ioutil.ReadFile(params.PrivKeyFile)
	if err != nil {
		return nil, err
	}

	started := time.Now()
	db, err := utils.Open(params.DbURI)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	// These endpoints assume WKS single cluster (no multi-cluster support)
	r.HandleFunc("/gitops/cluster/upgrades", upgrades.List).Methods("GET")
	r.HandleFunc("/gitops/cluster/version", version.Get(params.GitURL, params.GitBranch, privKey)).Methods("GET")
	r.HandleFunc("/gitops/cluster/version", version.Update(params.GitURL, params.GitBranch, privKey)).Methods("PUT")

	// These endpoints assume EKSCluster CRDs being present in git
	r.HandleFunc("/gitops/clusters/{namespace}/{name}", clusters.Get).Methods("GET")
	r.HandleFunc("/gitops/clusters/{namespace}/{name}", clusters.Update(params.GitURL, params.GitBranch, privKey)).Methods("POST")
	r.HandleFunc("/gitops/clusters", clusters.List).Methods("GET")

	r.HandleFunc("/gitops/repo/branches", branches.List(ctx, params.GitURL, params.PrivKeyFile)).Methods("GET")

	r.HandleFunc("/gitops/workspaces", workspaces.List).Methods("GET")
	r.HandleFunc("/gitops/workspaces", workspaces.MakeCreateHandler(
		params.GitURL, params.GitBranch, privKey, params.GitPath)).Methods("POST")

	r.HandleFunc("/gitops/api/agent.yaml", agent.NewGetHandler(
		db, params.AgentTemplateNatsURL, params.AgentTemplateAlertmanagerURL)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.ListClusters(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.FindCluster(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.RegisterCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent, api.Generate)).Methods("POST")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.UpdateCluster(db, json.Unmarshal, json.MarshalIndent)).Methods("PUT")

	r.HandleFunc("/gitops/started", healthcheck.Started(started))
	r.HandleFunc("/gitops/healthz", healthcheck.Healthz(started))
	r.HandleFunc("/gitops/redirect", healthcheck.Redirect)

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: params.HttpWriteTimeout,
		ReadTimeout:  params.HttpReadTimeout,
	}

	return srv, nil
}
