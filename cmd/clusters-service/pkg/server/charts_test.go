package server

import (
	"context"
	"encoding/base64"
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
				cachedCharts(
					clusterRefToString(
						ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
					), []Chart{{Name: "redis", Version: "1.0.1"}, {Name: "postgres", Version: "1.0.2"}})),
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
				cachedCharts(
					clusterRefToString(
						ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
					), []Chart{{Name: "redis", Version: "1.0.1"}, {Name: "redis", Version: "1.0.2"}})),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{
					{Name: "redis", Versions: []string{"1.0.2", "1.0.1"}},
				},
			},
		},
		{
			name: "no charts for cluster / repository",
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
				cachedCharts(
					clusterRefToString(
						ObjectReference{Kind: "HelmRepository", Name: "not-bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
					), []Chart{{Name: "redis", Version: "1.0.1"}, {Name: "postgres", Version: "1.0.2"}})),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{},
			},
		},
		{
			name: "filtering by kind",
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
				Kind: "profile",
			},
			fc: newFakeChartCache(
				cachedCharts(
					clusterRefToString(
						ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
					), []Chart{{Name: "weaveworks-profile", Version: "1.0.1"}, {Name: "postgres", Version: "1.0.2"}})),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{
					{Name: "weaveworks-profile", Versions: []string{"1.0.1"}},
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

func TestGetValuesForChart(t *testing.T) {
	testCases := []struct {
		name    string
		fc      *fakeChartCache
		request *protos.GetValuesForChartRequest
		want    *protos.GetValuesForChartResponse
	}{
		{
			name: "when value exists in cache",
			request: &protos.GetValuesForChartRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Namespace: "clusters",
						Name:      "demo-cluster",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Name:    "redis",
				Version: "1.0.1",
			},
			fc: newFakeChartCache(
				cachedValues(
					chartRefToString(
						ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "demo-cluster", Namespace: "clusters"},
						Chart{Name: "redis", Version: "1.0.1"}),
					[]byte("this:\n  is:\n    a: value\n"),
				)),
			want: &protos.GetValuesForChartResponse{
				// This is the base64 encoded version of "this:\n  is:\n    a: value\n"
				Values: "dGhpczoKICBpczoKICAgIGE6IHZhbHVlCg==",
			},
		},
		// {
		// 	name: "when the value does not exist in the cache",
		// },
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			s := createServer(t, serverOptions{
				chartsCache: tt.fc,
			})

			response, err := s.GetValuesForChart(context.TODO(), tt.request)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, response, protocmp.Transform()); diff != "" {
				t.Fatalf("failed to get response:\n%s", diff)
			}
		})
	}
}

func cachedCharts(key string, charts []Chart) func(*fakeChartCache) {
	return func(fc *fakeChartCache) {
		fc.charts[key] = charts
	}
}

func cachedValues(key string, chart []byte) func(*fakeChartCache) {
	return func(fc *fakeChartCache) {
		fc.chartValues[key] = []byte(base64.StdEncoding.EncodeToString(chart))
	}
}

func newFakeChartCache(opts ...func(*fakeChartCache)) *fakeChartCache {
	fc := &fakeChartCache{
		charts:      make(map[string][]Chart),
		chartValues: make(map[string][]byte),
	}
	for _, o := range opts {
		o(fc)
	}

	return fc
}

type fakeChartCache struct {
	charts      map[string][]Chart
	chartValues map[string][]byte
}

func (fc fakeChartCache) ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error) {
	if charts, ok := fc.charts[clusterRefToString(repoRef, clusterRef)]; ok {
		return charts, nil
	}
	return nil, errors.New("no charts found")
}
func (fc fakeChartCache) GetChartValues(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName, chart Chart) ([]byte, error) {
	if values, ok := fc.chartValues[chartRefToString(repoRef, clusterRef, chart)]; ok {
		return values, nil
	}
	return nil, errors.New("values not found")
}

func clusterRefToString(or ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}

func chartRefToString(or ObjectReference, cr types.NamespacedName, c Chart) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace, c.Name, c.Version)
}
