package controller

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/helmfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

var (
	clusterRef = types.NamespacedName{Name: "test-cluster", Namespace: "test-namespace"}
	repo1Index = &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"test-profiles-1": {
				{
					Metadata: &chart.Metadata{
						Name:    "test-profiles-1",
						Version: "0.0.1",
					},
				},
				{
					Metadata: &chart.Metadata{
						Name:    "test-profiles-1",
						Version: "0.0.2",
					},
				},
			},
			"test-profiles-2": {
				{
					Metadata: &chart.Metadata{
						Name:    "test-profiles-2",
						Version: "0.0.4",
					},
				},
			},
		},
	}
	repo1Charts = []helm.Chart{
		{
			Name:    "test-profiles-1",
			Version: "0.0.1",
			Kind:    "chart",
		},
		{
			Name:    "test-profiles-1",
			Version: "0.0.2",
			Kind:    "chart",
		},
		{
			Name:    "test-profiles-2",
			Version: "0.0.4",
			Kind:    "chart",
		},
	}
)

func TestReconcile(t *testing.T) {
	fakeCache := helmfakes.NewFakeChartCache()
	reconciler := setupReconcileAndFakes(
		makeTestHelmRepo(),
		&fakeValuesFetcher{repo1Index, nil},
		fakeCache,
	)
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	helmRepo := helm.ObjectReference{
		Namespace: "test-namespace",
		Name:      "test-name",
	}
	cacheData := fakeCache.Charts[helmfakes.ClusterRefToString(helmRepo, clusterRef)]
	// sort the cache data by name and version
	sort.Slice(cacheData, func(i, j int) bool {
		if cacheData[i].Name == cacheData[j].Name {
			return cacheData[i].Version < cacheData[j].Version
		}
		return cacheData[i].Name < cacheData[j].Name
	})

	expectedData := repo1Charts
	assert.Equal(t, expectedData, cacheData)
}

