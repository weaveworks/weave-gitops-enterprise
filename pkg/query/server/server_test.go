package server

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func TestNewServer(t *testing.T) {
	g := NewWithT(t)

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)

	tests := []struct {
		name       string
		options    ServerOpts
		errPattern string
	}{
		{
			name: "cannot create server with invalid arguments",
			options: ServerOpts{
				Logger:          logr.Discard(),
				ObjectKinds:     configuration.SupportedObjectKinds,
				DiscoveryClient: fakeDiscovery,
				ClustersManager: clustersManager,
			},
			errPattern: "service account name cannot be empty",
		},
		{
			name: "can create server with valid arguments",
			options: ServerOpts{
				Logger:      logr.Discard(),
				ObjectKinds: configuration.SupportedObjectKinds,
				ServiceAccount: collector.ImpersonateServiceAccount{
					Name:      "collector",
					Namespace: "flux-system",
				},
				DiscoveryClient: fakeDiscovery,
				ClustersManager: clustersManager,
				SkipCollection:  false,
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, err := NewServer(tt.options)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(s).NotTo(BeNil())

			s2 := s.(*server)

			if !tt.options.SkipCollection {
				g.Expect(s2.arc).NotTo(BeNil())
				g.Expect(s2.objs).NotTo(BeNil())
			}
		})
	}

}

func TestListEnabledComponents(t *testing.T) {
	g := NewWithT(t)

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)

	opts := ServerOpts{
		Logger:      logr.Discard(),
		ObjectKinds: configuration.SupportedObjectKinds,
		ServiceAccount: collector.ImpersonateServiceAccount{
			Name:      "collector",
			Namespace: "flux-system",
		},
		DiscoveryClient: fakeDiscovery,
		ClustersManager: clustersManager,
		SkipCollection:  true,
		EnabledFor:      []string{"applications"},
	}

	srv, stop, err := NewServer(opts)
	defer stop()
	g.Expect(err).To(BeNil())

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterQueryServer(s, srv)
	})

	qs := pb.NewQueryClient(conn)

	res, err := qs.ListEnabledComponents(context.Background(), &pb.ListEnabledComponentsRequest{})
	g.Expect(err).To(BeNil())

	g.Expect(res.Components).To(ContainElement(pb.EnabledComponent_applications))

	g.Expect(res.Components).NotTo(ContainElement(pb.EnabledComponent_sources))
}

func TestListEnabledComponents_Invalid(t *testing.T) {
	g := NewWithT(t)

	clustersManager := &clustersmngrfakes.FakeClustersManager{}
	cmw := clustersmngr.ClustersWatcher{
		Updates: make(chan clustersmngr.ClusterListUpdate),
	}
	clustersManager.SubscribeReturns(&cmw)

	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)

	opts := ServerOpts{
		Logger:      logr.Discard(),
		ObjectKinds: configuration.SupportedObjectKinds,
		ServiceAccount: collector.ImpersonateServiceAccount{
			Name:      "collector",
			Namespace: "flux-system",
		},
		DiscoveryClient: fakeDiscovery,
		ClustersManager: clustersManager,
		SkipCollection:  true,
		// Handling invalid component name to avoid
		// having the first value of the enum get returned on an unknown string.
		EnabledFor: []string{"foobar"},
	}

	srv, stop, err := NewServer(opts)
	defer stop()
	g.Expect(err).To(BeNil())

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterQueryServer(s, srv)
	})

	qs := pb.NewQueryClient(conn)

	res, err := qs.ListEnabledComponents(context.Background(), &pb.ListEnabledComponentsRequest{})
	g.Expect(err).To(BeNil())

	g.Expect(res.Components).NotTo(ContainElement(pb.EnabledComponent_applications))
}
