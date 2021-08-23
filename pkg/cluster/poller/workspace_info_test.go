package poller_test

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	clusterpoller "github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/poller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

var gvrToListKind = map[schema.GroupVersionResource]string{
	clusterpoller.WorkspaceGVR: "WorkspacesList",
}

func TestWorkspaceInfoPoller(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		token        string
		expected     *payload.WorkspaceInfo
	}{
		{
			name:         "No workspace objects found",
			clusterState: []runtime.Object{},
			token:        "derp",
			expected: &payload.WorkspaceInfo{
				Token: "derp",
			},
		},
		{
			name: "Workspace objects found",
			clusterState: []runtime.Object{
				createWorkspaceObj("foo-ws", "foo-ns"),
				createWorkspaceObj("bar-ws", "bar-ns"),
			},
			token: "derp",
			expected: &payload.WorkspaceInfo{
				Token: "derp",
				Workspaces: []payload.Workspace{
					{
						Name:      "bar-ws",
						Namespace: "bar-ns",
					},
					{
						Name:      "foo-ws",
						Namespace: "foo-ns",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind, tt.clusterState...)
			client := handlerstest.NewFakeCloudEventsClient()
			sender := handlers.NewWorkspaceInfoSender("test", client)
			poller := clusterpoller.NewWorkspaceInfoPoller(tt.token, clientset, time.Second, sender)

			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertWorkspaceInfoWasSent(t, *tt.expected)
			}
		})
	}
}

func createWorkspaceObj(name, namespace string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{Object: map[string]interface{}{}}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "wkp.weave.works",
		Version: "v1beta1",
		Kind:    "workspace",
	})
	u.SetName(name)
	u.SetNamespace(namespace)
	return u
}
