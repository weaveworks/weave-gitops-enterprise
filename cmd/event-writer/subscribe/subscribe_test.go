package subscribe

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tj/assert"
	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/cmd/event-writer/database/models"
	"github.com/weaveworks/wks/cmd/event-writer/database/utils"
	"github.com/weaveworks/wks/cmd/event-writer/queue"
	test "github.com/weaveworks/wks/cmd/event-writer/test"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	uuid, _ := uuid.NewUUID()
	event := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uuid.String()),
		},
		Reason: reason,
	}
	return event
}

func TestReceiveEvent(t *testing.T) {
	reason := "FailedToCreateContainer"
	namespace := "kube-system"
	name := "weave-net-5zqlf"
	testEvent := newk8sEvent(reason, namespace, name)

	queue.NewEventQueue()
	queue.BatchSize = 100
	queue.LastWriteTimestamp = time.Now()
	queue.TimeInterval = time.Duration(50) * time.Second

	ceEvent, err := newCloudEvent(testEvent)
	assert.NoError(t, err)

	err = ReceiveEvent(context.Background(), *ceEvent)
	assert.NoError(t, err)

	// Ensure the event queue length is 1
	assert.Equal(t, len(queue.EventQueue), 1)

	// Get the first event and assert it has the correct name
	firstEvent := queue.EventQueue[0]
	assert.Equal(t, firstEvent.Name, name)
}

type writeConditions struct {
	testQueue          []models.Event
	batchSize          int
	timeInterval       int
	lastWriteTimestamp time.Time
}

func generateRandomEventQueue(length int) []models.Event {
	q := []models.Event{}
	for i := 0; i < length; i++ {
		randomEvent := newk8sEvent(test.RandomString(8), test.RandomString(8), test.RandomString(8))
		dbEvent, _ := converter.ConvertEvent(randomEvent)
		q = append(q, dbEvent)
	}
	return q
}

func TestBatchWrite(t *testing.T) {
	testDB := utils.DB
	testDB, err := utils.Open("test.db")
	defer os.Remove("test.db")
	assert.NoError(t, err)

	err = utils.MigrateTables(testDB)
	assert.NoError(t, err)

	queue.BatchSize = 100
	queue.LastWriteTimestamp = time.Now().Add(-3 * time.Second)
	queue.TimeInterval = time.Duration(2) * time.Second
	testEventQueue := generateRandomEventQueue(150)
	queue.EventQueue = testEventQueue

	fmt.Println(len(queue.EventQueue))
	BatchWrite()

	// Ensure the row count of the event table is 2500000
	var events []models.Event
	var count int64
	testDB.Model(&events).Count(&count)
	assert.Equal(t, int(count), 150)
}

func TestWriteConditionsMet(t *testing.T) {
	tests := []struct {
		params writeConditions
		result bool
	}{
		{
			writeConditions{
				// empty queue
				testQueue:          []models.Event{},
				batchSize:          10,
				timeInterval:       2,
				lastWriteTimestamp: time.Now().Add(-3 * time.Second),
			},
			true,
		},
		{
			writeConditions{
				// empty queue
				testQueue:          []models.Event{},
				batchSize:          100,
				timeInterval:       20,
				lastWriteTimestamp: time.Now().Add(-3 * time.Second),
			},
			false,
		},
		{
			writeConditions{
				// queue length > batchSize, time interval not passed
				testQueue:          generateRandomEventQueue(150),
				batchSize:          100,
				timeInterval:       20,
				lastWriteTimestamp: time.Now().Add(-3 * time.Second),
			},
			true,
		},
		{
			writeConditions{
				// queue length < batchSize, time interval not passed
				testQueue:          generateRandomEventQueue(50),
				batchSize:          100,
				timeInterval:       20,
				lastWriteTimestamp: time.Now().Add(-3 * time.Second),
			},
			false,
		},
		{
			writeConditions{
				// queue length > batchSize, time interval passed
				testQueue:          generateRandomEventQueue(500),
				batchSize:          100,
				timeInterval:       2,
				lastWriteTimestamp: time.Now().Add(-3 * time.Second),
			},
			true,
		},
	}

	for _, test := range tests {
		queue.BatchSize = test.params.batchSize
		queue.EventQueue = test.params.testQueue
		queue.TimeInterval = time.Duration(test.params.timeInterval) * time.Second
		log.Info("time interval is:", queue.TimeInterval)
		queue.LastWriteTimestamp = test.params.lastWriteTimestamp

		result := WriteConditionsMet()
		assert.Equal(t, test.result, result)
	}
}
