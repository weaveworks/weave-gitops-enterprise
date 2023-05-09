package server

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops/core/logger"
	"k8s.io/client-go/discovery"

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
	Logger logr.Logger
	// required to watch clusters
	ClustersManager clustersmngr.ClustersManager
	SkipCollection  bool
	StoreType       string
	// required to map GVRs to GVKs for authz purporses
	DiscoveryClient discovery.DiscoveryInterface
	ObjectKinds     []configuration.ObjectKind
	ServiceAccount  collector.ImpersonateServiceAccount
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

// GVKs and GVRs are related. GVKs are served under HTTP paths identified by GVRs.
// The process of mapping a GVK to a GVR is called REST mapping.
// This method creates a map <resource,kind> to allow access checker
// to determine whether a policyRule (from GVR) allows a kind (from GVK)
// More info https://kubernetes.io/docs/reference/using-api/api-concepts/#standard-api-terminology
func createKindByResourceMap(dc discovery.DiscoveryInterface) (map[string]string, error) {
	_, resourcesList, err := dc.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}
	kindByResourceMap := map[string]string{}
	for _, resourceList := range resourcesList {
		for _, resource := range resourceList.APIResources {
			kindByResourceMap[resource.Name] = resource.Kind
		}
	}
	return kindByResourceMap, nil
}

func (so *ServerOpts) Validate() error {
	if len(so.ObjectKinds) == 0 {
		return fmt.Errorf("object kinds cannot be empty")
	}
	if so.DiscoveryClient == nil {
		return fmt.Errorf("discovery client cannot be nil")
	}
	if so.ClustersManager == nil {
		return fmt.Errorf("cluster manager cannot be nil")
	}
	if so.ServiceAccount.Name == "" {
		return fmt.Errorf("service account name cannot be empty")
	}
	if so.ServiceAccount.Namespace == "" {
		return fmt.Errorf("service account namespace cannot be empty")
	}
	return nil
}

func NewServer(opts ServerOpts) (pb.QueryServer, func() error, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid query server options: %w", err)
	}

	debug := opts.Logger.WithName("query-server").V(logger.LogLevelDebug)

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
		return nil, nil, fmt.Errorf("cannot resources mapper:%w", err)
	}

	checker, err := accesschecker.NewAccessChecker(kindByResourceMap)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access checker:%w", err)
	}
	debug.Info("access checker created")

	qs, err := query.NewQueryService(query.QueryServiceOpts{
		Log:           debug,
		StoreReader:   s,
		AccessChecker: checker,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query service: %w", err)
	}
	debug.Info("query service created")

	serv := &server{qs: qs, ac: checker}

	if !opts.SkipCollection {

		if len(opts.ObjectKinds) == 0 {
			return nil, nil, fmt.Errorf("cannot create collector for empty gvks")
		}

		rulesCollector, err := rolecollector.NewRoleCollector(s, collector.CollectorOpts{
			Log:            opts.Logger,
			ClusterManager: opts.ClustersManager,
			ServiceAccount: opts.ServiceAccount,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create access rules collector: %w", err)
		}

		if err = rulesCollector.Start(); err != nil {
			return nil, nil, fmt.Errorf("cannot start access rule collector: %w", err)
		}

		objsCollector, err := objectscollector.NewObjectsCollector(s, collector.CollectorOpts{
			Log:            opts.Logger,
			ClusterManager: opts.ClustersManager,
			ObjectKinds:    opts.ObjectKinds,
			ServiceAccount: opts.ServiceAccount,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create applications collector: %w", err)
		}

		if err = objsCollector.Start(); err != nil {
			return nil, nil, fmt.Errorf("cannot start applications collector: %w", err)
		}

		serv.arc = rulesCollector
		serv.objs = objsCollector
		debug.Info("collectors started")
	}
	debug.Info("query server created")
	return serv, serv.StopCollection, nil
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) (func() error, error) {
	s, stop, err := NewServer(opts)
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
