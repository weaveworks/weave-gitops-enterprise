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
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestStart(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log:      log,
		Clusters: []cluster.Cluster{},
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
	}
	collector, err := newWatchingCollector(opts, fakeStore, newFakeWatcher, fakeProcessRecordFunc)
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
			err := collector.Start(ctx)
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
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log: log,
		Clusters: []cluster.Cluster{
			&clusterfakes.FakeCluster{
				GetHostStub: nil,
				GetNameStub: func() string {
					return "faked"
				},
				GetServerConfigStub: func() (*rest.Config, error) {
					return &rest.Config{}, nil
				},
			},
		},
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
	}
	collector, err := newWatchingCollector(opts, fakeStore, newFakeWatcher, fakeProcessRecordFunc)
	g.Expect(err).To(BeNil())
	err = collector.Start(ctx)
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
			err := collector.Stop(ctx)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestWatch(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	ctx := context.Background()
	fakeStore := &storefakes.FakeStore{}
	opts := CollectorOpts{
		Log: log,
		ObjectKinds: []schema.GroupVersionKind{
			rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		},
	}
	collector, err := newWatchingCollector(opts, fakeStore, newFakeWatcher, fakeProcessRecordFunc)
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
			g.Expect(err).To(BeNil())
			err = collector.Watch(tt.cluster, collector.objectsChannel, ctx, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(collector.clusterWatchers[tt.cluster.GetName()]).NotTo(BeNil())
		})
	}
}

func makeCluster(name string, config *rest.Config, log logr.Logger) cluster.Cluster {
	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	cluster.GetServerConfigReturns(config, nil)
	log.Info("fake cluster created", "cluster", cluster.GetName())
	return &cluster
}

func TestStatusCluster(t *testing.T) {
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
	}
	collector, err := newWatchingCollector(options, fakeStore, newFakeWatcher, fakeProcessRecordFunc)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())
	g.Expect(len(collector.clusterWatchers)).To(Equal(0))
	clusterName := types.NamespacedName{
		Name: "test",
	}
	config := &rest.Config{
		Host: "http://idontexist",
	}

	c := makeCluster(clusterName.Name, config, log)
	err = collector.Watch(c, collector.objectsChannel, ctx, log)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		clusterName    types.NamespacedName
		errPattern     string
		expectedStatus string
	}{
		{
			name:        "cannot get status for cluster without name",
			clusterName: types.NamespacedName{},
			errPattern:  "cluster name is empty",
		},
		{
			name: "could get status for existing cluster",
			clusterName: types.NamespacedName{
				Name: "test",
			},
			expectedStatus: string(ClusterWatchingStopped),
		},
		{
			name: "could not get status for non existing clusterName",
			clusterName: types.NamespacedName{
				Name:      "dontexist",
				Namespace: "dontexist",
			},
			errPattern: "clusterName not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := makeCluster(tt.clusterName.Name, nil, log)
			status, err := collector.Status(c)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(status).To(Equal(tt.expectedStatus))
		})
	}
}

func newFakeWatcher(config *rest.Config, clusterName string, objectsChannel chan []models.ObjectRecord, kinds []schema.GroupVersionKind, log logr.Logger) (Watcher, error) {
	log.Info("created fake watcher")
	return &fakeWatcher{log: log}, nil
}

func fakeProcessRecordFunc(ctx context.Context, records []models.ObjectRecord, s store.Store, logger logr.Logger) error {
	log.Info("fake process record")
	return nil
}

type fakeWatcher struct {
	log logr.Logger
}

func (f fakeWatcher) Start(ctx context.Context, log logr.Logger) error {
	f.log.Info("fake watcher started")
	return nil
}

func (f fakeWatcher) Stop() error {
	f.log.Info("fake watcher stopped")
	return nil
}

func (f fakeWatcher) Status() (string, error) {
	f.log.Info("fake watcher status")
	return string(ClusterWatchingStopped), nil
}
