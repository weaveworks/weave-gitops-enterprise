package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// ListChartsForRepository returns a list of charts for a given repository.
func (s *server) ListChartsForRepository(ctx context.Context, request *protos.ListChartsForRepositoryRequest) (*protos.ListChartsForRepositoryResponse, error) {
	clusterRef := types.NamespacedName{
		Name:      request.Repository.Cluster.Name,
		Namespace: request.Repository.Cluster.Namespace,
	}

	repoRef := helm.ObjectReference{
		Kind:      request.Repository.Kind,
		Name:      request.Repository.Name,
		Namespace: request.Repository.Namespace,
	}

	charts, err := s.chartsCache.ListChartsByRepositoryAndCluster(ctx, clusterRef, repoRef, request.Kind)
	if err != nil {
		// FIXME: does this work?
		if err.Error() == "no charts found" {
			return &protos.ListChartsForRepositoryResponse{}, nil
		}
		return nil, fmt.Errorf("error listing charts: %w", err)
	}

	chartsWithVersions := map[string][]string{}
	for _, chart := range charts {
		chartsWithVersions[chart.Name] = append(chartsWithVersions[chart.Name], chart.Version)
	}

	responseCharts := []*protos.RepositoryChart{}
	for name, versions := range chartsWithVersions {
		sortedVersions, err := helm.ReverseSemVerSort(versions)
		if err != nil {
			return nil, fmt.Errorf("parsing chart %s: %w", name, err)
		}

		responseCharts = append(responseCharts, &protos.RepositoryChart{
			Name:     name,
			Versions: sortedVersions,
		})
	}

	sort.Slice(responseCharts, func(i, j int) bool {
		return responseCharts[i].Name < responseCharts[j].Name
	})

	return &protos.ListChartsForRepositoryResponse{Charts: responseCharts}, nil
}

// GetValuesForChart returns the values for a given chart.
func (s *server) GetValuesForChart(ctx context.Context, req *protos.GetValuesForChartRequest) (*protos.GetValuesForChartResponse, error) {
	if req.Repository == nil || req.Repository.Cluster == nil {
		return nil, fmt.Errorf("repository or cluster is nil")
	}

	clusterRef := types.NamespacedName{
		Name:      req.Repository.Cluster.Name,
		Namespace: req.Repository.Cluster.Namespace,
	}

	repoRef := helm.ObjectReference{
		Kind:      req.Repository.Kind,
		Name:      req.Repository.Name,
		Namespace: req.Repository.Namespace,
	}

	chart := helm.Chart{
		Name:    req.Name,
		Version: req.Version,
	}

	// FIXME: should be looking up the actual helm repository here to check RBAC

	jobId := s.chartJobs.New()

	go func() {
		res, err := s.GetOrFetchValues(context.Background(), repoRef, clusterRef, chart)
		s.chartJobs.Set(jobId, helm.JobResult{Result: res, Error: err})
	}()

	return &protos.GetValuesForChartResponse{JobId: jobId}, nil
}

func (s *server) GetChartsJob(ctx context.Context, req *protos.GetChartsJobRequest) (*protos.GetChartsJobResponse, error) {
	result, found := s.chartJobs.Get(req.JobId)
	if !found {
		return nil, &grpcruntime.HTTPStatusError{
			Err:        errors.New("job not found"),
			HTTPStatus: http.StatusOK,
		}
	}

	errString := ""
	if result.Error != nil {
		errString = result.Error.Error()
	}

	return &protos.GetChartsJobResponse{Values: result.Result, Error: errString}, nil
}

func (s *server) GetOrFetchValues(ctx context.Context, repoRef helm.ObjectReference, clusterRef types.NamespacedName, chart helm.Chart) (string, error) {
	values, err := s.chartsCache.GetChartValues(ctx, clusterRef, repoRef, chart)
	if err != nil {
		return "", fmt.Errorf("error getting chart values: %w", err)
	}

	if values != nil {
		return string(values), nil
	}

	config, err := s.GetClientConfigForCluster(ctx, clusterRef)
	if err != nil {
		return "", fmt.Errorf("error getting client config for cluster: %w", err)
	}

	data, err := s.valuesFetcher.GetValuesFile(ctx, config, types.NamespacedName{Namespace: repoRef.Namespace, Name: repoRef.Name}, chart, clusterRef.Name != helm.ManagementClusterName)
	if err != nil {
		return "", fmt.Errorf("error fetching values file: %w", err)
	}

	err = s.chartsCache.UpdateValuesYaml(ctx, clusterRef, repoRef, chart, data)
	if err != nil {
		return "", fmt.Errorf("error updating values yaml: %w", err)
	}

	return string(data), nil
}

// GetClientConfigForCluster returns the client config for a given cluster.
func (s *server) GetClientConfigForCluster(ctx context.Context, cluster types.NamespacedName) (*rest.Config, error) {
	cfg, _, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting rest config: %w", err)
	}
	managementCluster := clustersmngr.Cluster{
		Name:        helm.ManagementClusterName,
		Server:      cfg.Host,
		BearerToken: cfg.BearerToken,
		TLSConfig:   cfg.TLSClientConfig,
	}

	clusters := []clustersmngr.Cluster{managementCluster}

	// clusters := s.clustersManager.GetClusters()

	clusterName := cluster.Name
	if clusterName != helm.ManagementClusterName {
		clusterName = fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name)
	}

	for _, c := range clusters {
		if c.Name == clusterName {
			return clustersmngr.ClientConfigAsServer()(c)
		}
	}

	return nil, fmt.Errorf("cluster %s not found", clusterName)
}
