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

var capiGvrToListKind = map[schema.GroupVersionResource]string{
	clusterpoller.ClusterGVRs[0]: "ClusterList",
	clusterpoller.ClusterGVRs[1]: "ClusterList",
}

func TestCAPIClusterInfoPoller(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		token        string
		expected     *payload.CAPIClusterInfo
	}{
		{
			name:         "No workspace objects found",
			clusterState: []runtime.Object{},
			token:        "derp",
			expected: &payload.CAPIClusterInfo{
				Token: "derp",
			},
		},
		{
			name: "Workspace objects found",
			clusterState: []runtime.Object{
				createClusterObject("bar-name", "bar-ns", "v1alpha3"),
				createClusterObject("foo-name", "foo-ns", "v1alpha4"),
			},
			token: "derp",
			expected: &payload.CAPIClusterInfo{
				Token: "derp",
				CAPIClusters: []payload.CAPICluster{
					{
						Name:          "bar-name",
						Namespace:     "bar-ns",
						CAPIVersion:   "v1alpha3",
						EncodedObject: `{"apiVersion":"cluster.x-k8s.io/v1alpha3","kind":"cluster","metadata":{"name":"bar-name","namespace":"bar-ns"}}` + "\n",
					},
					{
						Name:          "foo-name",
						Namespace:     "foo-ns",
						CAPIVersion:   "v1alpha4",
						EncodedObject: `{"apiVersion":"cluster.x-k8s.io/v1alpha4","kind":"cluster","metadata":{"name":"foo-name","namespace":"foo-ns"}}` + "\n",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), capiGvrToListKind, tt.clusterState...)
			client := handlerstest.NewFakeCloudEventsClient()
			sender := handlers.NewCAPIClusterInfoSender("test", client)
			poller := clusterpoller.NewCAPIClusterInfoPoller(tt.token, clientset, time.Second, sender)

			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertCAPIClusterInfoWasSent(t, *tt.expected)
			}
		})
	}
}

func createClusterObject(name, namespace, version string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{Object: map[string]interface{}{}}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: version,
		Kind:    "cluster",
	})
	u.SetName(name)
	u.SetNamespace(namespace)
	return u
}
