package reconciler

import (
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestNewReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	s := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	tests := []struct {
		name           string
		kinds          []string
		client         client.Client
		objectsChannel chan []models.ObjectRecord
		errPattern     string
	}{
		{
			name:       "cannot create reconciler without client",
			errPattern: "invalid client",
		},
		{
			name:       "cannot create reconciler without kinds",
			client:     fakeClient,
			errPattern: "invalid kinds",
		},
		{
			name:       "cannot create reconciler without object channel",
			client:     fakeClient,
			kinds:      []string{"HelmRelease"},
			errPattern: "invalid objects channel",
		},
		{
			name:           "could create reconciler with valid arguments",
			client:         fakeClient,
			kinds:          []string{"HelmRelease"},
			objectsChannel: make(chan []models.ObjectRecord),
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewReconciler(tt.kinds, tt.client, tt.objectsChannel, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
		})
	}
}

func TestSetup(t *testing.T) {
	g := NewGomegaWithT(t)
	s := runtime.NewScheme()
	v2beta1.AddToScheme(s)
	logger := testr.New(t)
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	fakeManager, err := kubefakes.NewControllerManager(&rest.Config{
		Host: "http://idontexist",
	}, ctrl.Options{
		Logger: logger,
		Scheme: s,
	})
	g.Expect(err).To(BeNil())
	g.Expect(fakeManager).NotTo(BeNil())
	objectsChannel := make(chan []models.ObjectRecord)
	kinds := []string{"HelmRelease"}
	reconciler, err := NewReconciler(kinds, fakeClient, objectsChannel, logger)
	g.Expect(err).To(BeNil())
	g.Expect(reconciler).NotTo(BeNil())

	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "can setup reconciler",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reconciler.Setup(fakeManager)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(reconciler).NotTo(BeNil())
		})
	}
}