func TestReconcileWithMissingHelmRepository(t *testing.T) {
	reconciler := setupReconcileAndFakes(nil, nil, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

func TestReconcileDelete(t *testing.T) {
	key := helmfakes.ClusterRefToString(
		helm.ObjectReference{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
		clusterRef,
	)
	fakeCache := helmfakes.NewFakeChartCache(helmfakes.WithCharts(key, repo1Charts))
	reconciler := setupReconcileAndFakes(
		makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
			newTime := metav1.NewTime(time.Now())
			hr.ObjectMeta.DeletionTimestamp = &newTime
			hr.Finalizers = []string{"helm.weave.works/finalizer"}
		}),
		&fakeValuesFetcher{nil, nil},
		fakeCache,
	)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	// cache should be empty after delete
	assert.Equal(t, []helm.Chart(nil), fakeCache.Charts[key])
}

func TestReconcileDeletingTheCacheFails(t *testing.T) {
	deletedHelmRepo := makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
		newTime := metav1.NewTime(time.Now())
		hr.ObjectMeta.DeletionTimestamp = &newTime
		hr.Finalizers = []string{"helm.weave.works/finalizer"}
	})
	fakeErroringCache := helmfakes.NewFakeChartCache(func(fc *helmfakes.FakeChartCache) {
		fc.DeleteError = errors.New("nope")
	})
	reconciler := setupReconcileAndFakes(deletedHelmRepo, &fakeValuesFetcher{nil, nil}, fakeErroringCache)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestReconcileGetChartFails(t *testing.T) {
	helmRepo := makeTestHelmRepo()
	erroringValuesFetcher := fakeValuesFetcher{nil, errors.New("nope")}
	reconciler := setupReconcileAndFakes(helmRepo, &erroringValuesFetcher, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestLoadIndex(t *testing.T) {
	ctx := context.TODO()
	helmRepo := makeTestHelmRepo()

	index := &repo.IndexFile{
		APIVersion: "v1",
		Generated:  time.Now(),
		Entries: map[string]repo.ChartVersions{
			"chart1": {
				{
					Metadata: &chart.Metadata{
						Name:    "chart1",
						Version: "1.0.0",
					},
				},
			},
			"profile1": {
				{
					Metadata: &chart.Metadata{
						Name:    "profile1",
						Version: "2.0.0",
						Annotations: map[string]string{
							"weave.works/profile": "true",
							"weave.works/layer":   "layer-0",
						},
					},
				},
			},
		},
	}

	fakeCache := helmfakes.NewFakeChartCache()
	LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

	// Should detect the kind of the chart
	charts, err := fakeCache.ListChartsByRepositoryAndCluster(
		context.Background(),
		clusterRef,
		helm.ObjectReference{
			Name:      helmRepo.Name,
			Namespace: helmRepo.Namespace,
		},
		"chart",
	)
	expected := []helm.Chart{
		{
			Name:    "chart1",
			Version: "1.0.0",
			Kind:    "chart",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, charts)

	// and see we can get the profiles too
	profiles, err := fakeCache.ListChartsByRepositoryAndCluster(
		context.Background(),
		clusterRef,
		helm.ObjectReference{
			Name:      helmRepo.Name,
			Namespace: helmRepo.Namespace,
		},
		"profile",
	)
	expected = []helm.Chart{
		{
			Name:    "profile1",
			Version: "2.0.0",
			Kind:    "profile",
			Layer:   "layer-0",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, profiles)
}

func setupReconcileAndFakes(helmRepo client.Object, fakeFetcher *fakeValuesFetcher, fakeCache helm.ChartsCacheWriter) *HelmWatcherReconciler {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1beta2.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme)
	if helmRepo != nil {
		fakeClient = fakeClient.WithObjects(helmRepo)
	}

	return &HelmWatcherReconciler{
		ClusterRef:    clusterRef,
		Client:        fakeClient.Build(),
		Cache:         fakeCache,
		ValuesFetcher: fakeFetcher,
	}
}

func TestLoadIndex_filtering_charts(t *testing.T) {
	index := &repo.IndexFile{
		APIVersion: "v1",
		Generated:  time.Now(),
		Entries: map[string]repo.ChartVersions{
			"chart1": {
				{
					Metadata: &chart.Metadata{
						Name:    "chart1",
						Version: "1.0.0",
					},
				},
			},
			"chart2": {
				{
					Metadata: &chart.Metadata{
						Name:    "chart2",
						Version: "2.0.1-rc1",
					},
				},
			},

			"profile1": {
				{
					Metadata: &chart.Metadata{
						Name:    "profile1",
						Version: "2.0.0",
						Annotations: map[string]string{
							"weave.works/profile": "true",
							"weave.works/layer":   "layer-0",
						},
					},
				},
			},
		},
	}
	ctx := context.TODO()

	t.Run("simple version filtering", func(t *testing.T) {
		fakeCache := helmfakes.NewFakeChartCache()
		helmRepo := makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
			hr.ObjectMeta.Annotations = map[string]string{
				HelmVersionFilterAnnotation: ">= 2.0.0",
			}
		})

		LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

		// This should filter out the charts
		charts, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"chart",
		)

		assert.NoError(t, err)
		assert.Empty(t, charts)

		// But the profile (being version 2.0.0 should be retained
		profiles, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"profile",
		)
		expected := []helm.Chart{
			{
				Name:    "profile1",
				Version: "2.0.0",
				Kind:    "profile",
				Layer:   "layer-0",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, profiles)
	})

	t.Run("rc version filtering", func(t *testing.T) {
		fakeCache := helmfakes.NewFakeChartCache()
		helmRepo := makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
			hr.ObjectMeta.Annotations = map[string]string{
				HelmVersionFilterAnnotation: ">= 2.0.0-0",
			}
		})

		LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

		// This should retain the rc chart
		charts, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"chart",
		)
		assert.NoError(t, err)

		expected := []helm.Chart{
			{
				Name:    "chart2",
				Version: "2.0.1-rc1",
				Kind:    "chart",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, charts)

		// And the profile chart 2.0.0
		profiles, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"profile",
		)
		expected = []helm.Chart{
			{
				Name:    "profile1",
				Version: "2.0.0",
				Kind:    "profile",
				Layer:   "layer-0",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, profiles)
	})

	t.Run("bad semantic version ignores filtering and includes everything", func(t *testing.T) {
		fakeCache := helmfakes.NewFakeChartCache()
		helmRepo := makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
			hr.ObjectMeta.Annotations = map[string]string{
				HelmVersionFilterAnnotation: "BAR >= 1.2.3",
			}
		})

		LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

		// This should retain the rc chart
		charts, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"chart",
		)
		assert.NoError(t, err)

		expected := []helm.Chart{
			{
				Name:    "chart1",
				Version: "1.0.0",
				Kind:    "chart",
			},
			{
				Name:    "chart2",
				Version: "2.0.1-rc1",
				Kind:    "chart",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, charts)

		// And the profile chart 2.0.0
		profiles, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"profile",
		)
		expected = []helm.Chart{
			{
				Name:    "profile1",
				Version: "2.0.0",
				Kind:    "profile",
				Layer:   "layer-0",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, profiles)
	})

	t.Run("updating the cache removes items that have been filtered out", func(t *testing.T) {
		fakeCache := helmfakes.NewFakeChartCache()
		helmRepo := makeTestHelmRepo()
		LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

		// This should include all charts.
		charts, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"chart",
		)

		assert.NoError(t, err)
		expected := []helm.Chart{
			{
				Name:    "chart1",
				Version: "1.0.0",
				Kind:    "chart",
			},
			{
				Name:    "chart2",
				Version: "2.0.1-rc1",
				Kind:    "chart",
			},
		}
		assert.Equal(t, expected, charts)

		helmRepo = makeTestHelmRepo(func(hr *sourcev1beta2.HelmRepository) {
			hr.ObjectMeta.Annotations = map[string]string{
				HelmVersionFilterAnnotation: ">= 2.0.0",
			}
		})

		LoadIndex(ctx, index, fakeCache, clusterRef, helmRepo, logr.Discard())

		// This should now have no charts.
		charts, err = fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"chart",
		)

		assert.NoError(t, err)
		assert.Empty(t, charts)

		// But the profile (being version 2.0.0 should be retained
		profiles, err := fakeCache.ListChartsByRepositoryAndCluster(
			context.Background(),
			clusterRef,
			helm.ObjectReference{
				Name:      helmRepo.Name,
				Namespace: helmRepo.Namespace,
			},
			"profile",
		)
		expected = []helm.Chart{
			{
				Name:    "profile1",
				Version: "2.0.0",
				Kind:    "profile",
				Layer:   "layer-0",
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, profiles)
	})

}

// makeTestHelmRepo creates a HelmRepository object and accepts a list of options to modify it.
func makeTestHelmRepo(opts ...func(*sourcev1beta2.HelmRepository)) *sourcev1beta2.HelmRepository {
	repo := &sourcev1beta2.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta2.HelmRepositoryKind,
			APIVersion: "source.toolkit.fluxcd.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1beta2.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
			},
		},
	}

	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

type fakeValuesFetcher struct {
	indexFile         *repo.IndexFile
	getIndexFileError error
}

func (f *fakeValuesFetcher) GetIndexFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error) {
	return f.indexFile, f.getIndexFileError
}

func (f *fakeValuesFetcher) GetValuesFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, c helm.Chart, useProxy bool) ([]byte, error) {
	return nil, nil
}
