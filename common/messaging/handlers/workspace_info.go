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

type WorkspaceInfoSender struct {
	source string
	client cloudevents.Client
}

func NewWorkspaceInfoSender(source string, client cloudevents.Client) *WorkspaceInfoSender {
	return &WorkspaceInfoSender{
		source: source,
		client: client,
	}
}

func (s *WorkspaceInfoSender) Send(ctx context.Context, info payload.WorkspaceInfo) error {
	event, err := s.toEvent(info)
	if err != nil {
		log.Errorf("Unable to convert WorkspaceInfo to Event: %v.", err)
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

func (s *WorkspaceInfoSender) toEvent(info payload.WorkspaceInfo) (*cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("WorkspaceInfo")
	e.SetTime(time.Now())
	e.SetSource(s.source)
	if err := e.SetData(cloudevents.ApplicationJSON, info); err != nil {
		return nil, fmt.Errorf("failed to set event as data: %w", err)
	}
	return &e, nil
}
