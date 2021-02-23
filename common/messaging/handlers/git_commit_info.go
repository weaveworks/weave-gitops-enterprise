package handlers

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/payload"
)

type GitCommitInfoSender struct {
	source string
	client cloudevents.Client
}

func NewGitCommitInfoSender(source string, client cloudevents.Client) *GitCommitInfoSender {
	return &GitCommitInfoSender{
		source: source,
		client: client,
	}
}

func (s *GitCommitInfoSender) Send(ctx context.Context, info payload.GitCommitInfo) error {
	event, err := s.ToEvent(info)
	if err != nil {
		log.Errorf("Unable to convert GitCommitInfo to Event: %v.", err)
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

func (s *GitCommitInfoSender) ToEvent(info payload.GitCommitInfo) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("GitCommitInfo")
	e.SetTime(time.Now())
	e.SetSource(s.source)
	if err := e.SetData(cloudevents.ApplicationJSON, info); err != nil {
		log.Errorf("Unable to set event as data: %v.", err)
		return e, err
	}
	return e, nil
}
