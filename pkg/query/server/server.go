package server

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectcollector"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type server struct {
	pb.UnimplementedQueryServer

	qs query.QueryService
}

var DefaultKinds = []schema.GroupVersionKind{
	kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind),
	helmv2.GroupVersion.WithKind(helmv2.HelmReleaseKind),
}

type ServerOpts struct {
	Logger             logr.Logger
	StoreType          string
	ClustersManager    clustersmngr.ClustersManager
	CollectionInterval time.Duration
	ObjectKinds        []schema.GroupVersionKind
	SkipCollection     bool
}

func (s *server) DoQuery(ctx context.Context, msg *pb.QueryRequest) (*pb.QueryResponse, error) {
	objs, err := s.qs.RunQuery(ctx, msg.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}

	return &pb.QueryResponse{
		Objects: convertToPbObject(objs),
	}, nil
}

func NewServer(ctx context.Context, opts ServerOpts) (pb.QueryServer, error) {
	if opts.ObjectKinds == nil {
		opts.ObjectKinds = DefaultKinds
	}

	var w store.StoreWriter
	var r store.StoreReader

	switch opts.StoreType {
	case "memory":
		s := memorystore.NewInMemoryStore()
		w = s
		r = s
	default:
		return nil, fmt.Errorf("unknown store type: %s", opts.StoreType)
	}

	qs, err := query.NewQueryService(ctx, query.QueryServiceOpts{
		Log:         opts.Logger,
		StoreReader: r,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create query service: %w", err)
	}

	if !opts.SkipCollection {
		objCollector := objectcollector.NewObjectCollector(opts.Logger, opts.ClustersManager, w, nil)
		objCollector.Start()
	}

	return &server{qs: qs}, nil
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s, err := NewServer(ctx, opts)
	if err != nil {
		return err
	}

	return pb.RegisterQueryHandlerServer(ctx, mux, s)
}

func convertToPbObject(obj []models.Object) []*pb.Object {
	pbObjects := []*pb.Object{}

	for _, o := range obj {
		pbObjects = append(pbObjects, &pb.Object{
			Kind:      o.Kind,
			Name:      o.Name,
			Namespace: o.Namespace,
			Cluster:   o.Cluster,
			Status:    o.Status,
		})
	}

	return pbObjects
}
