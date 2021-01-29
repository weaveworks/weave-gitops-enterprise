package subscribe

import (
	"context"
	"os"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tj/assert"
	"github.com/weaveworks/wks/cmd/event-writer/database/models"
	"github.com/weaveworks/wks/cmd/event-writer/database/utils"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newCloudEvent(event v1.Event) (*ce.Event, error) {
	e := ce.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("Event")
	e.SetTime(time.Now())
	e.SetSource("tests")
	if err := e.SetData("application/json", event); err != nil {
		log.Errorf("Unable to set event as data: %v.", err)
		return nil, err
	}
	return &e, nil
}

func newk8sEvent(reason, namespace, name string) v1.Event {
	event := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Reason: reason,
	}
	return event
}

func TestReceiveEvent(t *testing.T) {
	testDB, err := utils.Open("test.db")
	defer os.Remove("test.db")
	assert.NoError(t, err)

	err = utils.MigrateTables(testDB)
	assert.NoError(t, err)

	reason := "FailedToCreateContainer"
	namespace := "kube-system"
	name := "weave-net-5zqlf"
	testEvent := newk8sEvent(reason, namespace, name)

	ceEvent, err := newCloudEvent(testEvent)
	assert.NoError(t, err)

	err = ReceiveEvent(context.Background(), *ceEvent)
	assert.NoError(t, err)

	assertEventInDB(t, testDB, name)
}

func assertEventInDB(t *testing.T, db *gorm.DB, name string) {
	// Ensure the row count of the event table is 1
	var events []models.Event
	var count int64
	db.Model(&events).Count(&count)
	assert.Equal(t, int(count), 1)

	// Get the first event and assert it has the correct name
	var result models.Event
	db.First(&result)
	assert.Equal(t, result.Name, name)
}
