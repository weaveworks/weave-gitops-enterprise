package objectscollector

import (
	"context"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestObjectsCollector(t *testing.T) {
	// TODD: We need to test the objects collector with a "running" cluster using the envtest library
}

func TestObjectsCollector_defaultProcessRecords(t *testing.T) {
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
			name: "can process non-empty record collection with no errors",
			objectRecords: []models.ObjectTransaction{
				testTransaction{
					clusterName:     "anyCluster",
					object:          testutils.NewHelmRelease("createdOrUpdatedHelmRelease", clusterName),
					transactionType: models.TransactionTypeUpsert,
				},
				testTransaction{
					clusterName: "anyCluster",
					object: testutils.NewHelmRelease("deletedHelmRelease1", clusterName, func(hr *v2beta1.HelmRelease) {
						now := metav1.Now()
						hr.DeletionTimestamp = &now
					}),
					transactionType: models.TransactionTypeDelete,
				},
				testTransaction{
					clusterName: "anyCluster2",
					object: testutils.NewHelmRelease("deletedHelmRelease2", clusterName, func(hr *v2beta1.HelmRelease) {
						now := metav1.Now()
						hr.DeletionTimestamp = &now
					}),
					transactionType: models.TransactionTypeDelete,
				},
			},
			expectedStoreNumCalls: map[models.TransactionType]int{
				models.TransactionTypeDelete:    1,
				models.TransactionTypeUpsert:    1,
				models.TransactionTypeDeleteAll: 0,
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
			g.Expect(fakeStore.StoreObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeUpsert]))
			g.Expect(fakeStore.DeleteObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDelete]))
		})
	}

}

type testTransaction struct {
	clusterName     string
	object          client.Object
	transactionType models.TransactionType
}

func (r testTransaction) ClusterName() string {
	return r.clusterName
}

func (r testTransaction) Object() client.Object {
	return r.object
}

func (r testTransaction) TransactionType() models.TransactionType {
	return r.transactionType
}
