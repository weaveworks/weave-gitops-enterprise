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
	"k8s.io/client-go/rest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestStart(t *testing.T) {
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
	g.Expect(collector).NotTo(BeNil())

	tests := []struct {
		name       string
		clusters   []cluster.Cluster
		errPattern string
	}{
		{
			name:       "can start collector for empty collection",
			clusters:   []cluster.Cluster{},
			errPattern: "",
		},
		{
			name:       "can start collector with not watchable clusters",
			clusters:   []cluster.Cluster{newNonWatchableCluster()},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.GetClustersReturns(tt.clusters)
			err := collector.Start()
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(fakeStore).NotTo(BeNil())
		})
	}
}

func newNonWatchableCluster() cluster.Cluster {
	cluster := new(clusterfakes.FakeCluster)
	cluster.GetNameReturns("non-watchable")

	return cluster
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

	config := &rest.Config{
		Host: "http://idontexist",
	}

	c := makeCluster("testcluster", config, log)

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

	config := &rest.Config{
		Host: "http://idontexist",
	}
	clusterName := "testCluster"
	c := makeCluster(clusterName, config, log)
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

func makeCluster(name string, config *rest.Config, log logr.Logger) cluster.Cluster {
	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	cluster.GetServerConfigReturns(config, nil)
	log.Info("fake watcher created", "watcher", cluster.GetName())
	return &cluster
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
	c := makeCluster(existingClusterName, &rest.Config{
		Host: "http://idontexist",
	}, log)
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
