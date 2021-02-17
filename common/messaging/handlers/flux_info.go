package handlers

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/payload"
)

type FluxInfoSender struct {
	source string
	client cloudevents.Client
}

func NewFluxInfoSender(source string, client cloudevents.Client) *FluxInfoSender {
	return &FluxInfoSender{
		source: source,
		client: client,
	}
}

func (s *FluxInfoSender) Send(ctx context.Context, info payload.FluxInfo) error {
	event, err := s.ToEvent(info)
	if err != nil {
		log.Errorf("Unable to convert FluxInfo to Event: %v.", err)
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

func (s *FluxInfoSender) ToEvent(info payload.FluxInfo) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("FluxInfo")
	e.SetTime(time.Now())
	e.SetSource(s.source)
	if err := e.SetData(cloudevents.ApplicationJSON, info); err != nil {
		log.Errorf("Unable to set event as data: %v.", err)
		return e, err
	}
	return e, nil
}
