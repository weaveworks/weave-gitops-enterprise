package collector

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"
	"os"
	"testing"

	l "github.com/weaveworks/weave-gitops/core/logger"

	. "github.com/onsi/gomega"
)

func TestStart(t *testing.T) {
	g := NewGomegaWithT(t)
	log, loggerPath := newLoggerWithLevel(t, "DEBUG")

	fakeStore := &storefakes.FakeStore{}
	cm := clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	cm.SubscribeReturns(&cmw)
	opts := CollectorOpts{
		Log:                log,
		ClusterManager:     &cm,
		ObjectKinds:        configuration.SupportedObjectKinds,
		ProcessRecordsFunc: fakeProcessRecordFunc,
		NewWatcherFunc:     newFakeWatcher,
	}
	collector, err := newWatchingCollector(opts, fakeStore)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())

	tests := []struct {
		name             string
		clusters         []cluster.Cluster
		expectedLogError string
	}{
		{
			name:             "can start collector for empty collection",
			clusters:         []cluster.Cluster{},
			expectedLogError: "",
		},
		{
			name:             "can start collector with not watchable clusters",
			clusters:         []cluster.Cluster{makeInvalidFakeCluster("test-cluster")},
			expectedLogError: "cannot watch cluster",
		},
		{
			name:             "can start collector with watchable clusters",
			clusters:         []cluster.Cluster{makeValidFakeCluster("test-cluster")},
			expectedLogError: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.GetClustersReturns(tt.clusters)
			err := collector.Start()

			// assert error has been logged
			if tt.expectedLogError != "" {
				logs, err := os.ReadFile(loggerPath)
				if err != nil {
					t.Fatalf("cannot get logs: %v", err)
				}
				logss := string(logs)
				g.Expect(logss).To(MatchRegexp(tt.expectedLogError))
			}

			g.Expect(err).To(BeNil())
			g.Expect(fakeStore).NotTo(BeNil())

		})
	}
}

func makeInvalidFakeCluster(name string) cluster.Cluster {
	cluster := new(clusterfakes.FakeCluster)
	cluster.GetNameReturns(name)
	return cluster
}

func makeValidFakeCluster(name string) cluster.Cluster {
	config := &rest.Config{
		Host: "http://idontexist",
	}

	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	cluster.GetServerConfigReturns(config, nil)
	return &cluster
}

func TestStop(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}

	cm := clustersmngrfakes.FakeClustersManager{}

	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	cm.SubscribeReturns(&cmw)
	opts := CollectorOpts{
		Log:                log,
		ClusterManager:     &cm,
		ObjectKinds:        configuration.SupportedObjectKinds,
		ProcessRecordsFunc: fakeProcessRecordFunc,
		NewWatcherFunc:     newFakeWatcher,
	}
	collector, err := newWatchingCollector(opts, fakeStore)
	g.Expect(err).To(BeNil())
	err = collector.Start()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "can stop collector",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := collector.Stop()
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestClusterWatcher_Watch(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log:                log,
		ObjectKinds:        configuration.SupportedObjectKinds,
		ProcessRecordsFunc: fakeProcessRecordFunc,
		NewWatcherFunc:     newFakeWatcher,
	}
	collector, err := newWatchingCollector(opts, fakeStore)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())

	c := makeValidFakeCluster("testcluster")

	tests := []struct {
		name       string
		cluster    cluster.Cluster
		errPattern string
	}{
		{
			name:       "can watch cluster",
			cluster:    c,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = collector.Watch(tt.cluster)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			status, err := collector.Status(tt.cluster.GetName())
			g.Expect(err).To(BeNil())
			g.Expect(ClusterWatchingStarted).To(BeIdenticalTo(ClusterWatchingStatus(status)))
		})
	}
}

