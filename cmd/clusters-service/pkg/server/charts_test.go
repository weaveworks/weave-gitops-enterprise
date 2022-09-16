package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
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
					ClusterName: "demo-cluster",
					Name:        "bitnami-charts",
					Namespace:   "demo",
					Kind:        "HelmRepository",
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
					ClusterName: "demo-cluster",
					Name:        "bitnami-charts",
					Namespace:   "demo",
					Kind:        "HelmRepository",
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
					{Name: "redis", Versions: []string{"1.0.2", "1.0.1"}},
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
			if diff := cmp.Diff(tt.want, response); diff != "" {
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

func (fc fakeChartCache) ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error) {
	if charts, ok := fc.charts[objectRefToString(repoRef, clusterRef)]; ok {
		return charts, nil
	}
	return nil, errors.New("no charts found")
}

func objectRefToString(or ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}
