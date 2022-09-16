package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/types"
)

func TestListChartsForRepository(t *testing.T) {
	testCases := []struct {
		name    string
		fc      *fakeChartCache
		request *protos.ListChartsForRepositoryRequest
		want    *protos.ListChartsForRepositoryResponse
	}{
		{
			name: "matching cluster and repo",
			request: &protos.ListChartsForRepositoryRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Namespace: "clusters",
						Name:      "demo-cluster",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "chart",
			},
			fc: newFakeChartCache(
				objectRefToString(
					ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
					types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
				), []Chart{{Name: "redis", Version: "1.0.1"}, {Name: "postgres", Version: "1.0.2"}}),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{
					{Name: "redis", Versions: []string{"1.0.1"}},
					{Name: "postgres", Versions: []string{"1.0.2"}},
				},
			},
		},
		{
			name: "multiple versions of the same chart",
			request: &protos.ListChartsForRepositoryRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Namespace: "clusters",
						Name:      "demo-cluster",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "chart",
			},
			fc: newFakeChartCache(
				objectRefToString(
					ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
					types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
				), []Chart{{Name: "redis", Version: "1.0.1"}, {Name: "redis", Version: "1.0.2"}}),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{
					{Name: "redis", Versions: []string{"1.0.1", "1.0.2"}},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			s := createServer(t, serverOptions{
				chartsCache: tt.fc,
			})

			response, err := s.ListChartsForRepository(context.TODO(), tt.request)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, response, protocmp.Transform()); diff != "" {
				t.Fatalf("failed to get response:\n%s", diff)
			}
		})
	}
}

func newFakeChartCache(key string, charts []Chart) *fakeChartCache {
	return &fakeChartCache{
		charts: map[string][]Chart{
			key: charts,
		},
	}
}

type fakeChartCache struct {
	charts map[string][]Chart
}

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

func (fc fakeChartCache) ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error) {
	if charts, ok := fc.charts[objectRefToString(repoRef, clusterRef)]; ok {
		return charts, nil
	}
	return nil, errors.New("no charts found")
}

func objectRefToString(or ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}
