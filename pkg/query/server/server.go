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
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectcollector"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type server struct {
	pb.UnimplementedQueryServer

	qs   query.QueryService
	arc  *accesscollector.AccessRulesCollector
	objs *objectcollector.ObjectCollector
}

func (s *server) StopCollection() error {
	// These collectors can be nil if we are doing collection elsewhere.
	// Controlled by the opts.SkipCollection flag.
	if s.arc != nil {
		if err := s.arc.Stop(); err != nil {
			return fmt.Errorf("failed to stop access rules collection: %w", err)
		}
	}

	if s.objs != nil {
		if err := s.objs.Stop(); err != nil {
			return fmt.Errorf("failed to stop object collection: %w", err)
		}
	}

	return nil
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
	// Go complains about using msq.Query directly, so we have to copy it into a slice.
	// query.Query is specifically designed to fit msg.Query.
	q := []query.Query{}
	for _, qm := range msg.Query {
		q = append(q, qm)
	}

	objs, err := s.qs.RunQuery(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}

	return &pb.QueryResponse{
		Objects: convertToPbObject(objs),
	}, nil
}

func (s *server) DebugGetAccessRules(ctx context.Context, msg *pb.DebugGetAccessRulesRequest) (*pb.DebugGetAccessRulesResponse, error) {
	rules, err := s.qs.GetAccessRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access rules: %w", err)
	}

	return &pb.DebugGetAccessRulesResponse{
		Rules: convertToPbAccessRule(rules),
	}, nil
}

func NewServer(ctx context.Context, opts ServerOpts) (pb.QueryServer, func() error, error) {
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
		return nil, nil, fmt.Errorf("unknown store type: %s", opts.StoreType)
	}

	qs, err := query.NewQueryService(ctx, query.QueryServiceOpts{
		Log:         opts.Logger,
		StoreReader: r,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query service: %w", err)
	}

	serv := &server{qs: qs}

	if !opts.SkipCollection {
		arc := accesscollector.NewAccessRulesCollector(w, collector.CollectorOpts{
			Log:            opts.Logger,
			ClusterManager: opts.ClustersManager,
			PollInterval:   opts.CollectionInterval,
		})
		arc.Start()

		objCollector := objectcollector.NewObjectCollector(opts.Logger, opts.ClustersManager, w, nil)
		objCollector.Start()

		serv.arc = arc
		serv.objs = objCollector
	}

	return serv, serv.StopCollection, nil
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) (func() error, error) {
	s, stop, err := NewServer(ctx, opts)
	if err != nil {
		return nil, err
	}

	return stop, pb.RegisterQueryHandlerServer(ctx, mux, s)
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

func convertToPbAccessRule(rules []models.AccessRule) []*pb.AccessRule {
	pbRules := []*pb.AccessRule{}

	for _, r := range rules {
		rule := &pb.AccessRule{
			Principal:       r.Principal,
			Namespace:       r.Namespace,
			Cluster:         r.Cluster,
			AccessibleKinds: []string{},
		}

		rule.AccessibleKinds = append(rule.AccessibleKinds, r.AccessibleKinds...)

		pbRules = append(pbRules, rule)

	}
	return pbRules
}
