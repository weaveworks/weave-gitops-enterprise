package objectscollector

import (
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestObjectsCollector_defaultProcessRecords(t *testing.T) {
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
			name: "can process non-empty record collection with no errors",
			objectRecords: []models.ObjectTransaction{
				testutils.NewObjectTransaction("anyCluster", testutils.NewHelmRelease("createdOrUpdatedHelmRelease", clusterName), models.TransactionTypeUpsert),
				testutils.NewObjectTransaction("anyCluster", testutils.NewHelmRelease("deletedHelmRelease1", clusterName, func(hr *v2beta1.HelmRelease) {
					now := metav1.Now()
					hr.DeletionTimestamp = &now
				}), models.TransactionTypeDelete),
				testutils.NewObjectTransaction("anyCluster2", testutils.NewHelmRelease("deletedHelmRelease2", clusterName, func(hr *v2beta1.HelmRelease) {
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
			err := processRecords(tt.objectRecords, fakeStore, fakeIndex, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(fakeStore.StoreObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeUpsert]))
			g.Expect(fakeStore.DeleteObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDelete]))
			g.Expect(fakeStore.DeleteAllObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDeleteAll]))
		})
	}
}

func TestObjectsCollector_removeAll(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}
	fakeIndex := &storefakes.FakeIndexWriter{}

	//setup data
	clusterName := "anyCluster"

	tx := []models.ObjectTransaction{
		&transaction{
			clusterName:     clusterName,
			object:          testutils.NewHelmRelease("anyHelmRelease", clusterName),
			transactionType: models.TransactionTypeDeleteAll,
		},
	}

	err := processRecords(tx, fakeStore, fakeIndex, log)
	g.Expect(err).To(BeNil())

	g.Expect(fakeStore.DeleteAllObjectsCallCount()).To(Equal(1))

	_, query := fakeIndex.RemoveByQueryArgsForCall(0)
	g.Expect(query).To(Equal("+cluster:anyCluster"))

}

type transaction struct {
	clusterName     string
	object          client.Object
	transactionType models.TransactionType
}

func (t *transaction) Object() client.Object {
	return t.object
}

func (t *transaction) ClusterName() string {
	return t.clusterName
}

func (t *transaction) TransactionType() models.TransactionType {
	return t.transactionType
}
