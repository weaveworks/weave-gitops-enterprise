package subscribe

import (
	"context"
	"fmt"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/cmd/event-writer/database/utils"
	v1 "k8s.io/api/core/v1"
)

// ToSubject subscribes to a subject given a nats connection
func ToSubject(server string, subject string, fn interface{}) error {
	// Channel Subscriber
	ctx := context.Background()
	log.Info("before new consumer")
	p, err := cenats.NewConsumer(server, subject, cenats.NatsOptions())
	if err != nil {
		log.Fatalf("failed to create nats protocol, %s", err.Error())
	}

	defer p.Close(ctx)

	log.Info(fmt.Sprintf("before new client, subject: %s, server: %s", subject, server))
	log.Info(fmt.Sprintf("nats servers: %v,", p.Conn.Servers()))
	c, err := ce.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	for {
		log.Info("before start receiver")
		if err := c.StartReceiver(ctx, fn); err != nil {
			log.Printf("failed to start nats receiver, %s", err.Error())
		}
	}
}

// ReceiveEvent can be passed as a callback function in ToSubject to store the received events to the DB
func ReceiveEvent(ctx context.Context, event ce.Event) error {
	fmt.Printf("event context: %+v\n", event.Context)

	data := &v1.Event{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("failed to parse event: %s\n", err.Error())
	}
	fmt.Printf("received event: %+v\n", data)

	dbEvent, err := converter.ConvertEvent(*data)
	if err != nil {
		fmt.Printf("failed to convert event to db event: %s\n", err.Error())
	}

	utils.DB.Create(&dbEvent)
	return nil
}
