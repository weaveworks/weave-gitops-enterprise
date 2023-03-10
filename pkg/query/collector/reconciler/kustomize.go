package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KustomizationWatcherReconciler runs the `reconcile` loop for the watcher.
type KustomizationWatcherReconciler struct {
	//TODO change me to function add, update or delete function
	//reconciler should not have access to the store
	store  store.Store
	client client.Client
}

func NewKustomizeWatcherReconciler(
	client client.Client, store store.Store, log logr.Logger) (*KustomizationWatcherReconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	//TODO add some Id for better instrumentation
	log.Info("creating helm reconciler")

	// add reconciler for helm releases
	return &KustomizationWatcherReconciler{
		store:  store,
		client: client,
	}, nil

}

func (r *KustomizationWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta2.Kustomization{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

// TODO add unit
// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
func (r *KustomizationWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"kustomization", req.NamespacedName)

	// get helm release
	var kustomization v1beta2.Kustomization
	if err := r.client.Get(ctx, req.NamespacedName, &kustomization); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Examine if the object is under deletion
	if !kustomization.ObjectMeta.GetDeletionTimestamp().IsZero() {
		log.Info("delete operation requested")
		return r.reconcileDelete(ctx, kustomization)
	}

	log.Info("add or update operation requested")
	return r.reconcileAddOrUpdate(ctx, kustomization)
}

// TODO add unit
func (r *KustomizationWatcherReconciler) reconcileDelete(ctx context.Context, kustomization v1beta2.Kustomization) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"kustomization", kustomization)

	//adapt to document
	document := store.Document{
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Kind:      kustomization.Kind,
	}

	if err := r.store.Delete(ctx, document); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("delete document from store")
	return ctrl.Result{}, nil
}

// TODO add unit
func (r *KustomizationWatcherReconciler) reconcileAddOrUpdate(ctx context.Context, kustomization v1beta2.Kustomization) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"kustomization", kustomization)

	//adapt to document
	document := store.Document{
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Kind:      kustomization.Kind,
	}

	id, err := r.store.Add(ctx, document)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("stored document with id", id)

	return ctrl.Result{}, nil
}
