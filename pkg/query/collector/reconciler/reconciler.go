package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Reconciler
type Reconciler interface {
	Setup(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

func NewReconciler(kind string, client client.Client, objectsChannel chan []models.ObjectRecord, logger logr.Logger) (Reconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if kind == "" {
		return nil, fmt.Errorf("invalid kind")
	}

	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objects channel")
	}

	return &GenericReconciler{
		kind:           kind,
		client:         client,
		objectsChannel: objectsChannel,
		log:            logger,
	}, nil
}

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type GenericReconciler struct {
	objectsChannel chan []models.ObjectRecord
	client         client.Client
	kind           string
	log            logr.Logger
}

func (g GenericReconciler) Setup(mgr ctrl.Manager) error {

	clientObject, err := GetClientObjectByKind(g.kind)
	if err != nil {
		return fmt.Errorf("could not get object client: %w", err)
	}
	err = ctrl.NewControllerManagedBy(mgr).For(clientObject).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(&g)
	if err != nil {
		return err
	}
	g.log.Info(fmt.Sprintf("reconciler added for kind: %s", g.kind))

	return nil
}

func GetClientObjectByKind(kind string) (client.Object, error) {
	switch kind {
	case v2beta1.HelmReleaseKind:
		return &v2beta1.HelmRelease{}, nil
	case v1beta2.KustomizationKind:
		return &v1beta2.Kustomization{}, nil
	default:
		return nil, fmt.Errorf("kind not supported: %s", kind)
	}
	return nil, fmt.Errorf("invalid")
}

// TODO add unit
func (r *GenericReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"resource", req.NamespacedName)

	clientObject, err := GetClientObjectByKind(r.kind)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not get client object: %w", err)
	}
	if err := r.client.Get(ctx, req.NamespacedName, clientObject); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("resource retrieved")
	//TODO manage error
	r.objectsChannel <- []models.ObjectRecord{record{
		clusterName: "change",
		object:      clientObject,
	}}
	log.Info("resource notified")
	return ctrl.Result{}, nil
}

type record struct {
	clusterName string
	object      client.Object
}

func (r record) ClusterName() string {
	return r.clusterName
}

func (r record) Object() client.Object {
	return r.object
}
