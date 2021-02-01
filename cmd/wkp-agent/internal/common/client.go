package common

import (
	"context"
	"time"

	cloudeventsnats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

// CreateClient returns a cloudevents client
func CreateClient(ctx context.Context, NatsURL string, Subject string) cloudevents.Client {
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
	log.Infof("NATS host: %s", sender.Conn.Servers())

	client, err := cloudevents.NewClient(sender)
	if err != nil {
		log.Fatalf("Failed to create CloudEvents client, %s", err.Error())
	}

	return client
}
