package query

import (
	"context"
	"fmt"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type server struct {
	pb.UnimplementedQueryServer

	qs QueryService
}

var DefaultKinds = []schema.GroupVersionKind{
	kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind),
}

type ServerOpts struct {
	Logger             logr.Logger
	StoreType          string
	ClustersManager    clustersmngr.ClustersManager
	CollectionInterval time.Duration
	ObjectKinds        []schema.GroupVersionKind
	SkipCollection     bool
}

func (s *server) Run(ctx context.Context, in *pb.QueryRequest) (*pb.QueryResponse, error) {
	objs, err := s.qs.RunQuery(ctx)
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

	col := NewCollector(opts.Logger, opts.ClustersManager, CollectorOpts{
		ObjectKinds: opts.ObjectKinds,
	})

	var w StoreWriter
	var r StoreReader

	switch opts.StoreType {
	case "memory":
		s := store.NewInMemoryStore()
		w = s
		r = s
	default:
		return nil, fmt.Errorf("unknown store type: %s", opts.StoreType)
	}

	qs, err := NewQueryService(ctx, opts.Logger, col, w, r, time.NewTicker(opts.CollectionInterval), time.NewTicker(opts.CollectionInterval))
	if err != nil {
		return nil, fmt.Errorf("failed to create query service: %w", err)
	}

	if !opts.SkipCollection {
		if err := qs.Start(); err != nil {
			return nil, fmt.Errorf("failed to start query service: %w", err)
		}
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
