package server

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectscollector"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type server struct {
	pb.UnimplementedQueryServer

	qs   query.QueryService
	arc  *accesscollector.AccessRulesCollector
	objs *objectscollector.ObjectsCollector
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

type ServerOpts struct {
	Logger          logr.Logger
	ClustersManager clustersmngr.ClustersManager
	ObjectKinds     []schema.GroupVersionKind
	SkipCollection  bool
	StoreType       string
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
	dbDir, err := os.MkdirTemp("", "db")
	if err != nil {
		return nil, nil, err
	}

	s, _, err := store.NewStore(dbDir, opts.Logger)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot create store:%w", err)
	}

	qs, err := query.NewQueryService(ctx, query.QueryServiceOpts{
		Log:         opts.Logger,
		StoreReader: s,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query service: %w", err)
	}

	serv := &server{qs: qs}

	if !opts.SkipCollection {

		optsCollector := collector.CollectorOpts{
			Log: opts.Logger,
			// ClusterManager: opts.ClustersManager,
		}
		// create collectors
		rulesCollector, err := accesscollector.NewAccessRulesCollector(s, optsCollector)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create access rules collector: %w", err)
		}

		if err = rulesCollector.Start(ctx); err != nil {
			return nil, nil, fmt.Errorf("cannot start access rule collector: %w", err)
		}

		objsCollector, err := objectscollector.NewObjectsCollector(s, optsCollector)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create applications collector: %w", err)
		}

		if err = objsCollector.Start(ctx); err != nil {
			return nil, nil, fmt.Errorf("cannot start applications collector: %w", err)
		}

		serv.arc = rulesCollector
		serv.objs = objsCollector
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
