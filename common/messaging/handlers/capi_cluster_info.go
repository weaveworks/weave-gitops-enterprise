package handlers

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/payload"
)

type CAPIClusterInfoSender struct {
	source string
	client cloudevents.Client
}

func NewCAPIClusterInfoSender(source string, client cloudevents.Client) *CAPIClusterInfoSender {
	return &CAPIClusterInfoSender{
		source: source,
		client: client,
	}
}

func (s *CAPIClusterInfoSender) Send(ctx context.Context, info payload.CAPIClusterInfo) error {
	event, err := s.toEvent(info)
	if err != nil {
		log.Errorf("Unable to convert CAPIClusterInfo to Event: %v.", err)
		return err
	}

	if result := s.client.Send(ctx, *event); cloudevents.IsUndelivered(result) {
		log.Errorf("Unable to send data: %v.", result)
		return result
	} else {
		log.Debugf("Successfully sent %s(%s), recipient acknowledged: %t.", event.Type(), event.ID(), cloudevents.IsACK(result))
		log.Tracef("Event payload: \n%v", event)
		return nil
	}
}

func (s *CAPIClusterInfoSender) toEvent(info payload.CAPIClusterInfo) (*cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("CAPIClusterInfo")
	e.SetTime(time.Now())
	e.SetSource(s.source)
	if err := e.SetData(cloudevents.ApplicationJSON, info); err != nil {
		return nil, fmt.Errorf("failed to set event as data: %w", err)
	}
	return &e, nil
}
