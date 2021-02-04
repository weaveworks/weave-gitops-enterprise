package poller_test

import (
	"context"
	"sync"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	clusterpoller "github.com/weaveworks/wks/pkg/cluster/poller"
	"github.com/weaveworks/wks/pkg/messaging/handlers"
	"github.com/weaveworks/wks/pkg/messaging/payload"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterInfoPoller(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object // state of cluster before executing tests
		expected     *payload.ClusterInfo
	}{
		{
			name: "No nodes",
			clusterState: []runtime.Object{
				&v1.Namespace{},
			},
			expected: nil,
		},
		{
			name: "No namespaces",
			clusterState: []runtime.Object{
				&v1.Node{},
			},
			expected: nil,
		},
		{
			name: "1CP 1W cluster",
			clusterState: []runtime.Object{
				&v1.Node{
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
				},
				&v1.Node{
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
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
						UID:  "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
					},
				},
			},
			expected: &payload.ClusterInfo{
				ID:   "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
				Type: "existingInfra",
				Nodes: []payload.NodeInfo{
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
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleClientset(tt.clusterState...)
			client := NewFakeCloudEventsClient()
			sender := handlers.NewClusterInfoSender("test", client)
			poller := clusterpoller.NewClusterInfoPoller(clientset, time.Second, sender)

			// Run poller enough time to send an event then cancel it
			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoEventsWereSent(t)
			} else {
				client.AssertClusterInfoSent(t, *tt.expected)
			}
		})
	}
}

type FakeCloudEventsClient struct {
	sync.Mutex
	sent map[string]cloudevents.Event
}

func NewFakeCloudEventsClient() *FakeCloudEventsClient {
	sent := make(map[string]cloudevents.Event)

	return &FakeCloudEventsClient{
		sent: sent,
	}
}

func (c *FakeCloudEventsClient) Request(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
	return nil, nil
}

func (c *FakeCloudEventsClient) Send(ctx context.Context, event cloudevents.Event) cloudevents.Result {

	c.Lock()
	defer c.Unlock()
	c.sent[event.ID()] = event

	return nil
}

func (c *FakeCloudEventsClient) StartReceiver(ctx context.Context, fn interface{}) error {
	return nil
}

func (c *FakeCloudEventsClient) AssertClusterInfoSent(t *testing.T, expected payload.ClusterInfo) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.ClusterInfo, 0)
	var info payload.ClusterInfo
	for _, e := range c.sent {
		_ = e.DataAs(&info)
		list = append(list, info)
	}

	assert.Subset(t, list, []payload.ClusterInfo{expected})
}

func (c *FakeCloudEventsClient) AssertNoEventsWereSent(t *testing.T) {
	assert.Empty(t, c.sent)
}
