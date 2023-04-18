package collector

import (
	"context"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Log:            log,
		ClusterManager: &cm,
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
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
			name:       "can start collector",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		Log:            log,
		ClusterManager: &cm,
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
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
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log: log,
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
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
			err = collector.Watch(ctx, tt.cluster)
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
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log: log,
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
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
	g.Expect(collector.Watch(ctx, c)).To(Succeed())
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
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	options := CollectorOpts{
		Log: log,
		ObjectKinds: []schema.GroupVersionKind{
			v2beta1.GroupVersion.WithKind(v2beta1.HelmReleaseKind),
			v1beta2.GroupVersion.WithKind(v1beta2.KustomizationKind),
		},
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
	err = collector.Watch(ctx, c)
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

func newFakeWatcher(config *rest.Config, clusterName string, objectsChannel chan []models.ObjectTransaction, kinds []schema.GroupVersionKind, log logr.Logger) (Watcher, error) {
	log.Info("created fake watcher")
	return &fakeWatcher{log: log}, nil
}

func fakeProcessRecordFunc(ctx context.Context, records []models.ObjectTransaction, s store.Store, logger logr.Logger) error {
	log.Info("fake process record")
	return nil
}

type fakeWatcher struct {
	log    logr.Logger
	status ClusterWatchingStatus
}

func (f *fakeWatcher) Start(ctx context.Context) error {
	f.status = ClusterWatchingStarted
	return nil
}

func (f *fakeWatcher) Stop(context.Context) error {
	f.status = ClusterWatchingStopped
	return nil
}

func (f *fakeWatcher) Status() (string, error) {
	return string(f.status), nil
}
