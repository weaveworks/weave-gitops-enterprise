package server

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

type ServerOpts struct {
	logr.Logger
	ClustersManager   clustersmngr.ClustersManager
	ManagementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
}

type server struct {
	pb.UnimplementedPipelinesServer

	log               logr.Logger
	clients           clustersmngr.ClustersManager
	managementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewPipelinesServer(opts)

	return pb.RegisterPipelinesHandlerServer(ctx, mux, s)
}

func NewPipelinesServer(opts ServerOpts) pb.PipelinesServer {
	return &server{
		log:               opts.Logger,
		clients:           opts.ClustersManager,
		managementFetcher: opts.ManagementFetcher,
	}
}
