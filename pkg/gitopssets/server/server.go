package server

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory    clustersmngr.ClustersManager
	ManagementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
	Scheme            *k8sruntime.Scheme
	Cluster           string
}

type server struct {
	pb.UnimplementedGitOpsSetsServer

	log               logr.Logger
	clients           clustersmngr.ClustersManager
	managementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
	scheme            *k8sruntime.Scheme
	cluster           string
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewGitOpsSetsServer(opts)

	return pb.RegisterGitOpsSetsHandlerServer(ctx, mux, s)
}

func NewGitOpsSetsServer(opts ServerOpts) pb.GitOpsSetsServer {
	return &server{
		log:               opts.Logger,
		clients:           opts.ClientsFactory,
		managementFetcher: opts.ManagementFetcher,
		scheme:            opts.Scheme,
	}
}
