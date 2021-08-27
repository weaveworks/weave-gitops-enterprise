package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/agent"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/api"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/utilities/healthcheck"
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
	DbName                       string
	DbUser                       string
	DbPassword                   string
	DbType                       string
	DbBusyTimeout                string
	Port                         string
}

func NewServer(ctx context.Context, params ParamSet) (*http.Server, error) {
	var err error
	uri := params.DbURI
	if params.DbType == "sqlite" {
		uri, err = utils.GetSqliteUri(params.DbURI, params.DbBusyTimeout)
		if err != nil {
			return nil, err
		}
	}

	started := time.Now()
	db, err := utils.Open(uri, params.DbType, params.DbName, params.DbUser, params.DbPassword)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	r.HandleFunc("/gitops/api/agent.yaml", agent.NewGetHandler(
		db, params.AgentTemplateNatsURL, params.AgentTemplateAlertmanagerURL)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.ListClusters(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.FindCluster(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.RegisterCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent, utils.Generate)).Methods("POST")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.UpdateCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent)).Methods("PUT")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.UnregisterCluster(db)).Methods("DELETE")
	r.HandleFunc("/gitops/api/alerts", api.ListAlerts(db, json.MarshalIndent)).Methods("GET")

	r.HandleFunc("/gitops/started", healthcheck.Started(started))
	r.HandleFunc("/gitops/healthz", healthcheck.Healthz(started))
	r.HandleFunc("/gitops/redirect", healthcheck.Redirect)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%s", params.Port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: params.HttpWriteTimeout,
		ReadTimeout:  params.HttpReadTimeout,
	}

	return srv, nil
}
