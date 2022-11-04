package controller

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/helmfakes"
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
		makeTestHelmRepo(func(hr *sourcev1.HelmRepository) {
			newTime := metav1.NewTime(time.Now())
			hr.ObjectMeta.DeletionTimestamp = &newTime
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
	deletedHelmRepo := makeTestHelmRepo(func(hr *sourcev1.HelmRepository) {
		newTime := metav1.NewTime(time.Now())
		hr.ObjectMeta.DeletionTimestamp = &newTime
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

func setupReconcileAndFakes(helmRepo client.Object, fakeFetcher *fakeValuesFetcher, fakeCache helm.ChartsCacherWriter) *HelmWatcherReconciler {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

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

// makeTestHelmRepo creates a HelmRepository object and accepts a list of options to modify it.
func makeTestHelmRepo(opts ...func(*sourcev1.HelmRepository)) *sourcev1.HelmRepository {
	repo := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: "source.toolkit.fluxcd.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
				Checksum: "checksum",
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

func (f *fakeValuesFetcher) GetIndexFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error) {
	return f.indexFile, f.getIndexFileError
}

func (f *fakeValuesFetcher) GetValuesFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, c helm.Chart, useProxy bool) ([]byte, error) {
	return nil, nil
}
