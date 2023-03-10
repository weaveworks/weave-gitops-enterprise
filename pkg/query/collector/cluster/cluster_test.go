package cluster

import (
	"context"
	"github.com/enekofb/collector/pkg/cluster/fakes"
	"github.com/enekofb/collector/pkg/cluster/store"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewMultiClusterWatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := fakes.NewStore(log)
	clustersWatcher, err := NewClustersWatcher(fakeStore, newFakeWatcher, log)
	g.Expect(err).To(BeNil())
	g.Expect(clustersWatcher).NotTo(BeNil())
	g.Expect(clustersWatcher.store).NotTo(BeNil())
	g.Expect(clustersWatcher.kinds).To(ContainElements(v2beta1.HelmReleaseKind,
		v1beta2.KustomizationKind))
}

func TestAddCluster(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeStore := fakes.NewStore(log)
	clustersWatcher, err := NewClustersWatcher(fakeStore, newFakeWatcher, log)
	g.Expect(err).To(BeNil())
	g.Expect(clustersWatcher).NotTo(BeNil())
	g.Expect(len(clustersWatcher.clusterWatchers)).To(Equal(0))
	ctx := context.Background()

	tests := []struct {
		name       string
		cluster    types.NamespacedName
		config     *rest.Config
		errPattern string
	}{
		{
			name:       "cannot add cluster without config",
			config:     nil,
			errPattern: "config not found",
		},
		{
			name: "cannot add cluster without name",
			config: &rest.Config{
				Host: "http://idontexist",
			},
			cluster: types.NamespacedName{
				Namespace: "test",
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "cannot add cluster without namespace",
			config: &rest.Config{
				Host: "http://idontexist",
			},
			cluster: types.NamespacedName{
				Name: "test",
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "could add cluster with cluster and config",
			cluster: types.NamespacedName{
				Name:      "test",
				Namespace: "test",
			},
			config: &rest.Config{
				Host: "http://idontexist",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clustersWatcher.AddCluster(tt.cluster, tt.config, ctx, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(clustersWatcher.clusterWatchers[tt.cluster.String()]).NotTo(BeNil())

		})
	}

}

func TestStatusCluster(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	ctx := context.Background()
	fakeStore := fakes.NewStore(log)
	clustersWatcher, err := NewClustersWatcher(fakeStore, newFakeWatcher, log)
	g.Expect(err).To(BeNil())
	g.Expect(clustersWatcher).NotTo(BeNil())
	g.Expect(len(clustersWatcher.clusterWatchers)).To(Equal(0))
	cluster := types.NamespacedName{
		Name:      "test",
		Namespace: "test",
	}
	config := &rest.Config{
		Host: "http://idontexist",
	}
	err = clustersWatcher.AddCluster(cluster, config, ctx, log)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		cluster        types.NamespacedName
		errPattern     string
		expectedStatus string
	}{
		{
			name: "cannot get status for cluster without name",
			cluster: types.NamespacedName{
				Namespace: "test",
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "cannot get status for cluster without namespace",
			cluster: types.NamespacedName{
				Name: "test",
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "could get status for existing cluster",
			cluster: types.NamespacedName{
				Name:      "test",
				Namespace: "test",
			},
			expectedStatus: string(WatcherStopped),
		},
		{
			name: "could not get status for non existing cluster",
			cluster: types.NamespacedName{
				Name:      "dontexist",
				Namespace: "dontexist",
			},
			errPattern: "cluster not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := clustersWatcher.StatusCluster(tt.cluster)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(status).To(Equal(tt.expectedStatus))
		})
	}
}

func newFakeWatcher(config *rest.Config, cluster types.NamespacedName, store store.Store, kinds []string, log logr.Logger) (Watcher, error) {
	log.Info("created fake watcher")
	return &fakeWatcher{log: log}, nil
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
	return string(WatcherStopped), nil
}
