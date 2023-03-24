package main

import (
	"fmt"
	"github.com/fluxcd/pkg/runtime/logger"
	flag "github.com/spf13/pflag"
	collectorapp "github.com/weaveworks/weave-gitops-enterprise/cmd/collector/app"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const (
	controllerName = "collector"
)

func main() {
	var (
		probeAddr            string
		metricsAddr          string
		enableLeaderElection bool
		querySeverUrl        string
		clusterNamespace     string
		logOptions           logger.Options
	)

	scheme := runtime.NewScheme()
	log := ctrl.Log.WithName("setup")
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&querySeverUrl, "query-server-url", os.Getenv("QUERY_SERVER_URL"), "Url for the query server to send collected objects.")
	flag.StringVar(&clusterNamespace, "cluster-namespace", os.Getenv("CLUSTER_NAMESPACE"), "namespace to watch gitops cluster to")

	logOptions.BindFlags(flag.CommandLine)

	flag.Parse()

	log = logger.NewLogger(logOptions)
	ctrl.SetLogger(log)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       fmt.Sprintf("%s-leader-election", controllerName),
	})

	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	token, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Error(err, "cannot read service account token")
		os.Exit(1)
	}

	remoteStore, err := store.NewStore(store.RemoteBackend, store.StoreOpts{
		Log:   log,
		Url:   querySeverUrl,
		Token: string(token),
	})
	if err != nil {
		log.Error(err, "unable to create remote store")
		os.Exit(1)
	}

	//TODO these arguments should come from
	opts := collectorapp.ServerOpts{
		Logger:            log,
		ClustersNamespace: "flux-system",
		Store:             remoteStore,
	}

	//TODO capture stop
	collector, _, err := collectorapp.NewServer(ctx, opts)
	if err != nil {
		log.Error(err, "unable to create collector server")
		os.Exit(1)
	}

	go func() {
		if err := collector.Start(ctx); err != nil {
			log.Error(err, "cannot start collector server")
			os.Exit(1)
		}
	}()

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
