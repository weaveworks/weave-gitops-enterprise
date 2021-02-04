package cmd

import (
	"context"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/wkp-agent/internal/common"
	"github.com/weaveworks/wks/cmd/wkp-agent/server/handlers/alertmanager"
	"k8s.io/apimachinery/pkg/util/wait"
)

var cmd = &cobra.Command{
	Use:   "agent-server",
	Short: "HTTP server for WKP agent",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runServer(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	alertmanagerURL          string
	alertmanagerPollInterval time.Duration
	listenAddress            string
	httpReadTimeout          time.Duration
	httpWriteTimeout         time.Duration
}

var globalParams paramSet

func init() {
	rootCmd.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&Subject, "subject", "weave.wkp.agent.events", "NATS subject to send Kubernetes events to")

	cmd.Flags().StringVar(&globalParams.listenAddress, "listen-address", "0.0.0.0:8000", "Address to listen for webhook requests.")
	cmd.Flags().StringVar(&globalParams.alertmanagerURL, "alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Address of Alertmanager HTTP API")
	cmd.Flags().DurationVar(&globalParams.alertmanagerPollInterval, "alertmanager-poll-interval", 15*time.Second, "How often to poll alertmanager for alerts")

	cmd.Flags().DurationVar(&globalParams.httpReadTimeout, "http-read-timeout", 30*time.Second, "ReadTimeout is the maximum duration for reading the entire request, including the body.")
	cmd.Flags().DurationVar(&globalParams.httpWriteTimeout, "http-write-timeout", 30*time.Second, "WriteTimeout is the maximum duration before timing out writes of the response.")
}

func runServer(params paramSet) error {
	r := mux.NewRouter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := common.CreateClient(ctx, NatsURL, Subject)

	log.Infof("Starting to poll Alertmanager at %v", params.alertmanagerURL)
	go wait.Until(func() {
		ce, err := alertmanager.GetAlertsAsEvent(params.alertmanagerURL)
		if err != nil {
			log.Errorf("Failed to ping Alertmanager at %v: %v", params.alertmanagerURL, err)
			return
		}
		if result := client.Send(ctx, ce); cloudevents.IsUndelivered(result) {
			log.Fatalf("failed to send, %v", result)
		}
	}, params.alertmanagerPollInterval, ctx.Done())

	r.HandleFunc("/api/alertmanager_webhook", alertmanager.NewWebhookHandler(func(ce event.Event) {
		if result := client.Send(ctx, ce); cloudevents.IsUndelivered(result) {
			log.Fatalf("failed to send, %v", result)
		}
	})).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    params.listenAddress,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: params.httpWriteTimeout,
		ReadTimeout:  params.httpReadTimeout,
	}

	log.Info("Server listening...")
	log.Fatal(srv.ListenAndServe())

	return nil
}
