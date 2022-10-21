package controller

import (
	"context"
	"fmt"
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
	repo1 = &sourcev1.HelmRepository{
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
)

func TestReconcile(t *testing.T) {
	reconciler, cache := setupReconcileAndFakes(repo1, repo1Index)

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
	cacheData := cache.charts[clusterRefToString(helmRepo, clusterRef)]

	expectedData := []helm.Chart{
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

	assert.Equal(t, expectedData, cacheData)
}

func TestReconcileDelete(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	key := clusterRefToString(
		helm.ObjectReference{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
		clusterRef,
	)
	reconciler, fakeCache := setupReconcileAndFakes(
		repo, nil, cachedCharts(
			key,
			[]helm.Chart{
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
			},
		),
	)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	// cache should be empty after delete
	assert.Equal(t, []helm.Chart(nil), fakeCache.charts[key])
}

/*

func TestReconcileDeletingTheCacheFails(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, fakeCache := setupReconcileAndFakes(repo, nil)

	fakeCache.DeleteReturns(errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestReconcileGetChartFails(t *testing.T) {
	reconciler, _, fakeRepoManager, _ := setupReconcileAndFakes(repo1)
	fakeRepoManager.ListChartsReturns(nil, errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}
func TestReconcileGetValuesFileFailsItWillContinue(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes(repo1)
	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	expectedData := cache.Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values:   map[string]map[string][]byte{},
	}
	_, namespace, name, cacheData := fakeCache.PutArgsForCall(0)
	assert.Equal(t, "test-namespace", namespace)
	assert.Equal(t, "test-name", name)
	assert.Equal(t, expectedData, cacheData)
}

func TestReconcileIgnoreReposWithoutArtifact(t *testing.T) {
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
	}
	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes(repo)

	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})

	assert.NoError(t, err)
	assert.Zero(t, fakeRepoManager.ListChartsCallCount())
	assert.Zero(t, fakeRepoManager.GetValuesFileCallCount())
	assert.Zero(t, fakeCache.PutCallCount())
}

func TestReconcileUpdateReturnsError(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes(repo1)
	fakeCache.PutReturns(errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestNotifyForGreaterVersion(t *testing.T) {
	reconciler, fakeCache, _, fakeEventRecorder := setupReconcileAndFakes(repo1)
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.0"}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	_, meta, severity, _, message, args := fakeEventRecorder.AnnotatedEventfArgsForCall(0)

	assert.Equal(t, map[string]string{"revision": "revision"}, meta)
	assert.Equal(t, "info", severity)
	assert.Equal(t, "New version available for profile %s with version %s", message)
	assert.Equal(t, []interface{}{profile1.Name, "0.0.2"}, args)
}

func TestDoNotNotifyForLesserOrEqualVersion(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, fakeEventRecorder := setupReconcileAndFakes(repo1)
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.2"}, nil)
	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
	assert.Zero(t, fakeEventRecorder.AnnotatedEventfCallCount())
}

func TestNotifyForGreaterVersionListAvailableVersionsReturnsErrorIsSkipped(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns(nil, errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

func TestNotifyForGreaterVersionListAvailableVersionsReturnsHigherVersion(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, fakeEventRecorder := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.1"}, nil)
	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
	assert.Zero(t, fakeEventRecorder.AnnotatedEventfCallCount())
}

func TestNotifyForGreaterVersionEventSenderFailureIsIgnored(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.0"}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

type mockClient struct {
	client.Client
	getErr    error
	updateErr error
	patchErr  error
	obj       *sourcev1.HelmRepository
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, object client.Object, opts ...client.GetOption) error {
	if m.obj != nil {
		if v, ok := object.(*sourcev1.HelmRepository); ok {
			*v = *m.obj
		}
	}

	return m.getErr
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return m.updateErr
}

func (m *mockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return m.patchErr
}

func TestReconcileKubernetesGetFails(t *testing.T) {
	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	reconciler := &HelmWatcherReconciler{
		Client:      &mockClient{getErr: errors.New("nope")},
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
	assert.Zero(t, fakeRepoManager.ListChartsCallCount())
	assert.Zero(t, fakeRepoManager.GetValuesFileCallCount())
	assert.Zero(t, fakeCache.PutCallCount())
}

func TestReconcileUpdateFailsDuringDelete(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, _, _, _ := setupReconcileAndFakes()
	reconciler.Client = &mockClient{
		obj:       repo,
		updateErr: errors.New("nope"),
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestReconcilePatchFails(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, _, _, _ := setupReconcileAndFakes()
	reconciler.Client = &mockClient{
		obj:      repo,
		patchErr: errors.New("nope"),
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}
*/

func setupReconcileAndFakes(helmRepo client.Object, indexFile *repo.IndexFile, opts ...func(*fakeChartCache)) (*HelmWatcherReconciler, fakeChartCache) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme)
	if helmRepo != nil {
		fakeClient = fakeClient.WithObjects(helmRepo)
	}
	fakeCache := newFakeChartCache(opts...)

	return &HelmWatcherReconciler{
		ClusterRef:    clusterRef,
		Client:        fakeClient.Build(),
		Cache:         fakeCache,
		ValuesFetcher: &fakeValuesFetcher{indexFile: indexFile},
	}, *fakeCache
}

func cachedCharts(key string, charts []helm.Chart) func(*fakeChartCache) {
	return func(fc *fakeChartCache) {
		fc.charts[key] = charts
	}
}

func newFakeChartCache(opts ...func(*fakeChartCache)) *fakeChartCache {
	fc := &fakeChartCache{
		charts: make(map[string][]helm.Chart),
	}
	for _, o := range opts {
		o(fc)
	}

	return fc
}

type fakeChartCache struct {
	charts map[string][]helm.Chart
}

func (fc fakeChartCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	k := clusterRefToString(repoRef, clusterRef)
	fmt.Printf("Adding chart %s to cache with key %s\n", name, k)
	fc.charts[k] = append(
		fc.charts[k],
		helm.Chart{
			Name:    name,
			Version: version,
			Layer:   layer,
			Kind:    kind,
		},
	)
	return nil
}

func (fc fakeChartCache) Delete(ctx context.Context, repoRef helm.ObjectReference, clusterRef types.NamespacedName) error {
	k := clusterRefToString(repoRef, clusterRef)
	delete(fc.charts, k)
	return nil
}

func clusterRefToString(or helm.ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}

func chartRefToString(or helm.ObjectReference, cr types.NamespacedName, c helm.Chart) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace, c.Name, c.Version)
}

type fakeValuesFetcher struct {
	indexFile *repo.IndexFile
}

func (f *fakeValuesFetcher) GetIndexFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error) {
	return f.indexFile, nil
}

func (f *fakeValuesFetcher) GetValuesFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, c helm.Chart, useProxy bool) ([]byte, error) {
	return nil, nil
}
