package collector

import (
	"context"
	"fmt"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/reconciler"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
	ClusterRef   types.NamespacedName
	ClientConfig *rest.Config
	Kinds        []schema.GroupVersionKind
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
	newWatcherManager newWatcherManagerFunc
	objectsChannel    chan []models.ObjectRecord
	// useProxy is a flag to indicate if the helm watcher should use the proxy
	useProxy bool
}

type newWatcherManagerFunc = func(config *rest.Config, kinds []schema.GroupVersionKind, objectsChannel chan []models.ObjectRecord, options manager.Options) (manager.Manager, error)

func defaultNewWatcherManager(config *rest.Config, kinds []schema.GroupVersionKind, objectsChannel chan []models.ObjectRecord, options manager.Options) (manager.Manager, error) {

	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if options.Scheme == nil {
		return nil, fmt.Errorf("invalid scheme")
	}

	log := options.Logger

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             options.Scheme,
		Logger:             options.Logger,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create controller manager: %v", err)
	}

	//create reconcilers for kinds
	for _, kind := range kinds {
		rec, err := reconciler.NewReconciler(kind, mgr.GetClient(), objectsChannel, log)
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
	log.Info("controller manager created")
	return mgr, nil
}

func NewWatcher(opts WatcherOptions, newManagerFunc newWatcherManagerFunc,
	objectsChannel chan []models.ObjectRecord, log logr.Logger) (*DefaultWatcher, error) {

	if opts.ClientConfig == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if opts.ClusterRef.Name == "" || opts.ClusterRef.Namespace == "" {
		return nil, fmt.Errorf("clusterName name or namespace is empty")
	}

	if len(opts.Kinds) == 0 {
		return nil, fmt.Errorf("at least one kind is required")
	}

	if newManagerFunc == nil {
		newManagerFunc = defaultNewWatcherManager
		log.Info("using default manager function")
	}

	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objects channel")
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
		newWatcherManager: newManagerFunc,
		objectsChannel:    objectsChannel,
	}, nil
}

// TODO we could make this configurable
func newDefaultScheme() (*runtime.Scheme, error) {
	sc := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sc); err != nil {
		return &runtime.Scheme{}, err
	}
	if err := v2beta1.AddToScheme(sc); err != nil {
		return &runtime.Scheme{}, err
	}
	if err := v1beta2.AddToScheme(sc); err != nil {
		return &runtime.Scheme{}, err
	}
	if err := rbacv1.AddToScheme(sc); err != nil {
		return &runtime.Scheme{}, err
	}
	return sc, nil
}

func (w *DefaultWatcher) Start(ctx context.Context, log logr.Logger) error {
	if ctx == nil {
		return fmt.Errorf("invalid context")
	}
	w.log = log.WithName("watcher")

	ctx, cancel := context.WithCancel(ctx)
	w.stopFn = cancel

	cfg, err := w.cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("invalid clusterName config")
	}

	w.watcherManager, err = w.newWatcherManager(cfg, w.kinds, w.objectsChannel, ctrl.Options{
		Scheme:             w.scheme,
		Logger:             w.log,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return fmt.Errorf("cannot create watcher manager: %v", err)
	}
	w.log.Info("watcher manager created")
	go func() {
		if err := w.watcherManager.Start(ctx); err != nil {
			log.Error(err, "cannot start watcher")
		}
	}()

	w.status = ClusterWatchingStarted
	w.log.Info("watcher with helm reconciler started")
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
