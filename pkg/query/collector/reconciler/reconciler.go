package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/logger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Reconciler
type Reconciler interface {
	Setup(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

type ProcessFunc func(models.ObjectTransaction) error

func NewReconciler(clusterName string, objectKind configuration.ObjectKind, client client.Client, process ProcessFunc, log logr.Logger) (Reconciler, error) {

	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if err := objectKind.Validate(); err != nil {
		return nil, fmt.Errorf("invalid object kind:%w", err)
	}

	return &GenericReconciler{
		objectKind:  objectKind,
		client:      client,
		processFunc: process,
		log:         log.WithName("query-collector-reconciler"),
		debug:       log.WithName("query-collector-reconciler").V(logger.LogLevelDebug),
		clusterName: clusterName,
	}, nil
}

type GenericReconciler struct {
	processFunc ProcessFunc
	client      client.Client
	objectKind  configuration.ObjectKind
	debug       logr.Logger
	log         logr.Logger
	clusterName string
}

func (g GenericReconciler) Setup(mgr ctrl.Manager) error {
	clientObject := g.objectKind.NewClientObjectFunc()
	err := ctrl.NewControllerManagedBy(mgr).
		For(clientObject).
		Complete(&g)
	if err != nil {
		return err
	}
	g.log.Info(fmt.Sprintf("reconciler added for gvk: %s", g.objectKind.Gvk))

	return nil
}

func (r *GenericReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	clientObject := r.objectKind.NewClientObjectFunc()
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
		retentionPolicy: r.objectKind.RetentionPolicy,
	}

	r.debug.Info("object transaction received", "transaction", tx.String())

	return ctrl.Result{}, r.processFunc(tx)
}

type transaction struct {
	clusterName     string
	object          client.Object
	transactionType models.TransactionType
	retentionPolicy configuration.RetentionPolicy
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

func (r transaction) RetentionPolicy() configuration.RetentionPolicy {
	return r.retentionPolicy
}
