package server

import (
	"context"
	"testing"
	"time"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/helmfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"google.golang.org/protobuf/testing/protocmp"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var defaultClusterState = []runtime.Object{
	&sourcev1beta2.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bitnami-charts",
			Namespace: "demo",
		},
	},
}

func TestListChartsForRepository(t *testing.T) {
	testCases := []struct {
		name         string
		fc           helm.ChartsCache
		clusterState []runtime.Object
		request      *protos.ListChartsForRepositoryRequest
		want         *protos.ListChartsForRepositoryResponse
	}{
		{
			name: "matching cluster and repo",
			request: &protos.ListChartsForRepositoryRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "chart",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					), []helm.Chart{{Name: "redis", Version: "1.0.1", Kind: "chart"}, {Name: "postgres", Version: "1.0.2", Kind: "chart"}})),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{
					{Name: "postgres", Versions: []string{"1.0.2"}},
					{Name: "redis", Versions: []string{"1.0.1"}},
				},
			},
		},
		{
			name: "multiple versions of the same chart",
			request: &protos.ListChartsForRepositoryRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "chart",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					), []helm.Chart{{Name: "redis", Version: "1.0.1", Kind: "chart"}, {Name: "redis", Version: "1.0.2", Kind: "chart"}})),
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
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "chart",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "not-bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					), []helm.Chart{{Name: "redis", Version: "1.0.1"}, {Name: "postgres", Version: "1.0.2"}})),
			want: &protos.ListChartsForRepositoryResponse{
				Charts: []*protos.RepositoryChart{},
			},
		},
		{
			name: "filtering by kind",
			request: &protos.ListChartsForRepositoryRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Kind: "profile",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					),
					[]helm.Chart{
						{Name: "weaveworks-profile", Version: "1.0.1", Kind: "profile"},
						{Name: "postgres", Version: "1.0.2", Kind: "chart"},
					},
				),
			),
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
				chartsCache:     tt.fc,
				clusterState:    tt.clusterState,
				clustersManager: makeTestClustersManager(t, tt.clusterState...),
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
func TestGetValuesForChartFromValuesFetcher(t *testing.T) {
	testCases := []struct {
		name         string
		fc           helm.ChartsCacheReader
		clusterState []runtime.Object
		request      *protos.GetValuesForChartRequest
		want         *protos.GetChartsJobResponse
	}{
		{
			name: "when value exists in cache",
			request: &protos.GetValuesForChartRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Name:    "redis",
				Version: "1.0.1",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					),
					[]helm.Chart{{Name: "redis", Version: "1.0.1"}},
				),
			),
			want: &protos.GetChartsJobResponse{
				Values: "this:\n  is:\n    a: value",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			clusterMngr := makeTestClustersManager(t, tt.clusterState...)
			s := createServer(t, serverOptions{
				chartsCache:     tt.fc,
				chartJobs:       helm.NewJobs(),
				valuesFetcher:   &fakeValuesFetcher{},
				cluster:         "management",
				clustersManager: clusterMngr,
			})

			response, err := s.GetValuesForChart(context.TODO(), tt.request)
			if err != nil {
				t.Fatal(err)
			}

			// Poll GetChartsJob until it's done
			var jobResponse *protos.GetChartsJobResponse
			err = wait.PollImmediate(time.Second, time.Second*5, func() (bool, error) {
				var err error
				jobResponse, err = s.GetChartsJob(context.TODO(), &protos.GetChartsJobRequest{JobId: response.JobId})
				if err != nil {
					return false, err
				}
				return jobResponse.Values != "", nil
			})

			if err != nil {
				t.Log(clusterMngr.Invocations())
				t.Fatalf("error on JobPoll: %s: %v", err, jobResponse)
			}

			if diff := cmp.Diff(tt.want, jobResponse, protocmp.Transform()); diff != "" {
				t.Fatalf("failed to get response:\n%s", diff)
			}

			cachedValue, err := tt.fc.GetChartValues(context.TODO(),
				types.NamespacedName{Name: "management"},
				helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
				helm.Chart{Name: "redis", Version: "1.0.1"})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want.Values, string(cachedValue), protocmp.Transform()); diff != "" {
				t.Fatalf("failed to get response:\n%s", diff)
			}
		})
	}
}

func TestGetValuesForChartCached(t *testing.T) {
	testCases := []struct {
		name         string
		fc           helm.ChartsCacheReader
		clusterState []runtime.Object
		request      *protos.GetValuesForChartRequest
		want         *protos.GetChartsJobResponse
	}{
		{
			name: "when value exists in cache",
			request: &protos.GetValuesForChartRequest{
				Repository: &protos.RepositoryRef{
					Cluster: &protos.ClusterNamespacedName{
						Name: "management",
					},
					Name:      "bitnami-charts",
					Namespace: "demo",
					Kind:      "HelmRepository",
				},
				Name:    "redis",
				Version: "1.0.1",
			},
			clusterState: defaultClusterState,
			fc: helmfakes.NewFakeChartCache(
				helmfakes.WithCharts(
					helmfakes.ClusterRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
					),
					[]helm.Chart{{Name: "redis", Version: "1.0.1"}}),
				helmfakes.WithValues(
					helmfakes.ChartRefToString(
						helm.ObjectReference{Kind: "HelmRepository", Name: "bitnami-charts", Namespace: "demo"},
						types.NamespacedName{Name: "management"},
						helm.Chart{Name: "redis", Version: "1.0.1"}),
					[]byte("this:\n  is:\n    a: value\n"),
				)),
			want: &protos.GetChartsJobResponse{
				// This is the base64 encoded version of "this:\n  is:\n    a: value\n"
				Values: "dGhpczoKICBpczoKICAgIGE6IHZhbHVlCg==",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			s := createServer(t, serverOptions{
				chartsCache:     tt.fc,
				chartJobs:       helm.NewJobs(),
				clustersManager: makeTestClustersManager(t, tt.clusterState...),
			})

			response, err := s.GetValuesForChart(context.TODO(), tt.request)
			if err != nil {
				t.Fatal(err)
			}

			// Poll GetChartsJob until it's done
			var jobResponse *protos.GetChartsJobResponse
			err = wait.PollImmediate(time.Second, time.Second*5, func() (bool, error) {
				var err error
				jobResponse, err = s.GetChartsJob(context.TODO(), &protos.GetChartsJobRequest{JobId: response.JobId})
				if err != nil {
					return false, err
				}
				return jobResponse.Values != "", nil
			})

			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, jobResponse, protocmp.Transform()); diff != "" {
				t.Fatalf("failed to get response:\n%s", diff)
			}
		})
	}
}

type fakeValuesFetcher struct {
}

func (f *fakeValuesFetcher) GetIndexFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error) {
	return nil, nil
}

func (f *fakeValuesFetcher) GetValuesFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, c helm.Chart, useProxy bool) ([]byte, error) {
	return []byte("this:\n  is:\n    a: value"), nil
}
