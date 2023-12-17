package objectscollector

import (
	"testing"
	"time"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectsCollector_defaultProcessRecords(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)

	clusterName := "anyCluster"

	tests := []struct {
		name                  string
		objectRecords         []models.ObjectTransaction
		expectedStoreNumCalls map[models.TransactionType]int
		expectedObject        []models.Object
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
			fakeStore := &storefakes.FakeStore{}
			fakeIndex := &storefakes.FakeIndexWriter{}

			err := processRecords(tt.objectRecords, fakeStore, fakeIndex, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(fakeStore.StoreObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeUpsert]))
			g.Expect(fakeStore.DeleteObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDelete]))
			g.Expect(fakeStore.DeleteAllObjectsCallCount()).To(Equal(tt.expectedStoreNumCalls[models.TransactionTypeDeleteAll]))

			if tt.expectedObject != nil {
				opt := cmpopts.IgnoreFields(models.Object{}, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Category", "Unstructured")
				_, storeResult := fakeStore.StoreObjectsArgsForCall(0)

				diff := cmp.Diff(tt.expectedObject[0], storeResult[0], opt)

				if diff != "" {
					t.Errorf("unexpected result (-want +got):\n%s", diff)
				}

			}
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
			object:          models.NewNormalizedObject(testutils.NewHelmRelease("anyHelmRelease", clusterName), configuration.HelmReleaseObjectKind),
			transactionType: models.TransactionTypeDeleteAll,
		},
	}

	err := processRecords(tx, fakeStore, fakeIndex, log)
	g.Expect(err).To(BeNil())

	g.Expect(fakeStore.DeleteAllObjectsCallCount()).To(Equal(1))

	_, query := fakeIndex.RemoveByQueryArgsForCall(0)
	g.Expect(query).To(Equal("+cluster:anyCluster"))

}

func TestObjectsCollector_retention(t *testing.T) {
	g := NewWithT(t)
	log := logr.Discard()
	fakeStore := &storefakes.FakeStore{}
	fakeIndex := &storefakes.FakeIndexWriter{}

	//setup data
	clusterName := "anyCluster"

	retentionPolicy := configuration.RetentionPolicy(10 * time.Second)

	obj := models.NewNormalizedObject(
		testutils.NewHelmRelease("anyHelmRelease", clusterName),
		configuration.HelmReleaseObjectKind,
	)
	// Deleted an hour ago, retention policy is only 10 seconds.
	obj.SetDeletionTimestamp(&metav1.Time{Time: time.Now().Add(1 * -time.Hour)})

	obj2 := models.NewNormalizedObject(
		testutils.NewHelmRelease("anyHelmRelease2", clusterName),
		configuration.HelmReleaseObjectKind,
	)

	// Deleted 5 seconds ago, retention policy is 10 seconds.
	obj2.SetDeletionTimestamp(&metav1.Time{Time: time.Now().Add(5 * -time.Second)})

	// ACD don't have finalizers, and we don't get a deletionTimestamp.
	obj3 := models.NewNormalizedObject(
		testutils.NewAutomatedClusterDiscovery("anyACD", "default"),
		configuration.AutomatedClusterDiscoveryKind,
	)

	tx := []models.ObjectTransaction{
		&transaction{
			clusterName:     clusterName,
			object:          obj,
			transactionType: models.TransactionTypeDelete,
			retentionPolicy: retentionPolicy,
		},
		&transaction{
			clusterName:     clusterName,
			object:          obj2,
			transactionType: models.TransactionTypeDelete,
			retentionPolicy: retentionPolicy,
		},
		&transaction{
			clusterName:     clusterName,
			object:          obj3,
			transactionType: models.TransactionTypeDelete,
			// No retention policy here, should be deleted.
		},
	}

	err := processRecords(tx, fakeStore, fakeIndex, log)
	g.Expect(err).To(BeNil())

	// Remove the expired object.
	_, deleteResult := fakeStore.DeleteObjectsArgsForCall(0)
	g.Expect(deleteResult).To(HaveLen(2))
	g.Expect(deleteResult[0].Name).To(Equal("anyHelmRelease"))
	g.Expect(deleteResult[1].Name).To(Equal("anyACD"))

	// Keep the object that was deleted but not expired.
	_, storeResult := fakeStore.StoreObjectsArgsForCall(0)
	g.Expect(storeResult).To(HaveLen(1))
	g.Expect(storeResult[0].Name).To(Equal("anyHelmRelease2"))
}

type transaction struct {
	clusterName     string
	object          models.NormalizedObject
	transactionType models.TransactionType
	retentionPolicy configuration.RetentionPolicy
}

func (t *transaction) Object() models.NormalizedObject {
	return t.object
}

func (t *transaction) ClusterName() string {
	return t.clusterName
}

func (t *transaction) TransactionType() models.TransactionType {
	return t.transactionType
}

func (t *transaction) RetentionPolicy() configuration.RetentionPolicy {
	return t.retentionPolicy
}
