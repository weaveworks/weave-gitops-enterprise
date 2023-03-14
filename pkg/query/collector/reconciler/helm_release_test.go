package reconciler

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var log logr.Logger

func TestNewHelmWatcherReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log = testr.New(t)
	s := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	tests := []struct {
		name           string
		client         client.Client
		objectsChannel chan []models.Object
		errPattern     string
	}{
		{
			name:       "cannot create helm reconciler without client",
			errPattern: "invalid client",
		},
		{
			name:       "cannot create helm reconciler without store",
			client:     fakeClient,
			errPattern: "invalid objects channel",
		},
		{
			name:           "can create helm reconciler with valid arguments",
			client:         fakeClient,
			objectsChannel: make(chan []models.Object),
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewHelmWatcherReconciler(tt.client, tt.objectsChannel, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
			g.Expect(reconciler.client).NotTo(BeNil())
		})
	}
}
