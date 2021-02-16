package subscribe

import (
	"context"
	"errors"
	"fmt"
	"time"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/cmd/event-writer/queue"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"github.com/weaveworks/wks/common/messaging/payload"
	"gorm.io/gorm/clause"
)

// ToSubject subscribes to a subject given a nats connection
func ToSubject(ctx context.Context, server string, subject string, fn interface{}) error {
	// Channel Subscriber
	log.Debug("Creating new cloudevents NATS consumer.")
	p, err := cenats.NewConsumer(server, subject, cenats.NatsOptions())
	if err != nil {
		log.Fatalf("Failed to create NATS consumer: %v.", err)
	}

	defer p.Close(ctx)

	log.Debugf("Creating client for NATS server: %v.", p.Conn.Servers())
	c, err := ce.NewClient(p)
	if err != nil {
		log.Fatalf("Failed to create NATS client: %v.", err)
	}

	for {
		log.Debug("Starting NATS receiver.")
		if err := c.StartReceiver(ctx, fn); err != nil {
			log.Warnf("Failed to start NATS receiver: %v.", err)
		}
	}
}

// ReceiveEvent can be passed as a callback function in ToSubject to store the received events to the DB
func ReceiveEvent(ctx context.Context, event ce.Event) error {
	switch event.Type() {
	case "Event":
		enqueueEvent(event)
		BatchWrite()
	case "ClusterInfo":
		return writeClusterInfo(event)
	default:
		log.Warnf("Unknown message type: %s.", event.Type())
	}

	return nil
}

func enqueueEvent(event ce.Event) {
	data := &payload.KubernetesEvent{}
	if err := event.DataAs(data); err != nil {
		log.Warn(fmt.Sprintf("failed to parse event: %s\n", err.Error()))
		return
	}
	log.Info(fmt.Sprintf("received event: %+v %+v %+v\n", data.Event.Name, data.Event.Namespace, data.Event.Message))

	dbEvent, err := converter.ConvertEvent(*data)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to convert event to db model: %s\n", err.Error()))
		return
	}

	queue.EventQueue = append(queue.EventQueue, dbEvent)
}

func writeClusterInfo(event ce.Event) error {
	var data payload.ClusterInfo
	if err := event.DataAs(&data); err != nil {
		log.Warnf("Failed to parse event as ClusterInfo object: %v.", err)
		return err
	}

	// If the ID is empty, the received message was probably of the wrong type
	if data.Cluster.ID == "" {
		log.Warnf("Failed to parse event %s correctly.", event.ID())
		return errors.New("failed to parse event as ClusterInfo object")
	}

	log.Infof("Received ClusterInfo: %s %s.", data.Cluster.ID, data.Cluster.Type)

	dbClusterInfo, err := converter.ConvertClusterInfo(data)
	if err != nil {
		log.Warnf("Failed to convert ClusterInfo object to db model: %v.", err)
		return err
	}

	dbNodeInfoArray, err := converter.ConvertNodeInfo(data, dbClusterInfo.UID)
	if err != nil {
		log.Warnf("Failed to convert NodeInfo array to db model: %v.", err)
		return err
	}

	utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&dbClusterInfo)

	utils.DB.Where("cluster_info_uid = ?", dbClusterInfo.UID).Delete(models.NodeInfo{})

	utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&dbNodeInfoArray)

	return nil
}

// BatchWrite writes the events in the event queue to the database
func BatchWrite() {
	if WriteConditionsMet() {
		WriteAllQueues()
		EmptyAllQueues()
		UpdateLastWriteTimestamp()
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

// WriteAllQueues batch writes all items currently queued
func WriteAllQueues() {
	utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(queue.EventQueue, queue.BatchSize)
}

// EmptyAllQueues clears all queues
func EmptyAllQueues() {
	log.Debug("emptying event queue")
	queue.EventQueue = make(queue.SingletonEventQueue, 0)
}

// UpdateLastWriteTimestamp sets the last write timestamp to time.Now()
func UpdateLastWriteTimestamp() {
	lastWriteTimestamp := time.Now()

	log.Debug(fmt.Sprintf("setting last write timestamp to %s", lastWriteTimestamp.String()))
	queue.LastWriteTimestamp = lastWriteTimestamp
}
