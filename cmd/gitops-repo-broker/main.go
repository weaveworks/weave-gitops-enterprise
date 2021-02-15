package main

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/server"
)

var cmd = &cobra.Command{
	Use:   "gitops-repo-broker",
	Short: "HTTP server for playing w/ git",
	RunE: func(_ *cobra.Command, _ []string) error {
		return RunServer(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func RunServer(params server.ParamSet) error {
	srv, err := server.NewServer(context.Background(), params)
	if err != nil {
		return err
	}
	log.Info("Server listening...")
	return srv.ListenAndServe()
}

var globalParams server.ParamSet

func init() {
	cmd.Flags().StringVar(&globalParams.PrivKeyFile, "git-private-key-file", "", "Path to a SSH private key that is authorized for pull/push from/to the git repo specified by --git-url")
	cobra.MarkFlagRequired(cmd.Flags(), "private-key-file")

	cmd.Flags().StringVar(&globalParams.GitURL, "git-url", "", "Remote URL of the GitOps repository. Only the SSH protocol is supported. No HTTP/HTTPS.")
	cobra.MarkFlagRequired(cmd.Flags(), "git-url")

	cmd.Flags().StringVar(&globalParams.GitBranch, "git-branch", "master", "Branch that will be used by GitOps")
	cobra.MarkFlagRequired(cmd.Flags(), "git-branch")

	cmd.Flags().StringVar(&globalParams.GitPath, "git-path", "/", "Subdirectory of the GitOps repository where configuration as code can be found.")
	cmd.Flags().DurationVar(&globalParams.HttpReadTimeout, "http-read-timeout", 30*time.Second, "ReadTimeout is the maximum duration for reading the entire request, including the body.")
	cmd.Flags().DurationVar(&globalParams.HttpWriteTimeout, "http-write-timeout", 30*time.Second, "WriteTimeout is the maximum duration before timing out writes of the response.")

	cmd.Flags().StringVar(&globalParams.AgentTemplateAlertmanagerURL, "agent-template-alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Value used to populate the alertmanager URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.AgentTemplateNatsURL, "agent-template-nats-url", "nats://nats-client.wkp-mccp:4222", "Value used to populate the nats URL in /api/agent.yaml")
	cmd.Flags().StringVar(&globalParams.DbURI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
