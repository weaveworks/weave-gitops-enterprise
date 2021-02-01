package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/wkp-agent/internal/common"
	clusterclient "github.com/weaveworks/wks/pkg/cluster/client"
	clusterwatcher "github.com/weaveworks/wks/pkg/cluster/watcher"
	"github.com/weaveworks/wks/pkg/messaging/handlers"
	"k8s.io/client-go/informers"
)

var (
	Subject string
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches for Kubernetes events and publishes them to NATS",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.PersistentFlags().StringVar(&Subject, "subject", "weave.wkp.agent.events", "NATS subject to send Kubernetes events to")
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k8sClient, err := clusterclient.GetClient(KubeconfigFile)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %s.", err.Error())
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
	factory := informers.NewSharedInformerFactory(k8sClient, 0)

	client := common.CreateClient(ctx, NatsURL, Subject)

	notifier, err := handlers.NewEventNotifier("wkp-agent", client)
	if err != nil {
		log.Fatalf("Failed to create event notifier, %s.", err.Error())
	}

	events := clusterwatcher.NewWatcher(factory.Core().V1().Events().Informer(), notifier.Notify)

	log.Info("Agent starting watchers.")
	go events.Run("Events", 1, ctx.Done())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
