package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	ent "github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/agent"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/api"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/utilities/healthcheck"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewAPIServerCommand(log logr.Logger) *cobra.Command {
	var globalParams ParamSet

	var cmd = &cobra.Command{
		Use:   "gitops-repo-broker",
		Short: "HTTP server for playing w/ git",
		RunE: func(_ *cobra.Command, _ []string) error {
			return RunServer(log, globalParams)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVar(&globalParams.PrivKeyFile, "git-private-key-file", "", "Path to a SSH private key that is authorized for pull/push from/to the git repo specified by --git-url")
	cmd.Flags().StringVar(&globalParams.GitURL, "git-url", "", "Remote URL of the GitOps repository. Only the SSH protocol is supported. No HTTP/HTTPS.")
	cmd.Flags().StringVar(&globalParams.GitBranch, "git-branch", "master", "Branch that will be used by GitOps")
	cmd.Flags().StringVar(&globalParams.GitPath, "git-path", "/", "Subdirectory of the GitOps repository where configuration as code can be found.")
	cmd.Flags().DurationVar(&globalParams.HttpReadTimeout, "http-read-timeout", 30*time.Second, "ReadTimeout is the maximum duration for reading the entire request, including the body.")
	cmd.Flags().DurationVar(&globalParams.HttpWriteTimeout, "http-write-timeout", 30*time.Second, "WriteTimeout is the maximum duration before timing out writes of the response.")

	cmd.Flags().StringVar(&globalParams.AgentTemplateAlertmanagerURL, "agent-template-alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Value used to populate the alertmanager URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.AgentTemplateNatsURL, "agent-template-nats-url", "nats://nats-client.wkp-gitops-repo-broker:4222", "Value used to populate the nats URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.DbURI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
	cmd.Flags().StringVar(&globalParams.DbType, "db-type", os.Getenv("DB_TYPE"), "database type, supported types [sqlite, postgres]")
	cmd.Flags().StringVar(&globalParams.DbName, "db-name", os.Getenv("DB_NAME"), "database name, applicable if type is postgres")
	cmd.Flags().StringVar(&globalParams.DbUser, "db-user", os.Getenv("DB_USER"), "database user")
	cmd.Flags().StringVar(&globalParams.DbPassword, "db-password", os.Getenv("DB_PASSWORD"), "database password")
	cmd.Flags().StringVar(&globalParams.DbBusyTimeout, "db-busy-timeout", "5000", "How long should sqlite wait when trying to write to the database")
	cmd.Flags().StringVar(&globalParams.Port, "port", "8000", "Port to run http server on")
	cmd.Flags().StringVar(&globalParams.EntitlementSecretName, "entitlement-secret-name", ent.DefaultSecretName, "The name of the entitlement secret")
	cmd.Flags().StringVar(&globalParams.EntitlementSecretNamespace, "entitlement-secret-namespace", ent.DefaultSecretNamespace, "The namespace of the entitlement secret")

	return cmd
}

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
	EntitlementSecretName        string
	EntitlementSecretNamespace   string
}

func RunServer(log logr.Logger, params ParamSet) error {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
	}
	_ = schemeBuilder.AddToScheme(scheme)
	kubeClientConfig := config.GetConfigOrDie()
	kubeClient, err := client.New(kubeClientConfig, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	entitlementSecretKey := client.ObjectKey{Name: params.EntitlementSecretName, Namespace: params.EntitlementSecretNamespace}

	srv, err := NewServer(context.Background(), kubeClient, entitlementSecretKey, log, params)
	if err != nil {
		return err
	}
	log.Info("Server listening...")
	return srv.ListenAndServe()
}

func NewServer(ctx context.Context, c client.Client, entitlementSecretKey client.ObjectKey, log logr.Logger, params ParamSet) (*http.Server, error) {
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

	entitled := r.PathPrefix("/gitops/api").Subrouter()
	// entitlementMiddleware adds the entitlement in the request context so
	// it needs to be added before checkEntitlementMiddleware which reads from
	// the request context.
	entitled.Use(entitlementMiddleware(ctx, log, c, entitlementSecretKey))
	entitled.Use(checkEntitlementMiddleware(log))

	entitled.HandleFunc("/agent.yaml", agent.NewGetHandler(db, params.AgentTemplateNatsURL, params.AgentTemplateAlertmanagerURL)).Methods("GET")
	entitled.HandleFunc("/clusters", api.ListClusters(db, json.MarshalIndent)).Methods("GET")
	entitled.HandleFunc("/clusters/{id:[0-9]+}", api.FindCluster(db, json.MarshalIndent)).Methods("GET")
	entitled.HandleFunc("/clusters", api.RegisterCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent, utils.Generate)).Methods("POST")
	entitled.HandleFunc("/clusters/{id:[0-9]+}", api.UpdateCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent)).Methods("PUT")
	entitled.HandleFunc("/clusters/{id:[0-9]+}", api.UnregisterCluster(db)).Methods("DELETE")
	entitled.HandleFunc("/alerts", api.ListAlerts(db, json.MarshalIndent)).Methods("GET")

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

func entitlementMiddleware(ctx context.Context, log logr.Logger, c client.Client, key types.NamespacedName) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return entitlement.EntitlementHandler(ctx, log, c, key, next)
	}
}

func checkEntitlementMiddleware(log logr.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return entitlement.CheckEntitlementHandler(log, next)
	}
}
