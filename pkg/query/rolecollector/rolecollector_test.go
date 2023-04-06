package rolecollector

import (
	"context"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestRoleCollector(t *testing.T) {
	// TODO: we need to test upserting and deleting access rules

}

func TestRoleCollector_defaultProcessRecords(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}

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
			err := defaultProcessRecords(ctx, tt.objectRecords, fakeStore, log)
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
