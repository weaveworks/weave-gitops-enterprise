package collector

import (
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var log logr.Logger
var g *WithT

func TestNewCollector(t *testing.T) {
	g = NewWithT(t)
	log = testr.New(t)

	kinds := []schema.GroupVersionKind{
		v2beta1.GroupVersion.WithKind(v2beta1.HelmReleaseKind),
		v1beta2.GroupVersion.WithKind(v1beta2.KustomizationKind),
	}

	fakeStore := &storefakes.FakeStore{}

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	tests := []struct {
		name       string
		options    CollectorOpts
		store      store.Store
		errPattern string
	}{
		{
			name: "can create collector with valid arguments",
			options: CollectorOpts{
				Log:                log,
				ObjectKinds:        kinds,
				ProcessRecordsFunc: fakeProcessRecordFunc,
				NewWatcherFunc:     newFakeWatcher,
				ClusterManager:     clustersManager,
			},
			store:      fakeStore,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			collector, err := NewCollector(tt.options, tt.store)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(collector).NotTo(BeNil())
		})
	}

}
