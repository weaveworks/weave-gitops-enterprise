package reconciler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

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
		name       string
		objectKind configuration.ObjectKind
		client     client.Client
		process    func(models.ObjectTransaction) error
		errPattern string
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
			name:       "could create reconciler with valid arguments",
			client:     fakeClient,
			objectKind: configuration.HelmReleaseObjectKind,
			process:    func(models.ObjectTransaction) error { return nil },
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewReconciler("test-cluster", tt.objectKind, tt.client, tt.process, log)
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
		t.Fatalf("could not add v2beta1 to scheme: %v", err)
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
	process := func(tx models.ObjectTransaction) error {
		return nil
	}

	reconciler, err := NewReconciler("test-cluster", configuration.HelmReleaseObjectKind, fakeClient, process, logger)
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
	deletedHelmRelease := testutils.NewHelmRelease("deletedHelmRelease", clusterName, func(hr *v2beta1.HelmRelease) {
		now := metav1.Now()
		hr.DeletionTimestamp = &now
		hr.Finalizers = append(hr.Finalizers, "finalizers.fluxcd.io")
	})
	deletedClusterRoleWithRules := testutils.NewClusterRole("deletedClusterRoleWithRules", true, func(cr *rbacv1.ClusterRole) {
		now := metav1.Now()
		cr.DeletionTimestamp = &now
		cr.Finalizers = append(cr.Finalizers, "kubernetes")
	})
	deletedClusterRoleWithoutRules := testutils.NewClusterRole("deletedClusterRoleWithoutRules", false, func(cr *rbacv1.ClusterRole) {
		now := metav1.Now()
		cr.DeletionTimestamp = &now
		cr.Finalizers = append(cr.Finalizers, "kubernetes")
	})
	deletedClusterRoleManuallyConstructed := testutils.NewClusterRole("deletedClusterRoleManuallyConstructed", false)
	objects := []runtime.Object{createdOrUpdatedHelmRelease, deletedHelmRelease, deletedClusterRoleWithRules, deletedClusterRoleWithoutRules}

	//setup reconciler
	s := runtime.NewScheme()
	if err := v2beta1.AddToScheme(s); err != nil {
		t.Fatalf("could not add v2beta1 to scheme: %v", err)
	}
	if err := rbacv1.AddToScheme(s); err != nil {
		t.Fatalf("could not add rbacv1 to scheme: %v", err)
	}
	fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objects...).WithScheme(s).Build()

	tests := []struct {
		name         string
		object       client.Object
		kind         configuration.ObjectKind
		request      ctrl.Request
		expectedTx   transaction
		errPattern   string
		shouldHaveTX bool
	}{
		{
			name:   "can reconcile created or updated resource requests",
			object: createdOrUpdatedHelmRelease,
			kind:   configuration.HelmReleaseObjectKind,
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
				config:          configuration.HelmReleaseObjectKind,
			},
			errPattern:   "",
			shouldHaveTX: true,
		},
		{
			name:   "can reconcile delete resource requests",
			object: deletedHelmRelease,
			kind:   configuration.HelmReleaseObjectKind,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      deletedHelmRelease.GetName(),
					Namespace: deletedHelmRelease.GetNamespace(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          deletedHelmRelease,
				transactionType: models.TransactionTypeDelete,
				config:          configuration.HelmReleaseObjectKind,
			},
			errPattern:   "",
			shouldHaveTX: true,
		},
		{
			name:   "can reconcile delete requests for existing clusterrole with rules",
			object: deletedClusterRoleWithRules,
			kind:   configuration.ClusterRoleObjectKind,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: deletedClusterRoleWithRules.GetName(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          deletedClusterRoleWithRules,
				transactionType: models.TransactionTypeDelete,
				config:          configuration.ClusterRoleObjectKind,
			},
			errPattern:   "",
			shouldHaveTX: true,
		},
		{
			name:   "can reconcile delete requests for existing clusterrole without rules",
			object: deletedClusterRoleWithoutRules,
			kind:   configuration.ClusterRoleObjectKind,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: deletedClusterRoleWithoutRules.GetName(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          deletedClusterRoleWithoutRules,
				transactionType: models.TransactionTypeDelete,
				config:          configuration.ClusterRoleObjectKind,
			},
			errPattern:   "",
			shouldHaveTX: true,
		},
		{
			name:   "can reconcile delete requests for already deleted clusterrole",
			object: deletedClusterRoleManuallyConstructed,
			kind:   configuration.ClusterRoleObjectKind,
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: deletedClusterRoleManuallyConstructed.GetName(),
				},
			},
			expectedTx: transaction{
				clusterName:     "anyCluster",
				object:          deletedClusterRoleManuallyConstructed,
				transactionType: models.TransactionTypeDelete,
				config:          configuration.ClusterRoleObjectKind,
			},
			errPattern:   "",
			shouldHaveTX: true,
		},
		{
			name:   "does not retain objects that do not pass the filter func",
			object: createdOrUpdatedHelmRelease,
			kind: configuration.ObjectKind{
				Gvk: v2beta1.GroupVersion.WithKind(v2beta1.HelmReleaseKind),
				NewClientObjectFunc: func() client.Object {
					return &v2beta1.HelmRelease{}
				},
				AddToSchemeFunc: v2beta1.AddToScheme,
				FilterFunc: func(object client.Object) bool {
					return false
				},
				StatusFunc: func(_ client.Object, _ configuration.ObjectKind) (configuration.ObjectStatus, error) {
					return configuration.NoStatus, nil
				},
				MessageFunc: func(_ client.Object, _ configuration.ObjectKind) (string, error) {
					return "", nil
				},
				Category: "test",
			},
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      createdOrUpdatedHelmRelease.GetName(),
					Namespace: createdOrUpdatedHelmRelease.GetNamespace(),
				},
			},
			expectedTx:   transaction{},
			shouldHaveTX: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reconcileError error
			var txns []models.ObjectTransaction
			process := func(tx models.ObjectTransaction) error {
				txns = append(txns, tx)
				return nil
			}
			reconciler, err := NewReconciler(clusterName, tt.kind, fakeClient, process, logger)
			g.Expect(err).To(BeNil())
			g.Expect(reconciler).NotTo(BeNil())
			_, reconcileError = reconciler.Reconcile(ctx, tt.request)
			if tt.errPattern != "" {
				g.Expect(reconcileError).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconcileError).To(BeNil())

			if !tt.shouldHaveTX {
				g.Expect(len(txns)).To(Equal(0))
				return
			}

			assertObjectTransaction(t, txns[0], tt.expectedTx)
		})
	}
}

func assertObjectTransaction(t *testing.T, actual models.ObjectTransaction, expected models.ObjectTransaction) {
	assert.Equal(t, expected.ClusterName(), actual.ClusterName(), "different cluster")
	assert.Equal(t, expected.TransactionType(), actual.TransactionType(), "different tx type")
	assert.Equal(t, expected.Object().GetName(), actual.Object().GetName(), "different object")

	cat, err := actual.Object().GetCategory()
	if err != nil {
		t.Fatalf("could not get category: %v", err)
	}

	expectedCat, err := expected.Object().GetCategory()
	if err != nil {
		t.Fatalf("could not get category: %v", err)
	}

	// This is effectively a check for NormalizedObject,config being populated.
	assert.Equal(t, cat, expectedCat, "different category")
}
