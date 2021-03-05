package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/wkp-agent/internal/common"
	"github.com/weaveworks/wks/common/messaging/handlers"
	clusterclient "github.com/weaveworks/wks/pkg/cluster/client"
	clusterpoller "github.com/weaveworks/wks/pkg/cluster/poller"
	clusterwatcher "github.com/weaveworks/wks/pkg/cluster/watcher"
	"github.com/weaveworks/wks/pkg/utilities/healthcheck"
	"k8s.io/client-go/informers"
)

const (
	WKPAgentEventSource string = "wkp-agent"
)

type watchCmdParamSet struct {
	ClusterInfoPollingInterval   time.Duration
	FluxInfoPollingInterval      time.Duration
	GitCommitInfoPollingInterval time.Duration
	WorkspaceInfoPollingInterval time.Duration
	HealthCheckPort              int
}

var watchCmdParams watchCmdParamSet

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches for Kubernetes events and publishes them to NATS",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.PersistentFlags().DurationVar(&watchCmdParams.ClusterInfoPollingInterval, "cluster-info-polling-interval", 10*time.Second, "Polling interval for ClusterInfo")
	watchCmd.PersistentFlags().DurationVar(&watchCmdParams.FluxInfoPollingInterval, "flux-info-polling-interval", 10*time.Second, "Polling interval for flux deployment info")
	watchCmd.PersistentFlags().DurationVar(&watchCmdParams.GitCommitInfoPollingInterval, "git-commit-info-polling-interval", 10*time.Second, "Polling interval for git commit info")
	watchCmd.PersistentFlags().DurationVar(&watchCmdParams.WorkspaceInfoPollingInterval, "workspace-info-polling-interval", 10*time.Second, "Polling interval for workspace info")
	watchCmd.PersistentFlags().IntVar(&watchCmdParams.HealthCheckPort, "health-check-port", 8080, "Port to expose health check")
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	token := os.Getenv(WKPAgentTokenEnvVar)
	if token == "" {
		log.Fatalf("The `%s` environment variable has not been set.  Please set it and try again.", WKPAgentTokenEnvVar)
	}

	typedClient, err := clusterclient.GetTypedClient(KubeconfigFile)
	if err != nil {
		log.Fatalf("Failed to create typed Kubernetes client: %s.", err.Error())
	}
	dynamicClient, err := clusterclient.GetDynamicClient(KubeconfigFile)
	if err != nil {
		log.Fatalf("Failed to create dynamic Kubernetes client: %s.", err.Error())
	}
	// The defaultResync duration is used as a default value for any handlers
	// added via AddEventHandler that don't specify a default value. This value
	// controls how often to re-list the resource that we want to be informed of.
	// Re-listing the resource results in re-fetching current items from the API
	// server and adding them to the store, resulting in Added events for new items,
	// Deleted events for removed items and Updated events for all existing items.
	// For our use case, if we set this value to anything other than 0, it will
	// fire OnUpdate calls for existing items in the store even if nothing has changed.
	// So we want to delay this as much as possible by setting it to 0.
	factory := informers.NewSharedInformerFactory(typedClient, 0)

	client := common.CreateClient(ctx, NatsURL, Subject)

	log.Info("Agent starting watchers.")

	// Watch for Events
	notifier := handlers.NewEventNotifier(token, WKPAgentEventSource, client)
	events := clusterwatcher.NewWatcher(factory.Core().V1().Events().Informer(), notifier.Notify)
	go events.Run("Events", 1, ctx.Done())

	// Poll for ClusterInfo
	clusterInfoSender := handlers.NewClusterInfoSender(WKPAgentEventSource, client)
	clusterInfo := clusterpoller.NewClusterInfoPoller(token, typedClient, watchCmdParams.ClusterInfoPollingInterval, clusterInfoSender)
	go clusterInfo.Run(ctx.Done())

	// Poll for FluxInfo
	fluxInfoSender := handlers.NewFluxInfoSender(WKPAgentEventSource, client)
	fluxInfo := clusterpoller.NewFluxInfoPoller(token, typedClient, watchCmdParams.FluxInfoPollingInterval, fluxInfoSender)
	go fluxInfo.Run(ctx.Done())

	// Poll for latest Git commit info
	gitCommitInfoSender := handlers.NewGitCommitInfoSender(WKPAgentEventSource, client)
	gitCommitInfo := clusterpoller.NewGitCommitInfoPoller(token, typedClient, watchCmdParams.GitCommitInfoPollingInterval, resty.New(), gitCommitInfoSender)
	go gitCommitInfo.Run(ctx.Done())

	workspaceInfoSender := handlers.NewWorkspaceInfoSender(WKPAgentEventSource, client)
	workspaceInfo := clusterpoller.NewWorkspaceInfoPoller(token, dynamicClient, watchCmdParams.WorkspaceInfoPollingInterval, workspaceInfoSender)
	go workspaceInfo.Run(ctx.Done())

	livenessCheck()
}

func livenessCheck() {
	started := time.Now()
	http.HandleFunc("/healthz", healthcheck.Healthz(started))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", watchCmdParams.HealthCheckPort), nil))
}
