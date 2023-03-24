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

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type server struct {
	pb.UnimplementedQueryServer

	qs  query.QueryService
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

	matching := accesschecker.NewAccessChecker().RelevantRulesForUser(user, rules)

	return &pb.DebugGetAccessRulesResponse{
		Rules: convertToPbAccessRule(matching),
	}, nil
}

func (s *server) StoreRoles(ctx context.Context, msg *pb.StoreRolesRequest) (*pb.StoreRolesResponse, error) {
	if len(msg.GetRoles()) == 0 {
		s.log.Info("ignored store roles request as empty")
		return &pb.StoreRolesResponse{}, nil
	}
	roles := convertToRoles(msg.GetRoles())
	err := s.qs.StoreRoles(ctx, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to store roles: %w", err)
	}
	return &pb.StoreRolesResponse{}, nil
}

func (s *server) StoreRoleBindings(ctx context.Context, msg *pb.StoreRoleBindingsRequest) (*pb.StoreRoleBindingsResponse, error) {
	if len(msg.GetRolebindings()) == 0 {
		s.log.Info("ignored store roles bindings as empty")
		return &pb.StoreRoleBindingsResponse{}, nil
	}
	rbs := convertToRoleBindings(msg.GetRolebindings())
	err := s.qs.StoreRoleBindings(ctx, rbs)
	if err != nil {
		return nil, fmt.Errorf("failed to store role bindings: %w", err)
	}
	return &pb.StoreRoleBindingsResponse{}, nil
}

func (s *server) StoreObjects(ctx context.Context, msg *pb.StoreObjectsRequest) (*pb.StoreObjectsResponse, error) {
	if len(msg.GetObjects()) == 0 {
		s.log.Info("ignored store objects request as empty")
		return &pb.StoreObjectsResponse{}, nil
	}

	objs := convertToObjects(msg.GetObjects())
	err := s.qs.StoreObjects(ctx, objs)
	if err != nil {
		return nil, fmt.Errorf("failed to store objects: %w", err)
	}
	return &pb.StoreObjectsResponse{}, nil
}

func (s *server) DeleteObjects(ctx context.Context, msg *pb.DeleteObjectsRequest) (*pb.DeleteObjectsResponse, error) {
	if len(msg.GetObjects()) == 0 {
		s.log.Info("ignored delete objects request as empty")
		return &pb.DeleteObjectsResponse{}, nil
	}

	objs := convertToObjects(msg.GetObjects())
	err := s.qs.DeleteObjects(ctx, objs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete objects: %w", err)
	}
	return &pb.DeleteObjectsResponse{}, nil
}

func (s *server) DeleteRoles(ctx context.Context, msg *pb.DeleteRolesRequest) (*pb.DeleteRolesResponse, error) {
	if len(msg.GetRoles()) == 0 {
		s.log.Info("ignored delete roles request as empty")
		return &pb.DeleteRolesResponse{}, nil
	}

	roles := convertToRoles(msg.GetRoles())
	err := s.qs.DeleteRoles(ctx, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to delete roles: %w", err)
	}
	return &pb.DeleteRolesResponse{}, nil
}

func (s *server) DeleteRoleBindings(ctx context.Context, msg *pb.DeleteRoleBindingsRequest) (*pb.DeleteRoleBindingsResponse, error) {
	if len(msg.GetRolebindings()) == 0 {
		s.log.Info("ignored delete rolebindings request as empty")
		return &pb.DeleteRoleBindingsResponse{}, nil
	}

	roleBindings := convertToRoleBindings(msg.GetRolebindings())
	err := s.qs.DeleteRoleBindings(ctx, roleBindings)
	if err != nil {

		return nil, fmt.Errorf("failed to delete rolebindings: %w", err)
	}
	return &pb.DeleteRoleBindingsResponse{}, nil
}

func NewServer(ctx context.Context, opts ServerOpts) (pb.QueryServer, func() error, error) {
	dbDir, err := os.MkdirTemp("", "db")
	if err != nil {
		return nil, nil, err
	}

	s, err := store.NewStore(store.StorageBackendSQLite, store.StoreOpts{
		Url: dbDir,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create store:%w", err)
	}

	qs, err := query.NewQueryService(ctx, query.QueryServiceOpts{
		Log:   opts.Logger,
		Store: s,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query service: %w", err)
	}

	serv := &server{qs: qs, log: opts.Logger}

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
			Cluster:    o.Cluster,
			Namespace:  o.Namespace,
			APIGroup:   o.ApiGroup,
			APIVersion: o.ApiVersion,
			Kind:       o.Kind,
			Name:       o.Name,
			Status:     o.Status,
			Message:    o.Message,
		})
	}

	return objects
}

func convertToPbAccessRule(rules []models.AccessRule) []*pb.AccessRule {
	pbRules := []*pb.AccessRule{}

	for _, r := range rules {
		rule := &pb.AccessRule{
			Namespace:       r.Namespace,
			Cluster:         r.Cluster,
			AccessibleKinds: []string{},
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

func convertToRoles(pbRoles []*pb.Role) []models.Role {
	roles := []models.Role{}

	for _, r := range pbRoles {
		role := models.Role{
			Cluster:     r.Cluster,
			Namespace:   r.Namespace,
			Kind:        r.Kind,
			Name:        r.Name,
			PolicyRules: []models.PolicyRule{},
		}

		for _, pr := range r.PolicyRules {
			role.PolicyRules = append(role.PolicyRules, models.PolicyRule{
				APIGroups: pr.ApiGroups,
				Resources: pr.Resources,
				Verbs:     pr.Verbs,
				RoleID:    pr.RoleId,
			})
		}

		roles = append(roles, role)

	}
	return roles
}

func convertToRoleBindings(pbRolesBindings []*pb.RoleBinding) []models.RoleBinding {
	roleBindings := []models.RoleBinding{}

	for _, r := range pbRolesBindings {
		roleBinding := models.RoleBinding{
			Cluster:     r.Cluster,
			Namespace:   r.Namespace,
			Kind:        r.Kind,
			Name:        r.Name,
			RoleRefName: r.RoleRefName,
			RoleRefKind: r.RoleRefKind,
			Subjects:    []models.Subject{},
		}

		for _, subject := range r.Subjects {
			roleBinding.Subjects = append(roleBinding.Subjects, models.Subject{
				Kind:          subject.Kind,
				Name:          subject.Name,
				Namespace:     subject.Namespace,
				APIGroup:      subject.ApiGroup,
				RoleBindingID: subject.RoleBindingId,
			})
		}

		roleBindings = append(roleBindings, roleBinding)

	}
	return roleBindings
}
