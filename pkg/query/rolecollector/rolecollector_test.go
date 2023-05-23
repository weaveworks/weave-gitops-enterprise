package rolecollector

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRolesCollector_NewRoleCollector(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name       string
		store      store.Store
		opts       collector.CollectorOpts
		errPattern string
	}{
		{
			name:  "cannot create collector without manager",
			store: &storefakes.FakeStore{},
			opts: collector.CollectorOpts{
				Log: logr.Discard(),
			},
			errPattern: "invalid cluster manager",
		},
		{
			name:  "can create object collector with valid arguments",
			store: &storefakes.FakeStore{},
			opts: collector.CollectorOpts{
				Log:            logr.Discard(),
				ClusterManager: &clustersmngrfakes.FakeClustersManager{},
				ServiceAccount: collector.ImpersonateServiceAccount{
					Name:      "anyName",
					Namespace: "anyNamespace",
				},
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := NewRoleCollector(tt.store, tt.opts)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(collector).ShouldNot(BeNil())
			g.Expect(collector.w).To(Equal(tt.store))
		})
	}
}

func TestRoleCollector_defaultProcessRecords(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}
	fakeIndex := &storefakes.FakeIndexWriter{}

	//setup data
	clusterName := "anyCluster"

	tests := []struct {
		name                  string
		objectRecords         []models.ObjectTransaction
		expectedStoreNumCalls map[models.TransactionType]int
		errPattern            string
	}{
		{
			name:          "can process empty records collection with no errors",
			objectRecords: []models.ObjectTransaction{},
			expectedStoreNumCalls: map[models.TransactionType]int{
				models.TransactionTypeDelete:    0,
				models.TransactionTypeUpsert:    0,
				models.TransactionTypeDeleteAll: 0,
			},
			errPattern: "",
		},
		{
			name: "can process non-empty roles collection with no errors",
			objectRecords: []models.ObjectTransaction{
				testutils.NewObjectTransaction("anyCluster", testutils.NewRole("createdOrUpdatedRole", clusterName), models.TransactionTypeUpsert),
				testutils.NewObjectTransaction("anyCluster", testutils.NewRole("deletedRole", clusterName, func(hr *rbacv1.Role) {
					now := metav1.Now()
					hr.DeletionTimestamp = &now
				}), models.TransactionTypeDelete),
				testutils.NewObjectTransaction("anyCluster2", testutils.NewRole("deletedRole2", clusterName, func(hr *rbacv1.Role) {
					now := metav1.Now()
					hr.DeletionTimestamp = &now
				}), models.TransactionTypeDelete),
				testutils.NewObjectTransaction("anyCluster3", nil, models.TransactionTypeDeleteAll),
			},
			expectedStoreNumCalls: map[models.TransactionType]int{
				models.TransactionTypeDelete:    1,
				models.TransactionTypeUpsert:    1,
				models.TransactionTypeDeleteAll: 1,
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := defaultProcessRecords(tt.objectRecords, fakeStore, fakeIndex, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(fakeStore.StoreRolesCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeUpsert]))
			g.Expect(fakeStore.DeleteRolesCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDelete]))
			g.Expect(fakeStore.DeleteAllRolesCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDeleteAll]))
		})
	}

}
