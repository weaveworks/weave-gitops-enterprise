package handlerstest

import (
	"context"
	"sync"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
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

func (c *FakeCloudEventsClient) AssertFluxInfoWasSent(t *testing.T, expected payload.FluxInfo) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.FluxInfo, 0)
	var info payload.FluxInfo
	for _, e := range c.sent {
		_ = e.DataAs(&info)
		list = append(list, info)
	}

	assert.Subset(t, list, []payload.FluxInfo{expected})
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

func (c *FakeCloudEventsClient) AssertGitCommitInfoWasSent(t *testing.T, expected payload.GitCommitInfo) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.GitCommitInfo, 0)
	var ev payload.GitCommitInfo
	for _, e := range c.sent {
		_ = e.DataAs(&ev)
		list = append(list, ev)
	}

	var found bool
	for _, gc := range list {
		if cmp.Equal(expected, gc, cmpopts.IgnoreFields(payload.UserView{}, "Date")) {
			found = true
		}
	}

	if !found {
		t.Errorf("Expected to have sent GitCommitInfo %v but has sent instead %v", expected, list)
	}
}

func (c *FakeCloudEventsClient) AssertWorkspaceInfoWasSent(t *testing.T, expected payload.WorkspaceInfo) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.WorkspaceInfo, 0)
	var ev payload.WorkspaceInfo
	for _, e := range c.sent {
		_ = e.DataAs(&ev)
		list = append(list, ev)
	}

	assert.Subset(t, list, []payload.WorkspaceInfo{expected})
}

func (c *FakeCloudEventsClient) AssertCAPIClusterInfoWasSent(t *testing.T, expected payload.CAPIClusterInfo) {
	c.Lock()
	defer c.Unlock()

	list := make([]payload.CAPIClusterInfo, 0)
	var ev payload.CAPIClusterInfo
	for _, e := range c.sent {
		_ = e.DataAs(&ev)
		list = append(list, ev)
	}

	assert.Subset(t, list, []payload.CAPIClusterInfo{expected})
}
