package server

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

const DefaultPipelineControllerAddress = "chart-pipeline-controller-promotion:8082"

type ServerOpts struct {
	logr.Logger
	ClustersManager           clustersmngr.ClustersManager
	ManagementFetcher         *mgmtfetcher.ManagementCrossNamespacesFetcher
	Cluster                   string
	PipelineControllerAddress string
	GitProvider               git.Provider
}

type server struct {
	pb.UnimplementedPipelinesServer

	log                       logr.Logger
	clients                   clustersmngr.ClustersManager
	managementFetcher         *mgmtfetcher.ManagementCrossNamespacesFetcher
	cluster                   string
	pipelineControllerAddress string
	gitProvider               git.Provider
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewPipelinesServer(opts)

	return pb.RegisterPipelinesHandlerServer(ctx, mux, s)
}

func NewPipelinesServer(opts ServerOpts) pb.PipelinesServer {
	return &server{
		log:                       opts.Logger,
		clients:                   opts.ClustersManager,
		managementFetcher:         opts.ManagementFetcher,
		cluster:                   opts.Cluster,
		pipelineControllerAddress: opts.PipelineControllerAddress,
		gitProvider:               opts.GitProvider,
	}
}
