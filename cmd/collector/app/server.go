package collector_app

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/rolecollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	core_fetcher "github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime"
	_ "os"
)

type Server struct {
	roles   *rolecollector.RoleCollector
	objs    *objectscollector.ObjectsCollector
	manager clustersmngr.ClustersManager
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.roles.Start(ctx); err != nil {
		return fmt.Errorf("cannot start access rule collector: %w", err)
	}
	if err := s.objs.Start(ctx); err != nil {
		return fmt.Errorf("cannot start applications collector: %w", err)
	}
	return nil
}

func (a *Server) Status(cluster cluster.Cluster) (string, error) {
	return a.roles.Status(cluster)
}

func (s *Server) StopCollection() error {
	// These collectors can be nil if we are doing collection elsewhere.
	// Controlled by the opts.SkipCollection flag.
	if s.roles != nil {
		if err := s.roles.Stop(); err != nil {
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
	Logger            logr.Logger
	ClustersNamespace string
	Store             store.StoreWriter
}

func NewServer(ctx context.Context, opts ServerOpts) (*Server, func() error, error) {
	if opts.Store == nil {
		return nil, nil, fmt.Errorf("invalid remote store")
	}

	manager, err := createClusterManager(opts.ClustersNamespace, opts.Logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create cluster manager: %w", err)
	}
	//TODO move me to start
	manager.Start(ctx)

	optsCollector := collector.CollectorOpts{
		Log:            opts.Logger,
		ClusterManager: manager,
	}

	// create collectors
	rolesCollector, err := rolecollector.NewRoleCollector(opts.Store, optsCollector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create access rules collector: %w", err)
	}

	objsCollector, err := objectscollector.NewObjectsCollector(opts.Store, optsCollector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create applications collector: %w", err)
	}

	s := &Server{
		roles:   rolesCollector,
		objs:    objsCollector,
		manager: manager,
	}
	return s, s.StopCollection, nil
}

func createClusterManager(clustersNamespace string, log logr.Logger) (clustersmngr.ClustersManager, error) {

	clustersManagerScheme, err := kube.CreateScheme()
	if err != nil {
		return nil, fmt.Errorf("could not create scheme: %w", err)
	}

	builder := runtime.NewSchemeBuilder(
		gitopsv1alpha1.AddToScheme,
	)
	if err := builder.AddToScheme(clustersManagerScheme); err != nil {
		return nil, err
	}

	rest, clusterName, err := kube.RestConfig()

	mgmtCluster, err := cluster.NewSingleCluster(clusterName, rest, clustersManagerScheme, cluster.DefaultKubeConfigOptions...)
	if err != nil {

		return nil, fmt.Errorf("could not create mgmt cluster: %w", err)
	}

	gcf := fetcher.NewGitopsClusterFetcher(log, mgmtCluster, clustersNamespace, clustersManagerScheme, false, cluster.DefaultKubeConfigOptions...)
	scf := core_fetcher.NewSingleClusterFetcher(mgmtCluster)
	fetchers := []clustersmngr.ClusterFetcher{scf, gcf}

	clustersManager := clustersmngr.NewClustersManager(
		fetchers,
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		log,
	)

	return clustersManager, nil
}
