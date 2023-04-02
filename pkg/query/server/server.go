package server

import (
	"context"
	"fmt"
	"k8s.io/client-go/discovery"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/rolecollector"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectscollector"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type server struct {
	pb.UnimplementedQueryServer

	ac   accesschecker.Checker
	qs   query.QueryService
	arc  *rolecollector.RoleCollector
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
	Resources       []string
	SkipCollection  bool
	StoreType       string
	DiscoveryClient discovery.DiscoveryInterface
}

func (s *server) DoQuery(ctx context.Context, msg *pb.QueryRequest) (*pb.QueryResponse, error) {
	clauses := []store.QueryClause{}
	for _, c := range msg.Query {
		clauses = append(clauses, c)
	}

	objs, err := s.qs.RunQuery(ctx, clauses, msg)

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

	user := auth.Principal(ctx)

	matching := s.ac.RelevantRulesForUser(user, rules)
	return &pb.DebugGetAccessRulesResponse{
		Rules: convertToPbAccessRule(matching),
	}, nil
}

// it checks whether the resource from the policy rule allows access to the kind.
// GVKs and GVRs are related. GVKs are served under HTTP paths identified by GVRs.
// The process of mapping a GVK to a GVR is called REST mapping.
func createKindByResourceMap(dc discovery.DiscoveryInterface) (map[string]string, error) {
	_, resourcesList, err := dc.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}
	groupResourceMap := map[string]string{}
	for _, resourceList := range resourcesList {
		for _, resource := range resourceList.APIResources {
			groupResourceMap[resource.Name] = resource.Kind
		}
	}
	return groupResourceMap, nil
}

func NewServer(ctx context.Context, opts ServerOpts) (pb.QueryServer, func() error, error) {
	log := opts.Logger.WithName("query-server")

	dbDir, err := os.MkdirTemp("", "db")
	if err != nil {
		return nil, nil, err
	}

	s, err := store.NewStore(store.StorageBackendSQLite, dbDir, opts.Logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create store:%w", err)
	}

	kindByResourceMap, err := createKindByResourceMap(opts.DiscoveryClient)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create store:%w", err)
	}

	checker, err := accesschecker.NewAccessChecker(kindByResourceMap)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access checker:%w", err)
	}
	qs, err := query.NewQueryService(ctx, query.QueryServiceOpts{
		Log:           opts.Logger,
		StoreReader:   s,
		AccessChecker: checker,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query service: %w", err)
	}

	serv := &server{qs: qs, ac: checker}

	if !opts.SkipCollection {

		optsCollector := collector.CollectorOpts{
			Log:            opts.Logger,
			ClusterManager: opts.ClustersManager,
		}

		rulesCollector, err := rolecollector.NewRoleCollector(s, optsCollector)
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
		log.Info("collectors created")
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
			Kind:       o.Kind,
			Name:       o.Name,
			Namespace:  o.Namespace,
			Cluster:    o.Cluster,
			Status:     o.Status,
			ApiGroup:   o.APIGroup,
			ApiVersion: o.APIVersion,
			Message:    o.Message,
		})
	}

	return pbObjects
}

func convertToPbAccessRule(rules []models.AccessRule) []*pb.AccessRule {
	pbRules := []*pb.AccessRule{}

	for _, r := range rules {
		rule := &pb.AccessRule{
			Namespace:         r.Namespace,
			Cluster:           r.Cluster,
			AccessibleKinds:   []string{},
			Subjects:          []*pb.Subject{},
			ProvidedByRole:    r.ProvidedByRole,
			ProvidedByBinding: r.ProvidedByBinding,
		}

		for _, s := range r.Subjects {
			rule.Subjects = append(rule.Subjects, &pb.Subject{
				Kind: s.Kind,
				Name: s.Name,
			})
		}

		rule.AccessibleKinds = append(rule.AccessibleKinds, r.AccessibleKinds...)

		pbRules = append(pbRules, rule)

	}
	return pbRules
}
