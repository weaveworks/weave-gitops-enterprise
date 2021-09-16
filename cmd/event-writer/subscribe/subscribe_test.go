package subscribe_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/converter"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/queue"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/subscribe"
	test "github.com/weaveworks/weave-gitops-enterprise/cmd/event-writer/test"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func RunServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = -1 // Allocate a port dynamically
	return natsserver.RunServer(&opts)
}

func SetupPipeline(t *testing.T) (ce.Client, *gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
		require.NoError(t, err)
	}()

	// Set up publisher
	sender, err := cenats.NewSender(s.ClientURL(), "test.subject", cenats.NatsOptions(
		nats.Name("sender"),
	))
	require.NoError(t, err)
	publisher, err := ce.NewClient(sender)
	require.NoError(t, err)

	// Shut it all down on cancel()
	go func() {
		<-ctx.Done()
		s.Shutdown()
		sender.Close(ctx)
	}()

	return publisher, db, cancel
}

func newCloudEvent(t *testing.T, typ string, obj interface{}) ce.Event {
	e := ce.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(typ)
	e.SetTime(time.Now())
	e.SetSource("tests")
	err := e.SetData("application/json", obj)
	assert.NoError(t, err)
	return e
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

	// Ensure that messages with unknown types are dropped
	ceEvent := newCloudEvent(t, "WrongType", testEvent)
	err := subscribe.ReceiveEvent(context.Background(), ceEvent)
	assert.NoError(t, err)

	// Ensure the event queue length is 0
	assert.Equal(t, len(queue.EventQueue), 0)

	// Send again with the correct type
	ceEvent = newCloudEvent(t, "Event", testEvent)
	err = subscribe.ReceiveEvent(context.Background(), ceEvent)
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
		dbEvent := converter.ConvertEvent(randomEvent)
		q = append(q, dbEvent)
	}
	return q
}

func TestBatchWrite(t *testing.T) {
	testDB := utils.DB
	testDB, err := utils.Open("", "sqlite", "", "", "")
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

func TestReceiveFluxInfo_NoMatchingCluster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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
	info := payload.FluxInfo{
		// list of flux deployments with 1 item
		Token: "derp",
		Deployments: []payload.FluxDeploymentInfo{
			{
				Name:      "flux",
				Namespace: "wkp-flux",
				Args: []string{
					"--memcached-service=",
					"--ssh-keygen-dir=/var/fluxd/keygen",
					"--sync-garbage-collection=true",
					"--git-poll-interval=10s",
					"--sync-interval=10s",
					"--manifest-generation=true",
					"--listen-metrics=:3031",
					"--git-url=git@github.com:dinosk/fluxes-1.git",
					"--git-branch=master",
					"--registry-exclude-image=*"},
				Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
			},
		},
	}

	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("FluxInfo")
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

	// FluxInfo should not be stored as there is no cluster matching the token
	// Query db
	var fluxes []models.FluxInfo
	fluxesResult := db.Find(&fluxes)
	assert.Equal(t, 0, int(fluxesResult.RowsAffected))
	assert.NoError(t, fluxesResult.Error)
}

func TestReceiveFluxInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	queue.BatchSize = 100

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	fluxInfo := payload.FluxInfo{
		// list of flux deployments with 1 item
		Token: "derp",
		Deployments: []payload.FluxDeploymentInfo{
			{
				Name:      "flux",
				Namespace: "wkp-flux",
				Args: []string{
					"--memcached-service=",
					"--ssh-keygen-dir=/var/fluxd/keygen",
					"--sync-garbage-collection=true",
					"--git-poll-interval=10s",
					"--sync-interval=10s",
					"--manifest-generation=true",
					"--listen-metrics=:3031",
					"--git-url=git@github.com:dinosk/fluxes-1.git",
					"--git-branch=master",
					"--registry-exclude-image=*"},
				Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
			},
		},
	}

	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("FluxInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, fluxInfo)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// FluxInfo should be stored as there is no cluster matching the token
	// Query db
	var fluxes []models.FluxInfo
	fluxesResult := db.Find(&fluxes)
	assert.Equal(t, 1, int(fluxesResult.RowsAffected))
	assert.NoError(t, fluxesResult.Error)

	assert.Equal(t, fluxInfo.Token, string(fluxes[0].ClusterToken))
	assert.Equal(t, fluxInfo.Deployments[0].Name, string(fluxes[0].Name))
	assert.Equal(t, fluxInfo.Deployments[0].Namespace, string(fluxes[0].Namespace))
	assert.Equal(t, converter.SerializeStringSlice(fluxInfo.Deployments[0].Args), string(fluxes[0].Args))
}

func TestReceiveClusterInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	info := payload.ClusterInfo{
		Token: "derp",
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				{
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
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	info := payload.ClusterInfo{
		Token: "derp",
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				{
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
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	info := payload.ClusterInfo{
		Token: "derp",
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				{
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
		Token: "derp",
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				{
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

func TestReceiveClusterInfo_NoMatchingCluster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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
		Token: "derp",
		Cluster: payload.Cluster{
			ID:   "8cb9581a-1de1-4a7b-ab2d-16791acc8f74",
			Type: "existinginfra",
			Nodes: []payload.Node{
				{
					MachineID:      "3f28d1dd7291784ed454f52ba0937337",
					Name:           "derp-wks-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19.7",
				},
				{
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
	var clusterInfos []models.ClusterInfo
	clustersResult := db.Find(&clusterInfos)
	assert.Equal(t, 0, int(clustersResult.RowsAffected))
	assert.NoError(t, clustersResult.Error)
}

func TestReceiveAlert(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Alert dates
	startDate := time.Now()
	endDate := startDate.Add(time.Duration(60) * time.Minute)
	updatedDate := startDate.Add(time.Duration(30) * time.Minute)

	annot := ammodels.LabelSet{
		"summary":     "Instance down",
		"description": "Instance has been down for more than 5 minutes.",
	}
	labls := ammodels.LabelSet{
		"severity": "critical",
	}

	var strSlice = []string{"Test1", "Test2", "Test3"}

	receiverName := "My Receiver 1"
	receivr := ammodels.Receiver{
		Name: &receiverName,
	}
	receivrs := []*ammodels.Receiver{&receivr}

	gAlert := newAlert("example.com", "test fingerprint", "active",
		startDate, endDate, updatedDate, annot, labls, strSlice, strSlice, receivrs)

	// Publish event
	info := payload.PrometheusAlerts{
		Token: "derp",
		Alerts: ammodels.GettableAlerts{
			&gAlert,
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("PrometheusAlerts")
	event.SetSource("test")
	event.SetDataContentType("application/json")
	event.SetDataSchema("test")
	event.SetSubject("test Alert")
	event.SetTime(time.Now())
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(1 * time.Second)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	cancel()

	// Query db
	var alerts []models.Alert
	alertsResult := db.Find(&alerts)
	t.Log(alerts)
	assert.Equal(t, 1, int(alertsResult.RowsAffected))
	assert.NoError(t, alertsResult.Error)

	assert.Equal(t, info.Token, string(alerts[0].ClusterToken))
}

func TestReceiveAlert_NoMatchingCluster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	// Alert dates
	startDate := time.Now()
	endDate := startDate.Add(time.Duration(60) * time.Minute)
	updatedDate := startDate.Add(time.Duration(30) * time.Minute)

	annot := ammodels.LabelSet{
		"summary":     "Instance down",
		"description": "Instance has been down for more than 5 minutes.",
	}
	labls := ammodels.LabelSet{
		"severity": "critical",
	}

	var strSlice = []string{"Test1", "Test2", "Test3"}

	receiverName := "My Receiver 1"
	receivr := ammodels.Receiver{
		Name: &receiverName,
	}
	receivrs := []*ammodels.Receiver{&receivr}

	gAlert := newAlert("example.com", "test fingerprint", "active",
		startDate, endDate, updatedDate, annot, labls, strSlice, strSlice, receivrs)

	// Publish event
	info := payload.PrometheusAlerts{
		Token: "derp",
		Alerts: ammodels.GettableAlerts{
			&gAlert,
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("PrometheusAlerts")
	event.SetSource("test")
	event.SetDataContentType("application/json")
	event.SetDataSchema("test")
	event.SetSubject("test Alert")
	event.SetTime(time.Now())
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(1 * time.Second)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	cancel()

	// Query db
	var alerts []models.Alert
	alertsResult := db.Find(&alerts)
	t.Log(alerts)
	assert.Equal(t, 0, int(alertsResult.RowsAffected))
	assert.NoError(t, alertsResult.Error)
}

func TestReceiveAlert_SameAlertReceivedAgain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Alert dates
	startDate := time.Now()
	endDate := startDate.Add(time.Duration(60) * time.Minute)
	updatedDate := startDate.Add(time.Duration(30) * time.Minute)

	annot := ammodels.LabelSet{
		"summary":     "Instance down",
		"description": "Instance has been down for more than 5 minutes.",
	}
	labls := ammodels.LabelSet{
		"severity": "critical",
	}

	var strSlice = []string{"Test1", "Test2", "Test3"}

	receiverName := "My Receiver 1"
	receivr := ammodels.Receiver{
		Name: &receiverName,
	}
	receivrs := []*ammodels.Receiver{&receivr}

	gAlert := newAlert("example.com", "test fingerprint", "active",
		startDate, endDate, updatedDate, annot, labls, strSlice, strSlice, receivrs)

	// Publish event
	info := payload.PrometheusAlerts{
		Token: "derp",
		Alerts: ammodels.GettableAlerts{
			&gAlert,
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("PrometheusAlerts")
	event.SetSource("test")
	event.SetDataContentType("application/json")
	event.SetDataSchema("test")
	event.SetSubject("test Alert")
	event.SetTime(time.Now())
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Query db
	var alerts []models.Alert
	alertsResult := db.Find(&alerts)
	t.Log(alerts)
	assert.Equal(t, 1, int(alertsResult.RowsAffected))
	assert.NoError(t, alertsResult.Error)

	assert.Equal(t, info.Token, string(alerts[0].ClusterToken))
}

func newAlert(generatorURL, finPrint, state string, start, end, update time.Time,
	annot, labels ammodels.LabelSet, inhibitedBy, silencedBy []string,
	receivers []*ammodels.Receiver) ammodels.GettableAlert {
	startDate := strfmt.DateTime(start)
	endDate := strfmt.DateTime(end)
	updatedDate := strfmt.DateTime(update)

	alertStatus := ammodels.AlertStatus{
		InhibitedBy: inhibitedBy,
		SilencedBy:  silencedBy,
		State:       &state,
	}

	alertStruct := ammodels.Alert{
		GeneratorURL: strfmt.URI(generatorURL),
		Labels:       labels,
	}

	alert := ammodels.GettableAlert{
		Annotations: annot,
		EndsAt:      &endDate,
		Fingerprint: &finPrint,
		Receivers:   receivers,
		StartsAt:    &startDate,
		Status:      &alertStatus,
		UpdatedAt:   &updatedDate,
		Alert:       alertStruct,
	}

	return alert
}

func TestReceiveGitCommitInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	// Create cluster
	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	info := payload.GitCommitInfo{
		Token: "derp",
		Commit: payload.CommitView{
			Sha:     "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
			Message: "Fixing prod",
			Author: payload.UserView{
				Name:  "foo",
				Email: "foo@weave.works",
				Date:  time.Now().UTC(),
			},
			Committer: payload.UserView{
				Name:  "bar",
				Email: "bar@weave.works",
				Date:  time.Now().UTC(),
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("GitCommitInfo")
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
	var commits []models.GitCommit
	commitsResult := db.Find(&commits)
	assert.Equal(t, 1, int(commitsResult.RowsAffected))
	assert.NoError(t, commitsResult.Error)

	actual := commits[0]
	assert.Equal(t, info.Token, actual.ClusterToken)
	assert.Equal(t, info.Commit.Sha, actual.Sha)
	assert.Equal(t, info.Commit.Message, actual.Message)
	assert.Equal(t, info.Commit.Author.Name, actual.AuthorName)
	assert.Equal(t, info.Commit.Author.Email, actual.AuthorEmail)
	assert.Equal(t, info.Commit.Author.Date, actual.AuthorDate)
	assert.Equal(t, info.Commit.Committer.Name, actual.CommitterName)
	assert.Equal(t, info.Commit.Committer.Email, actual.CommitterEmail)
	assert.Equal(t, info.Commit.Committer.Date, actual.CommitterDate)
}

func TestReceiveGitCommitInfo_NoMatchingCluster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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
	info := payload.GitCommitInfo{
		Token: "derp",
		Commit: payload.CommitView{
			Sha:     "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
			Message: "Fixing prod",
			Author: payload.UserView{
				Name:  "foo",
				Email: "foo@weave.works",
				Date:  time.Now().UTC(),
			},
			Committer: payload.UserView{
				Name:  "bar",
				Email: "bar@weave.works",
				Date:  time.Now().UTC(),
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("GitCommitInfo")
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
	var commits []models.GitCommit
	commitsResult := db.Find(&commits)
	assert.Equal(t, 0, int(commitsResult.RowsAffected))
	assert.NoError(t, commitsResult.Error)
}

func TestReceiveGitCommitInfo_DeleteOldCommits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	// Create cluster
	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Add old commits
	rightNow := time.Now().UTC()
	count := 60
	oneHourAgo := rightNow.Add(time.Duration(-count) * time.Minute)
	threeDays := rightNow.AddDate(0, 0, -3)
	fourDays := rightNow.AddDate(0, 0, -4)
	fiveDays := rightNow.AddDate(0, 0, -5)

	db.Create(&models.GitCommit{
		ClusterToken:   "derp",
		Sha:            "b218b84857d6d181efe58c1c12b9cef31b78cda9",
		AuthorName:     "foo",
		AuthorEmail:    "foo@weave.works",
		AuthorDate:     fiveDays,
		CommitterName:  "bar",
		CommitterEmail: "bar@weave.works",
		CommitterDate:  fiveDays,
		Message:        "Fixing prod - 5 days",
	})
	db.Create(&models.GitCommit{
		ClusterToken:   "derp",
		Sha:            "e5bc909de4c00a5266878bdea494203b9936328c",
		AuthorName:     "foo",
		AuthorEmail:    "foo@weave.works",
		AuthorDate:     fourDays,
		CommitterName:  "bar",
		CommitterEmail: "bar@weave.works",
		CommitterDate:  fourDays,
		Message:        "Fixing prod - 4 days",
	})
	db.Create(&models.GitCommit{
		ClusterToken:   "derp",
		Sha:            "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
		AuthorName:     "foo",
		AuthorEmail:    "foo@weave.works",
		AuthorDate:     threeDays,
		CommitterName:  "bar",
		CommitterEmail: "bar@weave.works",
		CommitterDate:  threeDays,
		Message:        "Fixing prod - 3 days",
	})
	utils.DB.Create(&models.GitCommit{
		ClusterToken:   "derp",
		Sha:            "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		AuthorName:     "foo",
		AuthorEmail:    "foo@weave.works",
		AuthorDate:     oneHourAgo,
		CommitterName:  "bar",
		CommitterEmail: "bar@weave.works",
		CommitterDate:  oneHourAgo,
		Message:        "Fixing prod - 1 hour",
	})

	// Publish event
	info := payload.GitCommitInfo{
		Token: "derp",
		Commit: payload.CommitView{
			Sha:     "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
			Message: "Fixing prod - now",
			Author: payload.UserView{
				Name:  "foo",
				Email: "foo@weave.works",
				Date:  rightNow,
			},
			Committer: payload.UserView{
				Name:  "bar",
				Email: "bar@weave.works",
				Date:  rightNow,
			},
		},
	}

	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("GitCommitInfo")
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

	// Expected results
	expected := []models.GitCommit{
		{
			ClusterToken:   "derp",
			Sha:            "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
			AuthorName:     "foo",
			AuthorEmail:    "foo@weave.works",
			AuthorDate:     threeDays,
			CommitterName:  "bar",
			CommitterEmail: "bar@weave.works",
			CommitterDate:  threeDays,
			Message:        "Fixing prod - 3 days",
		},
		{
			ClusterToken:   "derp",
			Sha:            "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			AuthorName:     "foo",
			AuthorEmail:    "foo@weave.works",
			AuthorDate:     oneHourAgo,
			CommitterName:  "bar",
			CommitterEmail: "bar@weave.works",
			CommitterDate:  oneHourAgo,
			Message:        "Fixing prod - 1 hour",
		},
		{
			ClusterToken:   "derp",
			Sha:            "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
			AuthorName:     "foo",
			AuthorEmail:    "foo@weave.works",
			AuthorDate:     rightNow,
			CommitterName:  "bar",
			CommitterEmail: "bar@weave.works",
			CommitterDate:  rightNow,
			Message:        "Fixing prod - now",
		},
	}

	// Query db
	var commits []models.GitCommit
	commitsResult := db.Find(&commits)
	assert.Equal(t, 3, int(commitsResult.RowsAffected))
	assert.NoError(t, commitsResult.Error)
	assert.Equal(t, expected, commits)
}

func TestReceiveWorkspaceInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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

	// Create cluster
	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	utils.DB.Create(&cluster)

	// Publish event
	info := payload.WorkspaceInfo{
		Token: "derp",
		Workspaces: []payload.Workspace{
			{
				Name:      "foo-ws",
				Namespace: "foo-ns",
			},
			{
				Name:      "bar-ws",
				Namespace: "bar-ns",
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("WorkspaceInfo")
	event.SetTime(time.Now())
	event.SetSource("test")
	err = event.SetData(ce.ApplicationJSON, info)
	require.NoError(t, err)
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err = publisher.Send(ctx, event)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	defer cancel()

	// Query db
	var workspaces []models.Workspace
	workspacesResult := db.Find(&workspaces)
	assert.Equal(t, 2, int(workspacesResult.RowsAffected))
	assert.NoError(t, workspacesResult.Error)

	ws1 := workspaces[0]
	assert.Equal(t, info.Token, ws1.ClusterToken)
	assert.Equal(t, info.Workspaces[0].Name, ws1.Name)
	assert.Equal(t, info.Workspaces[0].Namespace, ws1.Namespace)
	ws2 := workspaces[1]
	assert.Equal(t, info.Token, ws2.ClusterToken)
	assert.Equal(t, info.Workspaces[1].Name, ws2.Name)
	assert.Equal(t, info.Workspaces[1].Namespace, ws2.Namespace)
}

func TestReceiveWorkspaceInfo_NoMatchingCluster(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory database
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Start a NATS Server
	s := RunServer()
	defer s.Shutdown()

	// Start subscriber
	go func() {
		err := subscribe.ToSubject(ctx, s.ClientURL(), "test.subject", "", subscribe.ReceiveEvent)
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
	info := payload.WorkspaceInfo{
		Token: "derp",
		Workspaces: []payload.Workspace{
			{
				Name:      "foo-ws",
				Namespace: "foo-ns",
			},
			{
				Name:      "bar-ws",
				Namespace: "bar-ns",
			},
		},
	}
	event := ce.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType("WorkspaceInfo")
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
	var workspaces []models.Workspace
	workspacesResult := db.Find(&workspaces)
	assert.Equal(t, 0, int(workspacesResult.RowsAffected))
	assert.NoError(t, workspacesResult.Error)
}

func TestReceiveCAPIClusterInfo(t *testing.T) {
	publisher, db, cancel := SetupPipeline(t)
	defer cancel()

	// Create cluster
	cluster := models.Cluster{
		Token:      "derp",
		Name:       "test-cluster",
		IngressURL: "",
	}
	db.Create(&cluster)

	// Publish event
	// Give enough time for subscriber to subscribe to subject and process the event
	time.Sleep(500 * time.Millisecond)
	err := publisher.Send(context.Background(), newCloudEvent(t, "CAPIClusterInfo", payload.CAPIClusterInfo{
		Token: "derp",
		CAPIClusters: []payload.CAPICluster{
			{
				Name:          "foo-name",
				Namespace:     "foo-ns",
				EncodedObject: `"foo-data"`,
			},
			{
				Name:      "bar-name",
				Namespace: "bar-ns",
			},
		},
	}))
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)

	// Query db
	var capiClusters []models.CAPICluster
	rows := db.Find(&capiClusters)
	assert.Equal(t, 2, int(rows.RowsAffected))
	assert.NoError(t, rows.Error)

	assert.Equal(t, string(capiClusters[0].Name), "foo-name")
	assert.Equal(t, string(capiClusters[0].Namespace), "foo-ns")
	assert.Equal(t, string(capiClusters[0].Object), `"foo-data"`)
}
