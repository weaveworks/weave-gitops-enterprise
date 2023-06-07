package collector

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/reconciler"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ClusterWatchingStatus string

const (
	ClusterWatchingStarted ClusterWatchingStatus = "started"
	ClusterWatchingStopped ClusterWatchingStatus = "stopped"
)

type WatcherOptions struct {
	Log           logr.Logger
	ObjectChannel chan []models.ObjectTransaction
	ClusterRef    types.NamespacedName
	ClientConfig  *rest.Config
	Kinds         []configuration.ObjectKind
	ManagerFunc   WatcherManagerFunc
}

func (o WatcherOptions) Validate() error {
	if o.ClientConfig == nil {
		return fmt.Errorf("invalid config")
	}

	if o.ClusterRef.Name == "" || o.ClusterRef.Namespace == "" {
		return fmt.Errorf("clusterName name or namespace is empty")
	}

	if o.ManagerFunc == nil {
		return fmt.Errorf("invalid manager func")
	}

	return nil
}

type Watcher interface {
	Start() error
	Stop() error
	Status() (string, error)
}

type DefaultWatcher struct {
	kinds             []configuration.ObjectKind
	watcherManager    manager.Manager
	scheme            *runtime.Scheme
	clusterRef        types.NamespacedName
	cluster           cluster.Cluster
	log               logr.Logger
	status            ClusterWatchingStatus
	newWatcherManager WatcherManagerFunc
	objectsChannel    chan []models.ObjectTransaction
	stopFn            context.CancelFunc
}

type WatcherManagerFunc = func(opts WatcherManagerOptions) (manager.Manager, error)
type WatcherStopFunc = func(opts WatcherManagerOptions) (manager.Manager, error)

type WatcherManagerOptions struct {
	Log            logr.Logger
	Rest           *rest.Config
	Kinds          []configuration.ObjectKind
	ObjectsChannel chan []models.ObjectTransaction
	ManagerOptions manager.Options
	ClusterName    string
}

func (o WatcherManagerOptions) Validate() error {
	if o.ClusterName == "" {
		return fmt.Errorf("invalid cluster name")
	}

	if o.Rest == nil {
		return fmt.Errorf("invalid config")
	}

	if o.ManagerOptions.Scheme == nil {
		return fmt.Errorf("invalid scheme")
	}

	return nil
}

func defaultNewWatcherManager(opts WatcherManagerOptions) (manager.Manager, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	mgr, err := ctrl.NewManager(opts.Rest, ctrl.Options{
		Scheme:             opts.ManagerOptions.Scheme,
		Logger:             opts.Log,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create controller manager: %w", err)
	}

	//create reconciler for kinds
	for _, kind := range opts.Kinds {
		rec, err := reconciler.NewReconciler(opts.ClusterName, kind, mgr.GetClient(), opts.ObjectsChannel, opts.Log)
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

func NewWatcher(opts WatcherOptions) (Watcher, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if opts.ManagerFunc == nil {
		opts.ManagerFunc = defaultNewWatcherManager
	}

	scheme := runtime.NewScheme()
	for _, objectKind := range opts.Kinds {
		if err := objectKind.AddToSchemeFunc(scheme); err != nil {
			return nil, fmt.Errorf("cannot create runtime scheme: %w", err)
		}
	}

	cluster, err := cluster.NewSingleCluster(opts.ClusterRef.Name, opts.ClientConfig, scheme, kube.UserPrefixes{})
	if err != nil {
		return nil, err
	}
	return &DefaultWatcher{
		clusterRef:        opts.ClusterRef,
		cluster:           cluster,
		kinds:             opts.Kinds,
		scheme:            scheme,
		status:            ClusterWatchingStopped,
		newWatcherManager: opts.ManagerFunc,
		objectsChannel:    opts.ObjectChannel,
		log:               opts.Log,
	}, nil
}

func (w *DefaultWatcher) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.stopFn = cancel

	cfg, err := w.cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("invalid clusterName config")
	}

	opts := WatcherManagerOptions{
		Log:            w.log,
		Rest:           cfg,
		Kinds:          w.kinds,
		ObjectsChannel: w.objectsChannel,
		ClusterName:    w.cluster.GetName(),
		ManagerOptions: ctrl.Options{
			Scheme:             w.scheme,
			Logger:             w.log,
			LeaderElection:     false,
			MetricsBindAddress: "0",
		},
	}

	w.watcherManager, err = w.newWatcherManager(opts)
	if err != nil {
		return fmt.Errorf("cannot create watcher manager: %w", err)
	}

	// Manager.Start blocks until the context is canceled and unless
	// there was a problem starting a controller.
	go func() {
		if err := w.watcherManager.Start(ctx); err != nil {
			w.log.Error(err, "cannot start watcher", "cluster", w.cluster.GetName())
		}
	}()

	w.status = ClusterWatchingStarted

	return nil
}

// Default stop function will gracefully stop watching the cluste r
// and emmits a stop watcher watching event for upstreaming processing
func (w *DefaultWatcher) Stop() error {
	// stop watcher manager via cancelling context
	if w.stopFn == nil {
		return fmt.Errorf("cannot stop watcher without stop manager function")
	}
	w.stopFn()
	//emit delete all objects from watcher
	transactions := []models.ObjectTransaction{deleteAllTransaction{
		clusterName: w.cluster.GetName(),
	}}

	//TODO manage error
	w.objectsChannel <- transactions
	w.status = ClusterWatchingStopped
	return nil
}

func (w *DefaultWatcher) Status() (string, error) {
	return string(w.status), nil
}

type deleteAllTransaction struct {
	clusterName string
}

func (r deleteAllTransaction) ClusterName() string {
	return r.clusterName
}

func (r deleteAllTransaction) Object() client.Object {
	return nil
}

func (r deleteAllTransaction) String() string {
	return fmt.Sprintf("%s/%s", r.clusterName, r.TransactionType())
}

func (r deleteAllTransaction) TransactionType() models.TransactionType {
	return models.TransactionTypeDeleteAll
}
