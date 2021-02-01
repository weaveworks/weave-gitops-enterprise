package alertmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/prometheus/alertmanager/notify/webhook"
	log "github.com/sirupsen/logrus"
)

const (
	eventType   = "io.prometheus.alertmanager.alert"
	contentType = "application/json"
)

type AlertEvent struct {
	Key      string
	Type     string
	Resource interface{}

	ID     string
	Source url.URL

	DataContentType string
	DataSchema      url.URL
	Subject         string
	Time            *time.Time
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) PushHandler(req *http.Request) (event.Event, error) {
	var msg webhook.Message

	if req.Body == nil {
		return event.Event{}, errors.New("empty payload")
	}

	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	err := decoder.Decode(&msg)
	if err != nil {
		return event.Event{}, err
	}

	ce := cloudevents.NewEvent()
	ce.SetID(uuid.New().String())
	ce.SetType(eventType)
	ce.SetTime(time.Now())
	ce.SetSource(msg.ExternalURL)
	if err := ce.SetData(contentType, msg); err != nil {
		log.Errorf("Unable to set event as data: %v.", err)
		return event.Event{}, err
	}

	return ce, nil
}

func NewWebhookHandler(fn func(event.Event)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ce, err := NewParser().PushHandler(r)
		if err != nil {
			log.Fatalf("Failed to parse alert: %v", err)
		}
		fn(ce)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(""))
	}
}
