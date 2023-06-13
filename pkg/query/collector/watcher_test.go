package collector

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/onsi/gomega"
)

func newFakeWatcherManagerFunc(opts WatcherManagerOptions) (manager.Manager, error) {
	return kubefakes.NewControllerManager(opts.Rest, opts.ManagerOptions)
}

func TestWatcher_Start(t *testing.T) {
	g := NewGomegaWithT(t)
	//setup watcher
	fakeObjectsChannel := make(chan []models.ObjectTransaction)
	options := WatcherOptions{
		ClientConfig: &rest.Config{
			Host: "http://idontexist",
		},
		ClusterRef: types.NamespacedName{
			Name:      "clusterName",
			Namespace: "clusterNamespace",
		},
		Kinds:         configuration.SupportedObjectKinds,
		ManagerFunc:   newFakeWatcherManagerFunc,
		ObjectChannel: fakeObjectsChannel,
		Log:           log,
	}

	watcher, err := NewWatcher(options)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name       string
		ctx        context.Context
		errPattern string
	}{
		{
			name:       "could start watcher with valid arguments",
			ctx:        context.Background(),
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.Start()
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			assertClusterWatcher(g, watcher, ClusterWatchingStarted)
		})

	}
}

func TestWatcher_Stop(t *testing.T) {
	g := NewGomegaWithT(t)
	//setup watcher
	fakeObjectsChannel := make(chan []models.ObjectTransaction)
	watcher := makeWatcherAndStart(g, fakeObjectsChannel, testr.New(t))
	assertClusterWatcher(g, watcher, ClusterWatchingStarted)

	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "could stop watcher with valid arguments",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer close(fakeObjectsChannel)
			var err error
			//TODO review deadlocks
			go func() {
				err = watcher.Stop()
			}()
			objectTransactions := <-fakeObjectsChannel
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			assertClusterWatcher(g, watcher, ClusterWatchingStopped)
			g.Expect(objectTransactions[0].TransactionType() == models.TransactionTypeDeleteAll).To(BeTrue())
		})
	}

}

func makeWatcherAndStart(g *WithT, objectsChannel chan []models.ObjectTransaction, log logr.Logger) Watcher {
	options := WatcherOptions{
		ClientConfig: &rest.Config{
			Host: "http://idontexist",
		},
		ClusterRef: types.NamespacedName{
			Name:      "clusterName",
			Namespace: "clusterNamespace",
		},
		Kinds:         configuration.SupportedObjectKinds,
		ManagerFunc:   newFakeWatcherManagerFunc,
		ObjectChannel: objectsChannel,
		Log:           log,
	}

	watcher, err := NewWatcher(options)
	g.Expect(err).To(BeNil())
	g.Expect(watcher.Start()).To(Succeed())
	return watcher
}

func assertClusterWatcher(g *WithT, watcher Watcher, expectedStatus ClusterWatchingStatus) {
	g.Expect(watcher).NotTo(BeNil())
	status, err := watcher.Status()
	g.Expect(err).To(BeNil())
	g.Expect(expectedStatus).To(BeIdenticalTo(ClusterWatchingStatus(status)))
}

func TestWatcher_defaultNewWatcherManager(t *testing.T) {
	g := NewGomegaWithT(t)
	fakeObjectsChannel := make(chan []models.ObjectTransaction)
	defer close(fakeObjectsChannel)

	tests := []struct {
		name       string
		opts       WatcherManagerOptions
		errPattern string
	}{
		{
			name: "cannot create default watcher manager with invalid params",
			opts: WatcherManagerOptions{
				Log: log,
				Rest: &rest.Config{
					Host: "http://idontexist",
				},
				Kinds:          configuration.SupportedObjectKinds,
				ObjectsChannel: fakeObjectsChannel,
				ClusterName:    "anyCluster",
				ManagerOptions: manager.Options{},
			},
			errPattern: "invalid service account name",
		},
		{
			name: "cannot create default watcher manager with valid params",
			opts: WatcherManagerOptions{
				Log: log,
				Rest: &rest.Config{
					Host: "http://idontexist",
				},
				Kinds:          configuration.SupportedObjectKinds,
				ObjectsChannel: fakeObjectsChannel,
				ClusterName:    "anyCluster",
				ManagerOptions: manager.Options{},
			},
			errPattern: "invalid service account name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := defaultNewWatcherManager(tt.opts)
			if err != nil {
				return
			}
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(manager).NotTo(BeNil())
		})
	}
}
