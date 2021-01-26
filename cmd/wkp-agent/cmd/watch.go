package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	cloudeventsnats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	options := cloudeventsnats.NatsOptions(
		nats.MaxReconnects(-1), // Always reconnect
		nats.ReconnectWait(5*time.Second),
		nats.ErrorHandler(func(con *nats.Conn, sub *nats.Subscription, err error) {
			log.Debugf("Agent encountered an error: %v.", err)
		}),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Debugf("Agent disconnected from broker: %v.", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Debugf("Agent reconnected to broker.")
		}),
	)
	sender, err := cloudeventsnats.NewSender(NatsURL, Subject, options)
	if err != nil {
		log.Fatalf("Failed to create NATS client, %s.", err.Error())
	}
	defer sender.Close(ctx)
	log.Infof("NATS host: %s", sender.Conn.Servers())

	client, err := cloudevents.NewClient(sender)
	if err != nil {
		log.Fatalf("Failed to create CloudEvents client, %s", err.Error())
	}

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
