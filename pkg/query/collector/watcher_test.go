package collector

import (
	"context"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewWatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	fakeObjectsChannel := make(chan []models.Object)

	tests := []struct {
		name                      string
		options                   WatcherOptions
		managerFunc               newWatcherManagerFunc
		objectsChannel            chan []models.Object
		expectedRegisteredVersion schema.GroupVersion
		errPattern                string
	}{
		{
			name:       "cannot create watcher for empty options",
			options:    WatcherOptions{},
			errPattern: "invalid config",
		},
		{
			name: "cannot create watcher for empty config",
			options: WatcherOptions{
				ClientConfig: nil,
			},
			errPattern: "invalid config",
		},
		{
			name: "cannot create watcher for empty cluster",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{},
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "cannot create watcher for empty cluster",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{
					Name: "clusterName",
				},
			},
			errPattern: "cluster name or namespace is empty",
		},
		{
			name: "cannot create watcher for empty kinds",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{
					Name:      "clusterName",
					Namespace: "clusterNamespace",
				},
			},
			errPattern: "at least one kind is required",
		},
		{
			name: "cannot create watcher for empty objectsChannel",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{
					Name:      "clusterName",
					Namespace: "clusterNamespace",
				},
				Kinds: []string{
					v2beta1.HelmReleaseKind,
				},
			},
			errPattern: "invalid objects channel",
		},
		{
			name: "can create watcher with default func",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{
					Name:      "clusterName",
					Namespace: "clusterNamespace",
				},
				Kinds: []string{
					v2beta1.HelmReleaseKind,
				},
			},
			expectedRegisteredVersion: v2beta1.GroupVersion,
			objectsChannel:            fakeObjectsChannel,
			errPattern:                "",
		},
		{
			name: "can create watcher with custom manager func",
			options: WatcherOptions{
				ClientConfig: &rest.Config{
					Host: "http://idontexist",
				},
				ClusterRef: types.NamespacedName{
					Name:      "clusterName",
					Namespace: "clusterNamespace",
				},
				Kinds: []string{
					v2beta1.HelmReleaseKind,
				},
			},
			objectsChannel: fakeObjectsChannel,
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewWatcher(tt.options, tt.managerFunc, tt.objectsChannel, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(watcher).NotTo(BeNil())
			g.Expect(watcher.clusterRef).To(Equal(tt.options.ClusterRef))
			g.Expect(watcher.status).To(Equal(ClusterWatchingStopped))
			g.Expect(watcher.scheme).NotTo(BeNil())
			if tt.expectedRegisteredVersion.Version != "" {
				g.Expect(watcher.scheme.IsVersionRegistered(tt.expectedRegisteredVersion)).To(BeTrue())
			}
			g.Expect(watcher.cluster).NotTo(BeNil())
			g.Expect(watcher.newWatcherManager).NotTo(BeNil())
		})
	}

}

func newFakeWatcherManagerFunc(config *rest.Config, kinds []string, objectsChannel chan []models.Object, options manager.Options) (manager.Manager, error) {
	options.Logger.Info("created fake watcher manager")
	return kubefakes.NewControllerManager(config, options)
}

func TestStartWatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	//setup watcher
	options := WatcherOptions{
		ClientConfig: &rest.Config{
			Host: "http://idontexist",
		},
		ClusterRef: types.NamespacedName{
			Name:      "clusterName",
			Namespace: "clusterNamespace",
		},
		Kinds: []string{
			v2beta1.HelmReleaseKind,
		},
	}

	log := testr.New(t)
	fakeObjectsChannel := make(chan []models.Object)
	//setup a valid watcher
	watcher, err := NewWatcher(options, newFakeWatcherManagerFunc, fakeObjectsChannel, log)
	g.Expect(err).To(BeNil())
	g.Expect(watcher).NotTo(BeNil())
	g.Expect(watcher.objectsChannel).NotTo(BeNil())

	tests := []struct {
		name       string
		ctx        context.Context
		log        logr.Logger
		errPattern string
	}{
		{
			name:       "cannot start watcher without context",
			log:        log,
			errPattern: "invalid context",
		},
		{
			name:       "cannot start watcher without log",
			ctx:        context.Background(),
			log:        logr.Logger{},
			errPattern: "invalid log sink",
		},
		{
			name:       "could start watcher with valid arguments",
			ctx:        context.Background(),
			log:        log,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.Start(tt.ctx, tt.log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(watcher.status).To(Equal(ClusterWatchingStarted))
			g.Expect(watcher.watcherManager).NotTo(BeNil())
		})
	}

}

func Test_newScheme(t *testing.T) {

	g := NewGomegaWithT(t)

	tests := []struct {
		name       string
		kinds      []string
		errPattern string
	}{
		{
			name:       "cannot create scheme without kinds",
			errPattern: "at least one kind is required",
		},
		{
			name:       "cannot create scheme with unsupported kind",
			kinds:      []string{"InvalidKind"},
			errPattern: "kind not supported: InvalidKind",
		},
		{
			name:       "could create scheme for supported kinds",
			kinds:      []string{v2beta1.HelmReleaseKind, v1beta2.KustomizationKind},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := newScheme(tt.kinds)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(scheme).NotTo(BeNil())
		})
	}

}
