package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/types"
)

// ListChartsForRepository returns a list of charts for a given repository.
func (s *server) ListChartsForRepository(ctx context.Context, req *protos.ListChartsForRepositoryRequest) (*protos.ListChartsForRepositoryResponse, error) {
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

	err := s.checkUserCanAccessHelmRepo(ctx, clusterRef, repoRef)
	if err != nil {
		return nil, fmt.Errorf("error checking user can access helm repo: %w", err)
	}

	charts, err := s.chartsCache.ListChartsByRepositoryAndCluster(ctx, clusterRef, repoRef, req.Kind)
	if err != nil {
		// FIXME: does this work?
		if err.Error() == "no charts found" {
			return &protos.ListChartsForRepositoryResponse{}, nil
		}
		return nil, fmt.Errorf("error listing charts: %w", err)
	}

	chartsWithVersions := map[string][]string{}
	chartsLayers := map[string]string{}
	for _, chart := range charts {
		chartsWithVersions[chart.Name] = append(chartsWithVersions[chart.Name], chart.Version)
		chartsLayers[chart.Name] = chart.Layer
	}

	responseCharts := []*protos.RepositoryChart{}
	for name, versions := range chartsWithVersions {
		sortedVersions, err := helm.ReverseSemVerSort(versions)
		if err != nil {
			return nil, fmt.Errorf("parsing chart %s: %w", name, err)
		}

		responseCharts = append(responseCharts, &protos.RepositoryChart{
			Name:     name,
			Layer:    chartsLayers[name],
			Versions: sortedVersions,
			RepoName: req.Repository.Name,
			RepoNamespace: req.Repository.Namespace,
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

	err := s.checkUserCanAccessHelmRepo(ctx, clusterRef, repoRef)
	if err != nil {
		return nil, fmt.Errorf("error checking user can access helm repo: %w", err)
	}

	found, err := s.chartsCache.IsKnownChart(ctx, clusterRef, repoRef, chart)
	if err != nil {
		return nil, fmt.Errorf("error checking if chart is known: %w", err)
	}
	if !found {
		return nil, &grpcruntime.HTTPStatusError{
			Err:        errors.New("chart version not found"),
			HTTPStatus: http.StatusNotFound,
		}
	}

	jobId := s.chartJobs.New()

	go func() {
		res, err := s.GetOrFetchValues(context.Background(), repoRef, clusterRef, chart)
		s.chartJobs.Set(jobId, helm.JobResult{Result: res, Error: err})
	}()

	return &protos.GetValuesForChartResponse{JobId: jobId}, nil
}

func (s *server) GetChartsJob(ctx context.Context, req *protos.GetChartsJobRequest) (*protos.GetChartsJobResponse, error) {
	// We're do a little bit of security obscurity here. Job IDs are UUIDs, so
	// "it's not possible to guess the ID of a job that's not yours".
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

func (s *server) checkUserCanAccessHelmRepo(ctx context.Context, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	client, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), fetcher.ToClusterName(clusterRef))
	if err != nil {
		return fmt.Errorf("error getting impersonated client for cluster: %w", err)
	}

	// Get the helmRepo from the cluster
	hr := sourcev1.HelmRepository{}
	err = client.Get(ctx, fetcher.ToClusterName(clusterRef), types.NamespacedName{Name: repoRef.Name, Namespace: repoRef.Namespace}, &hr)
	if err != nil {
		return fmt.Errorf("error getting helm repository: %w", err)
	}

	return nil
}

func (s *server) GetOrFetchValues(ctx context.Context, repoRef helm.ObjectReference, clusterRef types.NamespacedName, chart helm.Chart) (string, error) {
	values, err := s.chartsCache.GetChartValues(ctx, clusterRef, repoRef, chart)
	if err != nil {
		return "", fmt.Errorf("error getting chart values: %w", err)
	}

	if values != nil {
		return string(values), nil
	}

	cluster, err := s.GetCluster(ctx, clusterRef)
	if err != nil {
		return "", fmt.Errorf("error getting client config for cluster: %w", err)
	}

	data, err := s.valuesFetcher.GetValuesFile(ctx, cluster, types.NamespacedName{Namespace: repoRef.Namespace, Name: repoRef.Name}, chart, clusterRef.Name != s.cluster)
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
func (s *server) GetCluster(ctx context.Context, cluster types.NamespacedName) (cluster.Cluster, error) {
	clusters := s.clustersManager.GetClusters()
	clusterName := fetcher.ToClusterName(cluster)

	for _, c := range clusters {
		if c.GetName() == clusterName {
			return c, nil
		}
	}

	return nil, fmt.Errorf("cluster %s not found", clusterName)
}
