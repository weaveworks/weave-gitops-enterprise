package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/logger"
	"k8s.io/apimachinery/pkg/api/errors"
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

var kindsWithoutFinalizers = map[string]bool{
	"ClusterRoleBinding": true,
	"ClusterRole":        true,
	"RoleBinding":        true,
	"Role":               true,
}

func NewReconciler(clusterName string, objectKind configuration.ObjectKind, client client.Client, process ProcessFunc, log logr.Logger) (Reconciler, error) {
	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}

	if objectKind.Gvk.Kind == "" {
		return nil, fmt.Errorf("missing gvk")
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
		// If the object being reconciled has no finalizers, then on deletion we will receive a reconcile request
		// for this object only after the object was deleted (not before deletion, as in the case of objects with finalizers)
		// and `client.Get` will return a `NotFound` error.
		// Thus, if a `NotFound` error is returned for an object whose Kind is included
		// in the list of known `Kinds` of objects, which do not have finalizers by default,
		// we can infer that the object was deleted.
		_, ok := kindsWithoutFinalizers[r.objectKind.Gvk.Kind]

		if !errors.IsNotFound(err) || !ok {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		clientObject.SetName(req.Name)
		clientObject.SetNamespace(req.Namespace)
		clientObject.GetObjectKind().SetGroupVersionKind(r.objectKind.Gvk)

		tx := transaction{
			clusterName:     r.clusterName,
			object:          clientObject,
			transactionType: models.TransactionTypeDelete,
			retentionPolicy: r.objectKind.RetentionPolicy,
			config:          r.objectKind,
		}

		r.debug.Info("object transaction received", "transaction", tx.String())

		return ctrl.Result{}, r.processFunc(tx)
	}

	if r.objectKind.FilterFunc != nil {
		retain := r.objectKind.FilterFunc(clientObject)
		if !retain {
			r.debug.Info("skipping object", "kind", clientObject.GetObjectKind(), "name", clientObject.GetName())
			return ctrl.Result{}, nil
		}
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
		config:          r.objectKind,
	}

	r.debug.Info("object transaction received", "transaction", tx.String())

	return ctrl.Result{}, r.processFunc(tx)
}

type transaction struct {
	clusterName     string
	object          client.Object
	config          configuration.ObjectKind
	transactionType models.TransactionType
	retentionPolicy configuration.RetentionPolicy
}

func (r transaction) ClusterName() string {
	return r.clusterName
}

func (r transaction) Object() models.NormalizedObject {
	return models.NewNormalizedObject(r.object, r.config)
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
