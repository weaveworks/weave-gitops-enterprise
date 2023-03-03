package collector

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	collectorfakes "github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/collectorfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAccessRulesCollector(t *testing.T) {
	store := &storefakes.FakeStoreWriter{}

	cr := &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Rules: []v1.PolicyRule{{
			APIGroups: []string{"somegroup"},
			Verbs:     []string{"get", "list", "watch"},
			Resources: []string{"somekind"},
		}},
	}

	col := &collectorfakes.FakeCollector{}
	ch := make(chan []collector.ObjectRecord)
	col.StartReturns(ch, nil)
	arc := &accessRulesCollector{
		log:       logr.Discard(),
		col:       col,
		w:         store,
		converter: runtime.DefaultUnstructuredConverter,
	}

	t.Run("Start", func(t *testing.T) {
		t.Run("stores access rules", func(t *testing.T) {

			arc.Start()

			fc := &collectorfakes.FakeObjectRecord{}
			fc.ClusterNameReturns("test-cluster")

			fc.ObjectReturns(cr)

			objs := []collector.ObjectRecord{fc}

			ch <- objs

			expected := models.AccessRule{
				Cluster:         "test-cluster",
				Role:            "test",
				AccessibleKinds: cr.Rules[0].Resources,
			}

			assert.Equal(t, store.StoreAccessRulesCallCount(), 1)
			assert.Equal(t, expected, store.StoreAccessRulesArgsForCall(0)[0])
		})
	})
}
