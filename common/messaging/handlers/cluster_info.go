package handlers

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/payload"
)

type ClusterInfoSender struct {
	source string
	client cloudevents.Client
}

func NewClusterInfoSender(source string, client cloudevents.Client) *ClusterInfoSender {
	return &ClusterInfoSender{
		source: source,
		client: client,
	}
}

func (s *ClusterInfoSender) Send(ctx context.Context, info payload.ClusterInfo) error {
	event, err := s.ToEvent(info)
	if err != nil {
		log.Errorf("Unable to convert ClusterInfo to Event: %v.", err)
		return err
	}

	if result := s.client.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Errorf("Unable to send data: %v.", result)
		return result
	} else {
		log.Debugf("Successfully sent %s(%s), recipient acknowledged: %t.", event.Type(), event.ID(), cloudevents.IsACK(result))
		log.Tracef("Event payload: \n%v", event)
		return nil
	}
}

func (s *ClusterInfoSender) ToEvent(info payload.ClusterInfo) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("ClusterInfo")
	e.SetTime(time.Now())
	e.SetSource(s.source)
	if err := e.SetData(cloudevents.ApplicationJSON, info); err != nil {
		log.Errorf("Unable to set event as data: %v.", err)
		return e, err
	}
	return e, nil
}
