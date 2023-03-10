package cluster

import (
	"context"
	"fmt"
	"github.com/enekofb/collector/pkg/cluster/reconciler"
	"github.com/enekofb/collector/pkg/cluster/store"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type WatcherStatus string

const (
	WatcherStarted WatcherStatus = "started"
	WatcherStopped WatcherStatus = "stopped"
)

type WatcherOptions struct {
	ClusterRef   types.NamespacedName
	ClientConfig *rest.Config
	Kinds        []string
}

type Watcher interface {
	Start(ctx context.Context, log logr.Logger) error
	Status() (string, error)
	Stop() error
}

type DefaultWatcher struct {
	kinds             []string
	scheme            *runtime.Scheme
	clusterRef        types.NamespacedName
	cluster           cluster.Cluster
	stopFn            context.CancelFunc
	log               logr.Logger
	status            WatcherStatus
	newWatcherManager newWatcherManagerFunc
	watcherManager    manager.Manager
	store             store.Store
	// useProxy is a flag to indicate if the helm watcher should use the proxy
	useProxy bool
}

type newWatcherManagerFunc = func(config *rest.Config, kinds []string, store store.Store, options manager.Options) (manager.Manager, error)

func defaultNewWatcherManager(config *rest.Config, kinds []string, store store.Store, options manager.Options) (manager.Manager, error) {

	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if options.Scheme == nil {
		return nil, fmt.Errorf("invalid scheme")
	}

	if options.Logger.GetSink() == nil {
		return nil, fmt.Errorf("invalid logger")
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

	for _, kind := range kinds {
		err := addReconcilerByKind(kind, mgr, store, log)
		if err != nil {
			return nil, err
		}
		log.Info(fmt.Sprintf("reconciler added for kind: %s", kind))
	}

	log.Info("default controller manager created")

	return mgr, nil
}

func NewWatcher(opts WatcherOptions, newManagerFunc newWatcherManagerFunc, store store.Store, log logr.Logger) (*DefaultWatcher, error) {

	if opts.ClientConfig == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if opts.ClusterRef.Name == "" || opts.ClusterRef.Namespace == "" {
		return nil, fmt.Errorf("cluster name or namespace is empty")
	}

	if len(opts.Kinds) == 0 {
		return nil, fmt.Errorf("at least one kind is required")
	}

	if newManagerFunc == nil {
		newManagerFunc = defaultNewWatcherManager
		log.Info("using default manager function")
	}

	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	scheme, err := newScheme(opts.Kinds)
	if err != nil {
		return nil, err
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
		status:            WatcherStopped,
		newWatcherManager: newManagerFunc,
		store:             store,
	}, nil
}

// creates a controller
func newScheme(kinds []string) (*runtime.Scheme, error) {
	if len(kinds) == 0 {
		return &runtime.Scheme{}, fmt.Errorf("at least one kind is required")
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return &runtime.Scheme{}, err
	}

	for _, kind := range kinds {
		switch kind {
		case v2beta1.HelmReleaseKind:
			if err := v2beta1.AddToScheme(scheme); err != nil {
				return &runtime.Scheme{}, err
			}
		case v1beta2.KustomizationKind:
			if err := v1beta2.AddToScheme(scheme); err != nil {
				return &runtime.Scheme{}, err
			}
		default:
			return &runtime.Scheme{}, fmt.Errorf("kind not supported: %s", kind)
		}
	}
	return scheme, nil
}

func (w *DefaultWatcher) Start(ctx context.Context, log logr.Logger) error {
	if ctx == nil {
		return fmt.Errorf("invalid context")
	}
	if log.GetSink() == nil {
		return fmt.Errorf("invalid log sink")

	}
	w.log = log.WithName("watcher")

	ctx, cancel := context.WithCancel(ctx)
	w.stopFn = cancel

	cfg, err := w.cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("invalid cluster config")
	}

	w.watcherManager, err = w.newWatcherManager(cfg, w.kinds, w.store, ctrl.Options{
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

	w.status = WatcherStarted
	w.log.Info("watcher with helm reconciler started")
	return nil
}

// TODO add unit
func addReconcilerByKind(kind string, watcherManager manager.Manager, store store.Store, log logr.Logger) error {
	var rec reconciler.Reconciler
	var err error
	switch kind {
	case v2beta1.HelmReleaseKind:
		rec, err = reconciler.NewHelmWatcherReconciler(watcherManager.GetClient(), store, log)
	case v1beta2.KustomizationKind:
		rec, err = reconciler.NewKustomizeWatcherReconciler(watcherManager.GetClient(), store, log)
	default:
		return fmt.Errorf("not supported: %s", kind)
	}
	if err != nil {
		return fmt.Errorf("cannot create helm reconciler: %v", err)
	}
	//add helm reconciler
	log.Info(fmt.Sprintf("created reconciler %s", kind))
	err = rec.SetupWithManager(watcherManager)
	if err != nil {
		return fmt.Errorf("cannot setup helm reconciler: %v", err)
	}
	log.Info(fmt.Sprintf("setup with manager for reconciler %s", kind))
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
