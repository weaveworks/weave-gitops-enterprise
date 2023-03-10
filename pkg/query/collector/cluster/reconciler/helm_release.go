package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/cluster/store"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	//TODO change me to function add, update or delete function
	//reconciler should not have access to the store
	store  store.Store
	client client.Client
}

func NewHelmWatcherReconciler(
	client client.Client, store store.Store, log logr.Logger) (*HelmWatcherReconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	//TODO add some Id for better instrumentation
	log.Info("creating helm reconciler")

	// add reconciler for helm releases
	return &HelmWatcherReconciler{
		store:  store,
		client: client,
	}, nil

}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v2beta1.HelmRelease{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

// TODO add unit
// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"helmRelease", req.NamespacedName)

	// get helm release
	var helmRelease v2beta1.HelmRelease
	if err := r.client.Get(ctx, req.NamespacedName, &helmRelease); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Examine if the object is under deletion
	if !helmRelease.ObjectMeta.GetDeletionTimestamp().IsZero() {
		log.Info("helm release delete")
		return r.reconcileDelete(ctx, helmRelease)
	}

	return r.reconcileAddOrUpdate(ctx, helmRelease)
}

// TODO add unit
func (r *HelmWatcherReconciler) reconcileDelete(ctx context.Context, helmRelease v2beta1.HelmRelease) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"helmRelease", helmRelease)

	//adapt to document
	document := store.Document{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Kind:      helmRelease.Kind,
	}

	if err := r.store.Delete(ctx, document); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("delete helm release document from store")
	return ctrl.Result{}, nil
}

// TODO add unit
func (r *HelmWatcherReconciler) reconcileAddOrUpdate(ctx context.Context, helmRelease v2beta1.HelmRelease) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"helmRelease", helmRelease)

	//adapt to document
	document := store.Document{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Kind:      helmRelease.Kind,
	}

	id, err := r.store.Add(ctx, document)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("stored helm release document with id", id)

	return ctrl.Result{}, nil
}
