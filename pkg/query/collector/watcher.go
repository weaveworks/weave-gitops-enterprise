package collector

import (
	"context"
	"fmt"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1beta2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/reconciler"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	Kinds         []schema.GroupVersionKind
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
	Start(ctx context.Context, log logr.Logger) error
	Status() (string, error)
	Stop() error
}

type DefaultWatcher struct {
	kinds             []schema.GroupVersionKind
	watcherManager    manager.Manager
	scheme            *runtime.Scheme
	clusterRef        types.NamespacedName
	cluster           cluster.Cluster
	stopFn            context.CancelFunc
	log               logr.Logger
	status            ClusterWatchingStatus
	newWatcherManager WatcherManagerFunc
	objectsChannel    chan []models.ObjectTransaction
	// useProxy is a flag to indicate if the helm watcher should use the proxy
	// useProxy bool
}

type WatcherManagerFunc = func(opts WatcherManagerOptions) (manager.Manager, error)

type WatcherManagerOptions struct {
	Log            logr.Logger
	Rest           *rest.Config
	Kinds          []schema.GroupVersionKind
	ObjectsChannel chan []models.ObjectTransaction
	ManagerOptions manager.Options
	ClusterName    string
}

func (o WatcherManagerOptions) Validate() error {
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
		return nil, fmt.Errorf("cannot create controller manager: %v", err)
	}

	//create reconcilers for kinds
	for _, kind := range opts.Kinds {
		rec, err := reconciler.NewReconciler(opts.ClusterName, kind, mgr.GetClient(), opts.ObjectsChannel, opts.Log)
		if err != nil {
			return nil, fmt.Errorf("cannot create reconciler: %v", err)
		}
		err = rec.Setup(mgr)
		if err != nil {
			return nil, fmt.Errorf("cannot setup reconciler: %v", err)
		}

	}
	if err != nil {
		return nil, fmt.Errorf("cannot setup reconciler: %v", err)
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

	scheme, err := newDefaultScheme()
	if err != nil {
		return nil, fmt.Errorf("cannot crete default scheme: %w", err)
	}

	cluster, err := cluster.NewSingleCluster(opts.ClusterRef.String(), opts.ClientConfig, scheme)
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
	}, nil
}

func newDefaultScheme() (*runtime.Scheme, error) {
	sc := runtime.NewScheme()
	// if err := clientgoscheme.AddToScheme(sc); err != nil {
	// 	return nil, err
	// }
	if err := helmv2beta1.AddToScheme(sc); err != nil {
		return nil, err
	}
	if err := kustomizev1beta2.AddToScheme(sc); err != nil {
		return nil, err
	}
	if err := rbacv1.AddToScheme(sc); err != nil {
		return nil, err
	}
	return sc, nil
}

func (w *DefaultWatcher) Start(ctx context.Context, log logr.Logger) error {
	w.log = log.WithName("watcher")

	ctx, cancel := context.WithCancel(ctx)
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
		ManagerOptions: ctrl.Options{
			Scheme:             w.scheme,
			Logger:             w.log,
			LeaderElection:     false,
			MetricsBindAddress: "0",
		},
	}

	w.watcherManager, err = w.newWatcherManager(opts)
	if err != nil {
		return fmt.Errorf("cannot create watcher manager: %v", err)
	}

	go func() {
		if err := w.watcherManager.Start(ctx); err != nil {
			log.Error(err, "cannot start watcher")
		}
	}()

	w.status = ClusterWatchingStarted

	return nil
}

func (w *DefaultWatcher) Stop() error {
	if w.stopFn == nil {
		return fmt.Errorf("Stop function not set yet")
	}
	w.stopFn()
	return nil
}

func (w *DefaultWatcher) Status() (string, error) {
	return string(w.status), nil
}
