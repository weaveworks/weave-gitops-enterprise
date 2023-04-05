package collector

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Kinds: []schema.GroupVersionKind{
			v2beta1.GroupVersion.WithKind("HelmRelease"),
		},
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
			err := watcher.Start(tt.ctx)
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
	ctx := context.Background()
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
				err = watcher.Stop(ctx)
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
		Kinds: []schema.GroupVersionKind{
			v2beta1.GroupVersion.WithKind("HelmRelease"),
		},
		ManagerFunc:   newFakeWatcherManagerFunc,
		ObjectChannel: objectsChannel,
		Log:           log,
	}

	watcher, err := NewWatcher(options)
	g.Expect(err).To(BeNil())
	g.Expect(watcher.Start(context.Background())).To(Succeed())
	return watcher
}

func Test_newScheme(t *testing.T) {

	g := NewGomegaWithT(t)

	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "can create default scheme",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := newDefaultScheme()
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(scheme).NotTo(BeNil())
		})
	}

}

func assertClusterWatcher(g *WithT, watcher Watcher, expectedStatus ClusterWatchingStatus) {
	g.Expect(watcher).NotTo(BeNil())
	status, err := watcher.Status()
	g.Expect(err).To(BeNil())
	g.Expect(ClusterWatchingStatus(status) == expectedStatus).To(BeTrue())
}
