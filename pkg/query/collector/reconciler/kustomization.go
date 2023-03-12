package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KustomizationReconciler runs the `reconcile` loop for the watcher.
type KustomizationReconciler struct {
	objectsChannel chan []models.Object
	client         client.Client
}

func NewKustomizationReconciler(
	client client.Client, objectsChannel chan []models.Object, log logr.Logger) (*KustomizationReconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objects channel")
	}

	//TODO add some Id for better instrumentation
	log.Info("creating kustomization reconciler")

	// add reconciler for kustomization
	return &KustomizationReconciler{
		objectsChannel: objectsChannel,
		client:         client,
	}, nil

}

func (r *KustomizationReconciler) Setup(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta2.Kustomization{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

// TODO add unit
// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
func (r *KustomizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"kustomizationRelease", req.NamespacedName)

	// get kustomization release
	var kustomization v1beta2.Kustomization
	if err := r.client.Get(ctx, req.NamespacedName, &kustomization); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Examine if the object is under deletion
	if !kustomization.ObjectMeta.GetDeletionTimestamp().IsZero() {
		log.Info("kustomization release delete")
		return r.notify(ctx, kustomization, "delete")
	}

	//TODO return type might not be correct
	return r.notify(ctx, kustomization, "update")
}

// TODO add unit
func (r *KustomizationReconciler) notify(ctx context.Context, kustomization v1beta2.Kustomization, operation string) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"kustomization", kustomization)

	//TODO add operation
	objectRecord := models.Object{
		Cluster:   "change",
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Kind:      kustomization.Kind,
		Operation: operation,
		//TODO conditions are multiple
		Status:  kustomization.Status.Conditions[0].String(),
		Message: kustomization.Status.Conditions[0].Message,
	}

	r.objectsChannel <- []models.Object{objectRecord}

	log.Info("notified operation")
	return ctrl.Result{}, nil
}
