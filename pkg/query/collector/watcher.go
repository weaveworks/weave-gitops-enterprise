package collector

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/reconciler"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewWatcher(clusterName string, cfg *rest.Config, kinds []configuration.ObjectKind, objectChannel chan []models.ObjectTransaction, log logr.Logger) (manager.Manager, error) {
	scheme := runtime.NewScheme()
	for _, objectKind := range kinds {
		if err := objectKind.AddToSchemeFunc(scheme); err != nil {
			return nil, fmt.Errorf("cannot create runtime scheme: %w", err)
		}
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		Logger:             log,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create controller manager: %w", err)
	}

	process := func(tx models.ObjectTransaction) error {
		objectChannel <- []models.ObjectTransaction{tx}
		return nil
	}

	// create reconciler for kinds
	for _, kind := range kinds {
		rec, err := reconciler.NewReconciler(clusterName, kind, mgr.GetClient(), process, log)
		if err != nil {
			return nil, fmt.Errorf("cannot create reconciler: %w", err)
		}
		err = rec.Setup(mgr)
		if err != nil {
			return nil, fmt.Errorf("cannot setup reconciler: %w", err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("cannot setup reconciler: %w", err)
	}

	return mgr, nil
}

// Here so it can be used in stop watcher hooks

// NewDeleteAllTransaction returns an object transaction that says
// "delete everything in this cluster", for particular collectors to
// process.
func NewDeleteAllTransaction(clusterName string) models.ObjectTransaction {
	return deleteAllTransaction(clusterName)
}

type deleteAllTransaction string

func (t deleteAllTransaction) ClusterName() string {
	return string(t)
}

func (t deleteAllTransaction) Object() client.Object {
	return nil
}

func (t deleteAllTransaction) String() string {
	return fmt.Sprintf("%s/%s", string(t), t.TransactionType())
}

func (t deleteAllTransaction) TransactionType() models.TransactionType {
	return models.TransactionTypeDeleteAll
}
