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

	"gorm.io/gorm"
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
	case "PrometheusAlerts":
		return writeAlert(event)
	case "FluxInfo":
		return writeFluxInfo(event)
	case "GitCommitInfo":
		return writeGitCommitInfo(event)
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

	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&dbClusterInfo).Error; err != nil {
			return err
		}
		if err := tx.Where("token = ?", data.Token).Delete(&models.NodeInfo{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&dbNodeInfoArray).Error; err != nil {
			return err
		}

		return nil
	})
}

func writeAlert(event ce.Event) error {
	var data payload.PrometheusAlerts

	if err := event.DataAs(&data); err != nil {
		log.Warnf("Failed to parse event as Alert object: %v.", err)
		return err
	}

	log.Infof("Received Alert")

	var dbAlerts []models.Alert

	for _, alert := range data.Alerts {
		dbAlert, err := converter.ConvertAlert(data.Token, alert)
		if err != nil {
			log.Warnf("Failed to convert Alert object to db model: %v.", err)
			return err
		}
		dbAlerts = append(dbAlerts, dbAlert)
	}

	return utils.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("token = ?", data.Token).Delete(&models.Alert{})
		tx.Create(&dbAlerts)
		return nil
	})
}

func writeFluxInfo(event ce.Event) error {
	var data payload.FluxInfo
	if err := event.DataAs(&data); err != nil {
		log.Warnf("failed to parse event as FluxInfo object: %v", err)
		return err
	}

	var cluster models.Cluster
	clusterResult := utils.DB.First(&cluster, "token = ?", data.Token)
	if errors.Is(clusterResult.Error, gorm.ErrRecordNotFound) {
		log.Warnf("received FluxInfo for unknown cluster")
		return fmt.Errorf("received FluxInfo did not match any registered clusters, token: %s", data.Token)
	}
	log.Infof("received FluxInfo for cluster %s", cluster.Name)

	dbFluxInfo, err := converter.ConvertFluxInfo(data)
	if err != nil {
		log.Warnf("failed to convert FluxInfo object to db model: %v", err)
		return err
	}

	utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(dbFluxInfo, queue.BatchSize)

	return nil
}

func writeGitCommitInfo(event ce.Event) error {
	var data payload.GitCommitInfo
	if err := event.DataAs(&data); err != nil {
		log.Warnf("Failed to parse event as GitCommitInfo object: %v", err)
		return err
	}

	var cluster models.Cluster
	clusterResult := utils.DB.First(&cluster, "token = ?", data.Token)
	if errors.Is(clusterResult.Error, gorm.ErrRecordNotFound) {
		log.Warnf("Received GitCommitInfo for unknown cluster")
		return fmt.Errorf("Received GitCommitInfo did not match any registered clusters, token: %s", data.Token)
	}
	log.Debugf("Received GitCommitInfo for cluster %s", cluster.Name)

	dbCommitInfo, err := converter.ConvertGitCommitInfo(data)
	if err != nil {
		log.Warnf("Failed to convert GitCommitInfo object to db model: %v.", err)
		return err
	}

	if err := utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cluster_token = ?", data.Token).Delete(&models.GitCommit{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&dbCommitInfo).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Errorf("Failed to add commit '%s' to the database: %v.", dbCommitInfo.Sha, err)
		return err
	}

	log.Debugf("Commit '%s' added.", dbCommitInfo.Sha)

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
