package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	objectsChannel chan []models.Object
	client         client.Client
}

func NewHelmWatcherReconciler(
	client client.Client, objectsChannel chan []models.Object, log logr.Logger) (*HelmWatcherReconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objectsChannel")
	}

	//TODO add some Id for better instrumentation
	log.Info("creating helm reconciler")

	// add reconciler for helm releases
	return &HelmWatcherReconciler{
		objectsChannel: objectsChannel,
		client:         client,
	}, nil

}

func (r *HelmWatcherReconciler) Setup(mgr ctrl.Manager) error {
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
		return r.notify(ctx, helmRelease, "delete")
	}

	//TODO return type might not be correct
	return r.notify(ctx, helmRelease, "update")
}

// TODO add unit
func (r *HelmWatcherReconciler) notify(ctx context.Context, helmRelease v2beta1.HelmRelease, operation string) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"helmRelease", helmRelease)

	//TODO add operation
	objectRecord := models.Object{
		Cluster:   "change",
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Kind:      helmRelease.Kind,
		Operation: operation,
		//TODO conditions are multiple
		Status:  helmRelease.Status.Conditions[0].String(),
		Message: helmRelease.Status.Conditions[0].Message,
	}

	r.objectsChannel <- []models.Object{objectRecord}

	log.Info("notified operation")
	return ctrl.Result{}, nil
}
