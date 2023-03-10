package reconciler

import (
	"github.com/enekofb/collector/pkg/cluster/fakes"
	"github.com/enekofb/collector/pkg/cluster/store"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var log logr.Logger

func TestNewHelmWatcherReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log = testr.New(t)
	fakeClient := fakes.NewClient(log)
	fakeStore := fakes.NewStore(log)
	tests := []struct {
		name       string
		client     client.Client
		store      store.Store
		errPattern string
	}{
		{
			name:       "cannot create helm reconciler without client",
			errPattern: "invalid client",
		},
		{
			name:       "cannot create helm reconciler without store",
			client:     fakeClient,
			errPattern: "invalid store",
		},
		{
			name:       "can create helm reconciler with valid arguments",
			client:     fakeClient,
			store:      fakeStore,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewHelmWatcherReconciler(tt.client, tt.store, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
			g.Expect(reconciler.store).NotTo(BeNil())
			g.Expect(reconciler.client).NotTo(BeNil())
		})
	}
}
