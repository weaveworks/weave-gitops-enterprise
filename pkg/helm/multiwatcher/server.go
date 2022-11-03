package multiwatcher

import (
	"context"
	"errors"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher/controller"
)

type Options struct {
	ClusterRef    types.NamespacedName
	ClientConfig  *rest.Config
	Cache         helm.ChartsCacherWriter
	ValuesFetcher helm.ValuesFetcher
	UseProxy      bool
}

type Watcher struct {
	scheme        *runtime.Scheme
	clusterRef    types.NamespacedName
	clientConfig  *rest.Config
	cache         helm.ChartsCacherWriter
	valuesFetcher helm.ValuesFetcher
	useProxy      bool
	stopFn        context.CancelFunc
	log           logr.Logger
}

func NewWatcher(opts Options) (*Watcher, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := sourcev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return &Watcher{
		clusterRef:    opts.ClusterRef,
		clientConfig:  opts.ClientConfig,
		cache:         opts.Cache,
		valuesFetcher: opts.ValuesFetcher,
		useProxy:      opts.UseProxy,
	}, nil
}

func (w *Watcher) StartWatcher(ctx context.Context, log logr.Logger) error {
	w.log = log.WithName("multi-helm-watcher")

	ctx, cancel := context.WithCancel(ctx)
	w.stopFn = cancel

	mgr, err := ctrl.NewManager(w.clientConfig, ctrl.Options{
		Scheme:             w.scheme,
		Logger:             w.log,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		w.log.Error(err, "unable to create manager")
		return err
	}

	if err = (&controller.HelmWatcherReconciler{
		ClusterRef:    w.clusterRef,
		ClientConfig:  w.clientConfig,
		Cache:         w.cache,
		ValuesFetcher: w.valuesFetcher,
		UseProxy:      w.useProxy,
		Client:        mgr.GetClient(),
		Scheme:        w.scheme,
	}).SetupWithManager(mgr); err != nil {
		w.log.Error(err, "unable to create controller", "controller", "HelmWatcherReconciler")
		return err
	}

	w.log.Info("starting manager")

	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "problem running manager")
		return err
	}

	return nil
}

func (w *Watcher) Stop() {
	if w.stopFn == nil {
		w.log.Error(errors.New("Stop function not set yet"), "unable to stop watcher")
		return
	}
	w.stopFn()
}
