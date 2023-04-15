package reconciler

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/logger"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Reconciler
type Reconciler interface {
	Setup(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

func NewReconciler(clusterName string, gvk schema.GroupVersionKind, client client.Client, objectsChannel chan []models.ObjectTransaction, logger logr.Logger) (Reconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if gvk.Empty() {
		return nil, fmt.Errorf("invalid gvk")
	}

	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objects channel")
	}

	return &GenericReconciler{
		gvk:            gvk,
		client:         client,
		objectsChannel: objectsChannel,
		log:            logger.WithName("query-collector-reconciler"),
		clusterName:    clusterName,
	}, nil
}

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type GenericReconciler struct {
	objectsChannel chan []models.ObjectTransaction
	client         client.Client
	gvk            schema.GroupVersionKind
	log            logr.Logger
	clusterName    string
}

func (g GenericReconciler) Setup(mgr ctrl.Manager) error {
	clientObject, err := getClientObjectByKind(g.gvk)
	if err != nil {
		return fmt.Errorf("could not get object client: %w", err)
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(clientObject).
		Complete(&g)
	if err != nil {
		return err
	}
	g.log.Info(fmt.Sprintf("reconciler added for gvk: %s", g.gvk))

	return nil
}

func getClientObjectByKind(gvk schema.GroupVersionKind) (client.Object, error) {
	switch gvk.Kind {
	case v2beta1.HelmReleaseKind:
		return &v2beta1.HelmRelease{}, nil
	case v1beta2.KustomizationKind:
		return &v1beta2.Kustomization{}, nil
	case "ClusterRole":
		return &rbacv1.ClusterRole{}, nil
	case "Role":
		return &rbacv1.Role{}, nil
	case "ClusterRoleBinding":
		return &rbacv1.ClusterRoleBinding{}, nil
	case "RoleBinding":
		return &rbacv1.RoleBinding{}, nil
	default:
		return nil, fmt.Errorf("gvk not supported: %s", gvk.Kind)
	}
}

func (r *GenericReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	clientObject, err := getClientObjectByKind(r.gvk)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not get client object: %w", err)
	}
	if err := r.client.Get(ctx, req.NamespacedName, clientObject); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	txType := models.TransactionTypeUpsert

	if !clientObject.GetDeletionTimestamp().IsZero() {
		txType = models.TransactionTypeDelete
	}

	tx := transaction{
		clusterName:     r.clusterName,
		object:          clientObject,
		transactionType: txType,
	}

	transactions := []models.ObjectTransaction{tx}

	r.log.V(logger.LogLevelDebug).Info("object transaction received", "transaction", tx.String())

	//TODO manage error
	r.objectsChannel <- transactions

	return ctrl.Result{}, nil
}

type transaction struct {
	clusterName     string
	object          client.Object
	transactionType models.TransactionType
}

func (r transaction) ClusterName() string {
	return r.clusterName
}

func (r transaction) Object() client.Object {
	return r.object
}

func (r transaction) TransactionType() models.TransactionType {
	return r.transactionType
}

func (r transaction) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", r.clusterName, r.object.GetNamespace(), r.object.GetName(), r.transactionType)
}
