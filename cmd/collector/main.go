package main

import (
	"fmt"
	"github.com/fluxcd/pkg/runtime/logger"
	flag "github.com/spf13/pflag"
	collector_app "github.com/weaveworks/weave-gitops-enterprise/cmd/collector/app"
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

	//TODO move me to configuration
	remoteStore, err := store.NewRemoteStore(store.RemoteStoreOpts{
		Log:     log,
		Address: "http://localhost:8000",
		Token:   "Tilt-Token=055ffbb8-bb22-4b2a-9531-d9c19f8c98a3; ajs_user_id=89391d989aae718fdf2870a8ac17a6be; ajs_anonymous_id=be83a59d-01a3-4229-9403-189ff3342a66; ph_phc_DjQgd6iqP8qrhQN6fjkuGeTIk004coiDRmIdbZLRooo_posthog={\"distinct_id\":\"186ac0bc5ba12b0-0dfb19159869d9-17525635-16a7f0-186ac0bc5bb281d\",\"$device_id\":\"186ac0bc5ba12b0-0dfb19159869d9-17525635-16a7f0-186ac0bc5bb281d\",\"$user_state\":\"anonymous\",\"$referrer\":\"$direct\",\"$referring_domain\":\"$direct\",\"$sesid\":[1679171555138,\"186f6621ac8118d-0b99d917e5c57d-1f525634-16a7f0-186f6621ac91289\",1679170869959],\"$session_recording_enabled_server_side\":true,\"$console_log_recording_enabled_server_side\":true,\"$active_feature_flags\":[],\"$enabled_feature_flags\":{},\"$feature_flag_payloads\":{},\"$session_recording_recorder_version_server_side\":\"v1\"}; id_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ3ZWdvLWFkbWluIiwiZXhwIjoxNjc5NDc4MjA4LCJuYmYiOjE2Nzk0NzQ2MDgsImlhdCI6MTY3OTQ3NDYwOH0.6yUJW2Ge_hTsLZ4rfYBX01CpR9a7cCwsbNQSHe-q2Kg",
	})
	if err != nil {
		log.Error(err, "unable to create remote store")
		os.Exit(1)
	}

	//TODO these arguments should come from
	opts := collector_app.ServerOpts{
		Logger:            log,
		ClustersNamespace: "flux-system",
		RemoteStore:       remoteStore,
	}

	//TODO capture stop
	collector, _, err := collector_app.NewServer(ctx, opts)
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
