package collector

import (
	"context"
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
)

type WatcherOptions struct {
	Log           logr.Logger
	ObjectChannel chan []models.ObjectTransaction
	ClientConfig  *rest.Config
	Kinds         []configuration.ObjectKind
}

func (o WatcherOptions) Validate() error {
	if o.ClientConfig == nil {
		return fmt.Errorf("invalid config")
	}
	return nil
}

type DefaultWatcher struct {
	clusterName    string
	manager        manager.Manager
	log            logr.Logger
	objectsChannel chan []models.ObjectTransaction
}

func newManager(clusterName string, cfg *rest.Config, kinds []configuration.ObjectKind, objectChannel chan []models.ObjectTransaction, log logr.Logger) (manager.Manager, error) {
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

func NewWatcher(clusterName string, opts WatcherOptions) (*DefaultWatcher, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	mgr, err := newManager(clusterName, opts.ClientConfig, opts.Kinds, opts.ObjectChannel, opts.Log)
	if err != nil {
		return nil, fmt.Errorf("could not create manager: %w", err)
	}

	return &DefaultWatcher{
		clusterName:    clusterName,
		manager:        mgr,
		status:         ClusterWatchingStopped,
		objectsChannel: opts.ObjectChannel,
		log:            opts.Log,
	}, nil
}

// Start the watcher; this will return only when there's an error, or
// the context is cancelled.
func (w *DefaultWatcher) Start(ctx context.Context) error {
	return w.manager.Start(ctx)
}
