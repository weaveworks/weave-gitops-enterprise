package collector

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPollingCollector(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		dep := &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1beta1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "test-namespace",
			},
		}
		k8s := createClient(t, dep)
		f := &clusterfakes.FakeCluster{}
		f.GetServerClientReturns(k8s, nil)

		mgr := &clustersmngrfakes.FakeClustersManager{}
		mgr.GetClustersReturns([]cluster.Cluster{f})
		log := logr.Discard()
		kinds := []schema.GroupVersionKind{appsv1.SchemeGroupVersion.WithKind("Deployment")}

		collector := NewCollector(CollectorOpts{
			Log:            log,
			ClusterManager: mgr,
			ObjectKinds:    kinds,
			PollInterval:   100 * time.Millisecond,
		})

		t.Run("returns objects on a channel", func(t *testing.T) {
			msg, err := collector.Start()
			assert.NoError(t, err)
			for {
				select {
				case obj := <-msg:

					assert.Len(t, obj, 1)
					assert.Equal(t, obj[0].Object().GetName(), dep.Name)
					return
				case <-time.After(1 * time.Second):
					t.Fatal("timed out waiting for objects")
				}
			}
		})

	})

	t.Run("Stop", func(t *testing.T) {
		t.Run("closes the channel", func(t *testing.T) {
			t.Skip("not implemented")
		})

		t.Run("stops the goroutine", func(t *testing.T) {
			t.Skip("not implemented")
		})
	})
}
