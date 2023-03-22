package reconciler

import (
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/kubefakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewReconciler(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	s := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	tests := []struct {
		name           string
		gvk            schema.GroupVersionKind
		client         client.Client
		objectsChannel chan []models.ObjectTransaction
		errPattern     string
	}{
		{
			name:       "cannot create reconciler without client",
			errPattern: "invalid client",
		},
		{
			name:       "cannot create reconciler without gvk",
			client:     fakeClient,
			errPattern: "invalid gvk",
		},
		{
			name:       "cannot create reconciler without object channel",
			client:     fakeClient,
			gvk:        v2beta1.GroupVersion.WithKind("HelmRelease"),
			errPattern: "invalid objects channel",
		},
		{
			name:           "could create reconciler with valid arguments",
			client:         fakeClient,
			gvk:            v2beta1.GroupVersion.WithKind("HelmRelease"),
			objectsChannel: make(chan []models.ObjectTransaction),
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := NewReconciler("test-cluster", tt.gvk, tt.client, tt.objectsChannel, log)
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
	if err := v2beta1.AddToScheme(s); err != nil {
		t.Fatalf("could not add v2beta1 to scheme: %v", err)
	}
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
	objectsChannel := make(chan []models.ObjectTransaction)
	gvk := v2beta1.GroupVersion.WithKind("HelmRelease")
	reconciler, err := NewReconciler("test-cluster", gvk, fakeClient, objectsChannel, logger)
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
