package subscribe_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/cmd/event-writer/queue"
	"github.com/weaveworks/wks/cmd/event-writer/subscribe"
	test "github.com/weaveworks/wks/cmd/event-writer/test"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"github.com/weaveworks/wks/common/messaging/payload"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func RunServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = -1 // Allocate a port dynamically
	return natsserver.RunServer(&opts)
}

func newCloudEvent(typ string, obj interface{}) (*ce.Event, error) {
	e := ce.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(typ)
	e.SetTime(time.Now())
	e.SetSource("tests")
	if err := e.SetData("application/json", obj); err != nil {
		log.Errorf("Unable to set object as data: %v.", err)
		return nil, err
	}
	return &e, nil
}

func newk8sEvent(reason, namespace, name string) payload.KubernetesEvent {
	uuid, _ := uuid.NewUUID()
	event := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uuid.String()),
		},
		Reason: reason,
	}
	ret := payload.KubernetesEvent{
		Event: event,
	}
	return ret
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

	ceEvent, err := newCloudEvent("Event", testEvent)
	assert.NoError(t, err)

	err = subscribe.ReceiveEvent(context.Background(), *ceEvent)
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
	subscribe.BatchWrite()

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

		result := subscribe.WriteConditionsMet()
		assert.Equal(t, test.result, result)
	}
}

func TestReceiveClusterInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", subscribe.ReceiveEvent)
		require.NoError(t, err)
	}()

	// Set up publisher
	sender, err := cenats.NewSender(s.ClientURL(), "test.subject", cenats.NatsOptions(
		nats.Name("sender"),
	))
	require.NoError(t, err)
	defer sender.Close(ctx)
	publisher, err := ce.NewClient(sender)
	require.NoError(t, err)

	// Publish event
	info := payload.ClusterInfo{
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				payload.Node{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				payload.Node{
					MachineID:      "953089b9924d3a45febe69bc3add4683",
					Name:           "derp-wks-2",
					IsControlPlane: false,
					KubeletVersion: "v1.19.3",
				},
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("ClusterInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Query db
	var clusters []models.ClusterInfo
	clustersResult := db.Find(&clusters)
	assert.Equal(t, 1, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)

	assert.Equal(t, info.Cluster.ID, string(clusters[0].UID))
	assert.Equal(t, info.Cluster.Type, clusters[0].Type)

	var nodes []models.NodeInfo
	nodesResult := db.Find(&nodes)
	assert.Equal(t, 2, int(nodesResult.RowsAffected))
	assert.NoError(t, nodesResult.Error)

	assert.Equal(t, info.Cluster.Nodes[0].MachineID, string(nodes[0].UID))
	assert.Equal(t, info.Cluster.Nodes[0].Name, nodes[0].Name)
	assert.Equal(t, info.Cluster.Nodes[0].KubeletVersion, nodes[0].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[0].IsControlPlane, nodes[0].IsControlPlane)
	assert.Equal(t, info.Cluster.Nodes[1].MachineID, string(nodes[1].UID))
	assert.Equal(t, info.Cluster.Nodes[1].Name, nodes[1].Name)
	assert.Equal(t, info.Cluster.Nodes[1].KubeletVersion, nodes[1].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[1].IsControlPlane, nodes[1].IsControlPlane)
}

func TestReceiveClusterInfo_PayloadNotClusterInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", subscribe.ReceiveEvent)
		require.NoError(t, err)
	}()

	// Set up publisher
	sender, err := cenats.NewSender(s.ClientURL(), "test.subject", cenats.NatsOptions(
		nats.Name("sender"),
	))
	require.NoError(t, err)
	defer sender.Close(ctx)
	publisher, err := ce.NewClient(sender)
	require.NoError(t, err)

	// Publish event
	info := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "not a cluster info object",
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("ClusterInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Query db
	var clusters []models.ClusterInfo
	clustersResult := db.Find(&clusters)
	assert.Equal(t, 0, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)

	var nodes []models.NodeInfo
	nodesResult := db.Find(&nodes)
	assert.Equal(t, 0, int(nodesResult.RowsAffected))
	assert.NoError(t, nodesResult.Error)
}

func TestReceiveClusterInfo_SamePayloadReceivedAgain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", subscribe.ReceiveEvent)
		require.NoError(t, err)
	}()

	// Set up publisher
	sender, err := cenats.NewSender(s.ClientURL(), "test.subject", cenats.NatsOptions(
		nats.Name("sender"),
	))
	require.NoError(t, err)
	defer sender.Close(ctx)
	publisher, err := ce.NewClient(sender)
	require.NoError(t, err)

	// Publish event
	info := payload.ClusterInfo{
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				payload.Node{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				payload.Node{
					MachineID:      "953089b9924d3a45febe69bc3add4683",
					Name:           "derp-wks-2",
					IsControlPlane: false,
					KubeletVersion: "v1.19.3",
				},
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("ClusterInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	event.SetID(uuid.New().String())
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Query db
	var clusters []models.ClusterInfo
	clustersResult := db.Find(&clusters)
	assert.Equal(t, 1, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)

	assert.Equal(t, info.Cluster.ID, string(clusters[0].UID))
	assert.Equal(t, info.Cluster.Type, clusters[0].Type)

	var nodes []models.NodeInfo
	nodesResult := db.Find(&nodes)
	assert.Equal(t, 2, int(nodesResult.RowsAffected))
	assert.NoError(t, nodesResult.Error)

	assert.Equal(t, info.Cluster.Nodes[0].MachineID, string(nodes[0].UID))
	assert.Equal(t, info.Cluster.Nodes[0].Name, nodes[0].Name)
	assert.Equal(t, info.Cluster.Nodes[0].KubeletVersion, nodes[0].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[0].IsControlPlane, nodes[0].IsControlPlane)
	assert.Equal(t, info.Cluster.Nodes[1].MachineID, string(nodes[1].UID))
	assert.Equal(t, info.Cluster.Nodes[1].Name, nodes[1].Name)
	assert.Equal(t, info.Cluster.Nodes[1].KubeletVersion, nodes[1].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[1].IsControlPlane, nodes[1].IsControlPlane)
}

func TestReceiveClusterInfo_ClusterUpdated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", subscribe.ReceiveEvent)
		require.NoError(t, err)
	}()

	// Set up publisher
	sender, err := cenats.NewSender(s.ClientURL(), "test.subject", cenats.NatsOptions(
		nats.Name("sender"),
	))
	require.NoError(t, err)
	defer sender.Close(ctx)
	publisher, err := ce.NewClient(sender)
	require.NoError(t, err)

	// Publish event
	info := payload.ClusterInfo{
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				payload.Node{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				payload.Node{
					MachineID:      "953089b9924d3a45febe69bc3add4683",
					Name:           "derp-wks-2",
					IsControlPlane: false,
					KubeletVersion: "v1.19.3",
				},
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("ClusterInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)

	// Query db
	var clusters []models.ClusterInfo
	clustersResult := db.Find(&clusters)
	assert.Equal(t, 1, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)

	assert.Equal(t, info.Cluster.ID, string(clusters[0].UID))
	assert.Equal(t, info.Cluster.Type, clusters[0].Type)

	var nodes []models.NodeInfo
	nodesResult := db.Find(&nodes)
	assert.Equal(t, 2, int(nodesResult.RowsAffected))
	assert.NoError(t, nodesResult.Error)

	assert.Equal(t, info.Cluster.Nodes[0].MachineID, string(nodes[0].UID))
	assert.Equal(t, info.Cluster.Nodes[0].Name, nodes[0].Name)
	assert.Equal(t, info.Cluster.Nodes[0].KubeletVersion, nodes[0].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[0].IsControlPlane, nodes[0].IsControlPlane)
	assert.Equal(t, info.Cluster.Nodes[1].MachineID, string(nodes[1].UID))
	assert.Equal(t, info.Cluster.Nodes[1].Name, nodes[1].Name)
	assert.Equal(t, info.Cluster.Nodes[1].KubeletVersion, nodes[1].KubeletVersion)
	assert.Equal(t, info.Cluster.Nodes[1].IsControlPlane, nodes[1].IsControlPlane)

	// Publish 2nd event
	info2 := payload.ClusterInfo{
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				payload.Node{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "foo-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
			},
		},
	}
	event = ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("ClusterInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info2)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Query db
	clustersResult = db.Find(&clusters)
	assert.Equal(t, 1, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)

	assert.Equal(t, info.Cluster.ID, string(clusters[0].UID))
	assert.Equal(t, info.Cluster.Type, clusters[0].Type)

	nodesResult = db.Find(&nodes)
	assert.Equal(t, 1, int(nodesResult.RowsAffected))
	assert.NoError(t, nodesResult.Error)

	assert.Equal(t, info2.Cluster.Nodes[0].MachineID, string(nodes[0].UID))
	assert.Equal(t, info2.Cluster.Nodes[0].Name, nodes[0].Name)
	assert.Equal(t, info2.Cluster.Nodes[0].KubeletVersion, nodes[0].KubeletVersion)
	assert.Equal(t, info2.Cluster.Nodes[0].IsControlPlane, nodes[0].IsControlPlane)
}
