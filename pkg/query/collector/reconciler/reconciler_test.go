package reconciler

import (
	"context"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	s := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	tests := []struct {
		name           string
		objectKind     configuration.ObjectKind
		client         client.Client
		objectsChannel chan []models.ObjectTransaction
		errPattern     string
	}{
		{
			name:       "cannot create reconciler without client",
			errPattern: "invalid client",
		},
		{
			name:       "cannot create reconciler without gvk",
			client:     fakeClient,
			errPattern: "missing gvk",
		},
		{
			name:       "cannot create reconciler without object channel",
			client:     fakeClient,
			objectKind: configuration.HelmReleaseObjectKind,
			errPattern: "invalid objects channel",
		},
		{
			name:           "could create reconciler with valid arguments",
			client:         fakeClient,
			objectKind:     configuration.HelmReleaseObjectKind,
			objectsChannel: make(chan []models.ObjectTransaction),
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewReconciler("test-cluster", tt.objectKind, tt.client, tt.objectsChannel, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
		})
	}
}

func TestSetup(t *testing.T) {
	g := NewGomegaWithT(t)
	s := runtime.NewScheme()
	if err := v2beta1.AddToScheme(s); err != nil {
		t.Fatalf("could not add v2beta1 to scheme: %w", err)
	}
	logger := testr.New(t)
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	fakeManager, err := kubefakes.NewControllerManager(&rest.Config{
		Host: "http://idontexist",
	}, ctrl.Options{
		Logger: logger,
		Scheme: s,
	})
	g.Expect(err).To(BeNil())
	g.Expect(fakeManager).NotTo(BeNil())
	objectsChannel := make(chan []models.ObjectTransaction)

	reconciler, err := NewReconciler("test-cluster", configuration.HelmReleaseObjectKind, fakeClient, objectsChannel, logger)
	g.Expect(err).To(BeNil())
	g.Expect(reconciler).NotTo(BeNil())

	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "can setup reconciler",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reconciler.Setup(fakeManager)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
		})
	}
}

func TestReconciler_Reconcile(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	logger := testr.New(t)

	clusterName := "anyCluster"
	//setup data
	createdOrUpdatedHelmRelease := testutils.NewHelmRelease("createdOrUpdatedHelmRelease", clusterName)
	deleteHelmRelease := testutils.NewHelmRelease("deletedHelmRelease", clusterName, func(hr *v2beta1.HelmRelease) {
		now := metav1.Now()
		hr.DeletionTimestamp = &now
	})
	objects := []runtime.Object{createdOrUpdatedHelmRelease, deleteHelmRelease}

	//setup reconciler
	s := runtime.NewScheme()
	if err := v2beta1.AddToScheme(s); err != nil {
		t.Fatalf("could not add v2beta1 to scheme: %w", err)
	}
	fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objects...).WithScheme(s).Build()

	tests := []struct {
		name       string
		object     client.Object
		request    ctrl.Request
		expectedTx transaction
		errPattern string
	}{
		{
			name:   "can reconcile created or updated resource requests",
			object: createdOrUpdatedHelmRelease,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      createdOrUpdatedHelmRelease.GetName(),
					Namespace: createdOrUpdatedHelmRelease.GetNamespace(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          createdOrUpdatedHelmRelease,
				transactionType: models.TransactionTypeUpsert,
			},
			errPattern: "",
		},
		{
			name:   "can reconcile delete resource requests",
			object: deleteHelmRelease,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      deleteHelmRelease.GetName(),
					Namespace: deleteHelmRelease.GetNamespace(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          deleteHelmRelease,
				transactionType: models.TransactionTypeDelete,
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reconcileError error
			objectsChannel := make(chan []models.ObjectTransaction)
			defer close(objectsChannel)
			reconciler, err := NewReconciler(clusterName, configuration.HelmReleaseObjectKind, fakeClient, objectsChannel, logger)
			g.Expect(err).To(BeNil())
			g.Expect(reconciler).NotTo(BeNil())
			go func() {
				_, reconcileError = reconciler.Reconcile(ctx, tt.request)
			}()
			objectTransactions := <-objectsChannel
			if tt.errPattern != "" {
				g.Expect(reconcileError).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconcileError).To(BeNil())
			assertObjectTransaction(t, objectTransactions[0], tt.expectedTx)
		})
	}
}

func assertObjectTransaction(t *testing.T, actual models.ObjectTransaction, expected models.ObjectTransaction) {

	assert.Assert(t, expected.ClusterName() == actual.ClusterName(), "different cluster")
	assert.Assert(t, expected.TransactionType() == actual.TransactionType(), "different tx type")
	assert.Assert(t, expected.Object().GetName() == actual.Object().GetName(), "different object")
}
