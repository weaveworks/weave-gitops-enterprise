package applicationscollector

import (
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"testing"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
)

var log logr.Logger
var g *WithT

func TestNewApplicationsCollector(t *testing.T) {
	g = NewWithT(t)
	log = testr.New(t)

	fakeStore := &storefakes.FakeStore{}

	tests := []struct {
		name       string
		store      store.Store
		options    collector.CollectorOpts
		errPattern string
	}{
		{
			name: "can create applications collector with valid arguments",
			options: collector.CollectorOpts{
				Log: log,
			},
			store:      fakeStore,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applicationsCollector, err := NewApplicationsCollector(tt.store, tt.options)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(applicationsCollector).NotTo(BeNil())
			g.Expect(applicationsCollector.col).NotTo(BeNil())
			g.Expect(applicationsCollector.store).NotTo(BeNil())

		})
	}
}
