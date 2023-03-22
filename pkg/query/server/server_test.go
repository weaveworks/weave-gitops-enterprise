package server

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestServer_StoreAccessRules(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	log := testr.New(t)

	queryClient, _ := setup(t, ctx, log)
	g.Expect(queryClient).NotTo(BeNil())

	tests := []struct {
		name       string
		request    *pb.StoreAccessRulesRequest
		errPattern string
	}{
		{
			name: "cannot store access rules for empty rules",
			request: &pb.StoreAccessRulesRequest{
				Rules: []*pb.AccessRule{},
			},
			errPattern: "empty slice found",
		},
		{
			name: "can store access rule with rule",
			request: &pb.StoreAccessRulesRequest{
				Rules: []*pb.AccessRule{
					{
						Cluster:         "dev-cluster",
						Principal:       "john",
						Namespace:       "search",
						AccessibleKinds: []string{"HelmRelease"},
					},
				},
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesResponse, err := queryClient.StoreAccessRules(ctx, tt.request)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(rulesResponse).NotTo(BeNil())
		})
	}
}

func TestServer_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	log := testr.New(t)

	queryClient, _ := setup(t, ctx, log)
	g.Expect(queryClient).NotTo(BeNil())
	tests := []struct {
		name       string
		request    *pb.StoreObjectsRequest
		errPattern string
	}{
		{
			name: "cannot store object for empty objects",
			request: &pb.StoreObjectsRequest{
				Objects: []*pb.Object{},
			},
			errPattern: "empty slice found",
		},
		{
			name: "can store object rule with object",
			request: &pb.StoreObjectsRequest{
				Objects: []*pb.Object{
					{
						Cluster:   "dev-cluster",
						Name:      "podinfo",
						Namespace: "default",
						Kind:      "HelmRelease",
						Status:    "released",
					},
				},
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := queryClient.StoreObjects(ctx, tt.request)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(response).NotTo(BeNil())
		})
	}
}

func setup(t *testing.T, ctx context.Context, log logr.Logger) (pb.QueryClient, client.Client) {
	k8s, fakeClusterManager := grpctesting.MakeFactoryWithObjects()
	opts := ServerOpts{
		Logger:          log,
		ClustersManager: fakeClusterManager,
	}
	srv, _, _ := NewServer(ctx, opts)
	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterQueryServer(s, srv)
	}, WithClientsPoolInterceptor(&auth.UserPrincipal{ID: "bob"}))
	return pb.NewQueryClient(conn), k8s
}

func WithClientsPoolInterceptor(user *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = auth.WithPrincipal(ctx, user)
		return handler(ctx, req)
	})
}
