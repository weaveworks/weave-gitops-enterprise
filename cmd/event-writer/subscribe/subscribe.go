package subscribe

import (
	"context"
	"fmt"
	"time"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"

	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/cmd/event-writer/queue"
	"github.com/weaveworks/wks/common/database/utils"
	v1 "k8s.io/api/core/v1"
)

// ToSubject subscribes to a subject given a nats connection
func ToSubject(server string, subject string, fn interface{}) error {
	// Channel Subscriber
	ctx := context.Background()
	log.Debug("creating new cloudevents NATS consumer")
	p, err := cenats.NewConsumer(server, subject, cenats.NatsOptions())
	if err != nil {
		log.Fatalf("failed to create nats protocol, %s", err.Error())
	}

	defer p.Close(ctx)

	log.Debug(fmt.Sprintf("creating client for NATS server: %v,", p.Conn.Servers()))
	c, err := ce.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	for {
		log.Debug("starting NATS receiver")
		if err := c.StartReceiver(ctx, fn); err != nil {
			log.Printf("failed to start nats receiver, %s", err.Error())
		}
	}
}

// ReceiveEvent can be passed as a callback function in ToSubject to store the received events to the DB
func ReceiveEvent(ctx context.Context, event ce.Event) error {
	data := &v1.Event{}
	if err := event.DataAs(data); err != nil {
		log.Warn(fmt.Sprintf("failed to parse event: %s\n", err.Error()))
	}
	log.Info(fmt.Sprintf("received event: %+v %+v %+v\n", data.Name, data.Namespace, data.Message))

	dbEvent, err := converter.ConvertEvent(*data)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to convert event to db model: %s\n", err.Error()))
	}

	queue.EventQueue = append(queue.EventQueue, dbEvent)
	BatchWrite()
	return nil
}

// BatchWrite writes the events in the event queue to the database
func BatchWrite() {
	if WriteConditionsMet() {
		log.Info("writing event batch to database")
		utils.DB.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).CreateInBatches(queue.EventQueue, queue.BatchSize)
		lastWriteTimestamp := time.Now()

		log.Debug(fmt.Sprintf("setting last write timestamp to %s", lastWriteTimestamp.String()))
		queue.LastWriteTimestamp = lastWriteTimestamp

		log.Debug("emptying event queue")
		queue.EventQueue = make(queue.SingletonEventQueue, 0)
	}
}

// WriteConditionsMet checks if the batch size has been reached or the time interval has passed
func WriteConditionsMet() bool {
	if len(queue.EventQueue) >= queue.BatchSize {
		log.Debug("batch size condition met")
		return true
	} else if time.Since(queue.LastWriteTimestamp) >= queue.TimeInterval {
		log.Debug("time interval condition met")
		return true
	}
	log.Debug("write conditions not met")
	return false
}
