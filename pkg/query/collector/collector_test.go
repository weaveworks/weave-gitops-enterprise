package collector

import (
	"runtime"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
)

var log logr.Logger
var g *WithT

func TestNewCollector(t *testing.T) {
	g = NewWithT(t)
	log = testr.New(t)

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	tests := []struct {
		name       string
		options    CollectorOpts
		errPattern string
	}{
		{
			name: "can create collector with valid arguments",
			options: CollectorOpts{
				Log:            log,
				NewWatcherFunc: newFakeWatcher,
				Clusters:       clustersManager,
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			collector, err := NewCollector(tt.options)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(collector).NotTo(BeNil())
		})
	}

}

func TestCleanShutdown(t *testing.T) {
	g = NewWithT(t)
	log = testr.New(t)

	routineCountBefore := runtime.NumGoroutine()

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	opts := CollectorOpts{
		Log:            log,
		NewWatcherFunc: newFakeWatcher,
		Clusters:       clustersManager,
	}
	col, err := NewCollector(opts)
	g.Expect(err).NotTo(HaveOccurred())

	col.Start()
	col.Stop()
	g.Expect(runtime.NumGoroutine()).To(Equal(routineCountBefore), "number of goroutines before starting = number of goroutines after stopping (no leaked goroutines)")
}