func TestClusterWatcher_Unwatch(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log:                log,
		ObjectKinds:        configuration.SupportedObjectKinds,
		ProcessRecordsFunc: fakeProcessRecordFunc,
		NewWatcherFunc:     newFakeWatcher,
	}
	collector, err := newWatchingCollector(opts, fakeStore)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())

	clusterName := "testCluster"
	c := makeValidFakeCluster(clusterName)
	g.Expect(collector.Watch(c)).To(Succeed())
	watcher := collector.clusterWatchers[clusterName]
	tests := []struct {
		name        string
		watcher     Watcher
		clusterName string
		errPattern  string
	}{
		{
			name:        "unwatch empty cluster throws error",
			watcher:     nil,
			clusterName: "",
			errPattern:  "cluster name is empty",
		},
		{
			name:        "unwatch non-existing cluster throws error",
			watcher:     nil,
			clusterName: "idontexist",
			errPattern:  "cluster watcher not found",
		},
		{
			name:        "unwatch existing cluster unwatches it",
			watcher:     watcher,
			clusterName: clusterName,
			errPattern:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.watcher != nil {
				s, err := tt.watcher.Status()
				g.Expect(err).To(BeNil())
				g.Expect(ClusterWatchingStarted).To(BeIdenticalTo(ClusterWatchingStatus(s)))
			}
			err = collector.Unwatch(tt.clusterName)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(collector.clusterWatchers[tt.clusterName]).To(BeNil())
			s, err := tt.watcher.Status()
			g.Expect(err).To(BeNil())
			g.Expect(ClusterWatchingStopped).To(BeIdenticalTo(ClusterWatchingStatus(s)))
		})
	}
}

func TestClusterWatcher_Status(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := &storefakes.FakeStore{}
	options := CollectorOpts{
		Log:                log,
		ObjectKinds:        configuration.SupportedObjectKinds,
		ProcessRecordsFunc: fakeProcessRecordFunc,
		NewWatcherFunc:     newFakeWatcher,
	}
	collector, err := newWatchingCollector(options, fakeStore)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())
	g.Expect(len(collector.clusterWatchers)).To(Equal(0))
	existingClusterName := "test"
	c := makeValidFakeCluster(existingClusterName)
	err = collector.Watch(c)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		clusterName    string
		errPattern     string
		expectedStatus string
	}{
		{
			name:        "cannot get status for non existing cluster",
			clusterName: "",
			errPattern:  "cluster name is empty",
		},
		{
			name:        "could not get status for non existing cluster",
			clusterName: "dontexist",
			errPattern:  "cluster not found",
		},
		{
			name:           "could get status for existing cluster",
			clusterName:    existingClusterName,
			expectedStatus: string(ClusterWatchingStarted),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := collector.Status(tt.clusterName)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(status).To(Equal(tt.expectedStatus))
		})
	}
}

func newFakeWatcher(config *rest.Config, serviceAccount ImpersonateServiceAccount, clusterName string, objectsChannel chan []models.ObjectTransaction, kinds []configuration.ObjectKind, log logr.Logger) (Watcher, error) {
	log.Info("created fake watcher")
	return &fakeWatcher{log: log}, nil
}

func fakeProcessRecordFunc(records []models.ObjectTransaction, s store.Store, logger logr.Logger) error {
	log.Info("fake process record")
	return nil
}

type fakeWatcher struct {
	log    logr.Logger
	status ClusterWatchingStatus
}

func (f *fakeWatcher) Start() error {
	f.status = ClusterWatchingStarted
	return nil
}

func (f *fakeWatcher) Stop() error {
	f.status = ClusterWatchingStopped
	return nil
}

func (f *fakeWatcher) Status() (string, error) {
	return string(f.status), nil
}

func newLoggerWithLevel(t *testing.T, logLevel string) (logr.Logger, string) {
	g := NewGomegaWithT(t)

	file, err := os.CreateTemp(os.TempDir(), "query-server-log")
	g.Expect(err).ShouldNot(HaveOccurred())

	name := file.Name()
	g.Expect(err).ShouldNot(HaveOccurred())

	level, err := zapcore.ParseLevel(logLevel)
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
		l.WithOutAndErrPaths(name, name),
	)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	return log, name
}
