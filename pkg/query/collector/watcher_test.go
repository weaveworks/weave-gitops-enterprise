package collector

import (
	"context"
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/onsi/gomega"
)

func newFakeWatcherManagerFunc(config *rest.Config, kinds []schema.GroupVersionKind, objectsChannel chan []models.ObjectRecord, options manager.Options) (manager.Manager, error) {
	return kubefakes.NewControllerManager(config, options)
}

func TestStartWatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	//setup watcher
	fakeObjectsChannel := make(chan []models.ObjectRecord)
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
			err := watcher.Start(tt.ctx, logr.Discard())
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}

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
