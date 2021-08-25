package test

import (
	contxt "context"
	"fmt"
	"testing"
	"time"

	cloudeventsnats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	clusterpoller "github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/poller"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func RunServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = -1 // Allocate a port dynamically
	return natsserver.RunServer(&opts)
}

func TestAgent(t *testing.T) {
	t.Run("Send event to NATS", func(t *testing.T) {
		ctx, cancel := contxt.WithCancel(contxt.Background())
		defer cancel()

		s := RunServer()
		defer s.Shutdown()

		// Set up publisher
		sender, err := cloudeventsnats.NewSender(s.ClientURL(), "test.subject", cloudeventsnats.NatsOptions(
			nats.Name("sender"),
		))
		require.NoError(t, err)
		defer sender.Close(ctx)
		publisher, err := cloudevents.NewClient(sender)
		require.NoError(t, err)
		notifier := handlers.NewEventNotifier("derp", "test", publisher)
		require.NoError(t, err)

		// Set up subscriber
		consumer, err := cloudeventsnats.NewConsumer(s.ClientURL(), "test.subject", cloudeventsnats.NatsOptions(
			nats.Name("consumer"),
		))
		require.NoError(t, err)
		subscriber, err := cloudevents.NewClient(consumer)
		require.NoError(t, err)

		events := make(chan cloudevents.Event)

		expected := &v1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-event",
				Namespace: "wkp-ns",
			},
		}

		go func() {
			if err := subscriber.StartReceiver(ctx, func(event cloudevents.Event) {
				events <- event
				// Shut down subscriber after receiving one event
				cancel()
			}); err != nil {
				t.Logf("Failed to start NATS subscriber: %v.", err)
			}
		}()

		// Wait enough time for subscriber to subscribe
		time.Sleep(50 * time.Millisecond)
		err = notifier.Notify("added", expected)
		require.NoError(t, err)

		var actual payload.KubernetesEvent
		select {
		case e := <-events:
			err := e.DataAs(&actual)
			require.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Logf("Time out waiting for event to arrive")
		}

		assert.Equal(t, expected, &actual.Event)
	})

	t.Run("Poll for ClusterInfo", func(t *testing.T) {
		ctx, cancel := contxt.WithCancel(contxt.Background())
		defer cancel()

		s := RunServer()
		defer s.Shutdown()

		// Set up publisher
		sender, err := cloudeventsnats.NewSender(s.ClientURL(), "test.subject", cloudeventsnats.NatsOptions(
			nats.Name("sender"),
		))
		require.NoError(t, err)
		defer sender.Close(ctx)
		publisher, err := cloudevents.NewClient(sender)
		require.NoError(t, err)

		expected := payload.ClusterInfo{
			Token: "derp",
			Cluster: payload.Cluster{
				ID:   "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
				Type: "existingInfra",
				Nodes: []payload.Node{
					{
						MachineID:      "e3801e6f-13b6-4e39-a234-435b4f6b0011",
						Name:           "derp-wks-1",
						IsControlPlane: true,
						KubeletVersion: "v1.19.4",
					},
					{
						MachineID:      "9c6708f5-9aa0-4a09-8d41-362b49f62a76",
						Name:           "derp-wks-2",
						IsControlPlane: false,
						KubeletVersion: "v1.19.3",
					},
				},
			},
		}

		controlplane := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "derp-wks-1",
				Labels: map[string]string{
					"node-role.kubernetes.io/master": "",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "existingInfra://derp-wks-1",
			},
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					MachineID:      "e3801e6f-13b6-4e39-a234-435b4f6b0011",
					KubeletVersion: "v1.19.4",
				},
			},
		}
		worker := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "derp-wks-2",
			},
			Spec: v1.NodeSpec{
				ProviderID: "existingInfra://derp-wks-2",
			},
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					MachineID:      "9c6708f5-9aa0-4a09-8d41-362b49f62a76",
					KubeletVersion: "v1.19.3",
				},
			},
		}
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
				UID:  "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
			},
		}
		clientset := fake.NewSimpleClientset(controlplane, worker, namespace)
		cis := handlers.NewClusterInfoSender("test", publisher)
		poller := clusterpoller.NewClusterInfoPoller("derp", clientset, time.Minute, cis)

		// Set up subscriber
		consumer, err := cloudeventsnats.NewConsumer(s.ClientURL(), "test.subject", cloudeventsnats.NatsOptions(
			nats.Name("consumer"),
		))
		require.NoError(t, err)
		subscriber, err := cloudevents.NewClient(consumer)
		require.NoError(t, err)

		events := make(chan cloudevents.Event)
		go func() {
			if err := subscriber.StartReceiver(ctx, func(event cloudevents.Event) {
				fmt.Printf("Received event: %v", event)
				events <- event
				// Shut down subscriber after receiving one event
				cancel()
			}); err != nil {
				t.Logf("Failed to start NATS subscriber: %v.", err)
			}
		}()

		go poller.Run(ctx.Done())

		// Stop poller after 50 sec
		time.Sleep(50 * time.Millisecond)
		cancel()

		var actual payload.ClusterInfo
		select {
		case e := <-events:
			err := e.DataAs(&actual)
			require.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Logf("Time out waiting for event to arrive")
		}

		assert.Equal(t, expected, actual)
	})

}
