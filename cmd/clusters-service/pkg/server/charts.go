package server

import (
	"context"
	"fmt"

	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	wegohelm "github.com/weaveworks/weave-gitops/pkg/helm"
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
		if err.Error() == "no charts found" {
			return &protos.ListChartsForRepositoryResponse{}, nil
		}
		return nil, err
	}

	chartsWithVersions := map[string][]string{}
	for _, chart := range charts {
		if request.Kind != "" {
			if chart.Kind == request.Kind {
				chartsWithVersions[chart.Name] = append(chartsWithVersions[chart.Name], chart.Version)
			}
		} else {
			chartsWithVersions[chart.Name] = append(chartsWithVersions[chart.Name], chart.Version)
		}
	}

	responseCharts := []*protos.RepositoryChart{}
	for name, versions := range chartsWithVersions {
		sortedVersions, err := wegohelm.ReverseSemVerSort(versions)
		if err != nil {
			return nil, fmt.Errorf("parsing chart %s: %w", name, err)
		}

		responseCharts = append(responseCharts, &protos.RepositoryChart{
			Name:     name,
			Versions: sortedVersions,
		})
	}

	return &protos.ListChartsForRepositoryResponse{Charts: responseCharts}, nil
}
