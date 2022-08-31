package terraform

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory clustersmngr.ClientsFactory
}

type server struct {
	pb.UnimplementedTerraformServer

	log     logr.Logger
	clients clustersmngr.ClientsFactory
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewTerraformServer(opts)

	return pb.RegisterTerraformHandlerServer(ctx, mux, s)
}

func NewTerraformServer(opts ServerOpts) pb.TerraformServer {
	return &server{
		log:     opts.Logger,
		clients: opts.ClientsFactory,
	}
}

func (s *server) ListTerraformObjects(ctx context.Context, msg *pb.ListTerraformObjectsRequest) (*pb.ListTerraformObjectsResponse, error) {
	return &pb.ListTerraformObjectsResponse{}, nil
}
