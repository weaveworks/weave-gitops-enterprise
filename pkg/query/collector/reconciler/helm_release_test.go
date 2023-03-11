package reconciler

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var log logr.Logger

func TestNewHelmWatcherReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log = testr.New(t)
	fakeClient := kubefakes.NewClient(log)
	fakeStore := storefakes.NewStore(log)
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
