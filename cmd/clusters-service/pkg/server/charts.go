package server

import (
	"context"

	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"k8s.io/apimachinery/pkg/types"
)

// ListChartsForRepository returns a list of charts for a given repository.
func (s *server) ListChartsForRepository(ctx context.Context, request *protos.ListChartsForRepositoryRequest) (*protos.ListChartsForRepositoryResponse, error) {
	clusterRef := types.NamespacedName{
		Name:      request.Repository.Cluster.Name,
		Namespace: request.Repository.Cluster.Namespace,
	}

	repoRef := ObjectReference{
		Kind:      request.Repository.Kind,
		Name:      request.Repository.Name,
		Namespace: request.Repository.Namespace,
	}

	charts, err := s.chartsCache.ListChartsByRepositoryAndCluster(ctx, repoRef, clusterRef)
	if err != nil {
		return nil, err
	}

	chartsWithVersions := map[string][]string{}
	for _, chart := range charts {
		chartsWithVersions[chart.Name] = append(chartsWithVersions[chart.Name], chart.Version)
	}

	responseCharts := []*protos.RepositoryChart{}
	for name, versions := range chartsWithVersions {
		responseCharts = append(responseCharts, &protos.RepositoryChart{
			Name:     name,
			Versions: versions,
		})
	}

	return &protos.ListChartsForRepositoryResponse{Charts: responseCharts}, nil
}
