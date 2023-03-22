package server

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

type server struct {
	pb.UnimplementedQueryServer

	qs  query.QueryService
	ss  query.StoreService
	log logr.Logger
}

func (s *server) Stop() error {
	return nil
}

type ServerOpts struct {
	Logger          logr.Logger
	ClustersManager clustersmngr.ClustersManager
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

func (s *server) DebugGetAccessRules(ctx context.Context, msg *pb.DebugGetAccessRulesRequest) (*pb.DebugGetAccessRulesResponse, error) {
	rules, err := s.qs.GetAccessRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access rules: %w", err)
	}

	return &pb.DebugGetAccessRulesResponse{
		Rules: convertToPbAccessRule(rules),
	}, nil
}

func (s *server) StoreAccessRules(ctx context.Context, msg *pb.StoreAccessRulesRequest) (*pb.StoreAccessRulesResponse, error) {
	if len(msg.GetRules()) == 0 {
		s.log.Info("ignored store access rules request as empty")
		return &pb.StoreAccessRulesResponse{}, nil
	}
	rules := convertToAccessRules(msg.GetRules())
	err := s.ss.StoreAccessRules(ctx, rules)
	if err != nil {
		return nil, fmt.Errorf("failed to store access rules: %w", err)
	}
	return &pb.StoreAccessRulesResponse{}, nil
}

func (s *server) StoreObjects(ctx context.Context, msg *pb.StoreObjectsRequest) (*pb.StoreObjectsResponse, error) {
	objs := convertToObjects(msg.GetObjects())
	err := s.ss.StoreObjects(ctx, objs)
	if err != nil {
		return nil, fmt.Errorf("failed to store objects: %w", err)
	}
	return &pb.StoreObjectsResponse{}, nil
}

func NewServer(ctx context.Context, opts ServerOpts) (pb.QueryServer, func() error, error) {
	dbDir, err := os.MkdirTemp("", "db")
	if err != nil {
		return nil, nil, err
	}

	s, err := store.NewStore(store.StorageBackendSQLite, dbDir, opts.Logger)
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

	ss, err := query.NewStoreService(ctx, query.StoreServiceOpts{
		Log:         opts.Logger,
		StoreWriter: s,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create store service: %w", err)
	}

	serv := &server{qs: qs, ss: ss, log: opts.Logger}

	return serv, serv.Stop, nil
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

func convertToObjects(pbObj []*pb.Object) []models.Object {
	objects := []models.Object{}

	for _, o := range pbObj {
		objects = append(objects, models.Object{
			Kind:      o.Kind,
			Name:      o.Name,
			Namespace: o.Namespace,
			Cluster:   o.Cluster,
			Status:    o.Status,
		})
	}

	return objects
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

func convertToAccessRules(pbrules []*pb.AccessRule) []models.AccessRule {
	rules := []models.AccessRule{}

	for _, r := range pbrules {
		rule := models.AccessRule{
			Principal:       r.Principal,
			Namespace:       r.Namespace,
			Cluster:         r.Cluster,
			AccessibleKinds: []string{},
		}

		rule.AccessibleKinds = append(rule.AccessibleKinds, r.AccessibleKinds...)

		rules = append(rules, rule)

	}
	return rules
}
