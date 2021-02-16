package handlerstest

import (
	"context"
	"sync"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/common/messaging/payload"
)

type FakeCloudEventsClient struct {
	sync.Mutex
	sendErr error
	sent    map[string]cloudevents.Event
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

	if c.sendErr != nil {
		return c.sendErr
	}

	c.Lock()
	defer c.Unlock()
	c.sent[event.ID()] = event

	return nil
}

func (c *FakeCloudEventsClient) StartReceiver(ctx context.Context, fn interface{}) error {
	return nil
}

func (c *FakeCloudEventsClient) SetupErrorForSend(err error) {
	c.sendErr = err
}

func (c *FakeCloudEventsClient) AssertClusterInfoWasSent(t *testing.T, expected payload.ClusterInfo) {
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

func (c *FakeCloudEventsClient) AssertEventWasSent(t *testing.T, expected payload.KubernetesEvent) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.KubernetesEvent, 0)
	var ev payload.KubernetesEvent
	for _, e := range c.sent {
		_ = e.DataAs(&ev)
		list = append(list, ev)
	}

	assert.Subset(t, list, []payload.KubernetesEvent{expected})
}

func (c *FakeCloudEventsClient) AssertNoCloudEventsWereSent(t *testing.T) {
	assert.Empty(t, c.sent)
}
