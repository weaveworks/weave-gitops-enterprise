package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/untar"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type ValuesFetcher interface {
	GetIndexFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error)
	GetValuesFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, c Chart, useProxy bool) ([]byte, error)
}

// use apimachinery wait package to wait for the HelmChart to be ready
func waitForReady(ctx context.Context, cl client.Client, helmChart *sourcev1beta2.HelmChart) error {
	err := util.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		err := cl.Get(ctx, types.NamespacedName{Namespace: helmChart.Namespace, Name: helmChart.Name}, helmChart)
		if err != nil {
			return false, fmt.Errorf("failed to get HelmChart: %w", err)
		}
		return conditions.IsReady(helmChart), nil
	})

	if err != nil {
		return fmt.Errorf("%w: HelmChart %s/%s is not ready: %s", err, helmChart.Namespace, helmChart.Name, conditions.GetMessage(helmChart, meta.ReadyCondition))
	}

	return nil
}

type valuesFetcher struct {
}

func NewValuesFetcher() ValuesFetcher {
	return &valuesFetcher{}
}

func (v *valuesFetcher) GetIndexFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, useProxy bool) (*repo.IndexFile, error) {
	// Get the HelmRepository
	helmRepoObj := &sourcev1beta2.HelmRepository{}
	cl, err := cluster.GetServerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	kcl, err := cluster.GetServerClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	if err := cl.Get(ctx, helmRepo, helmRepoObj); err != nil {
		return nil, fmt.Errorf("failed to get HelmRepository: %w", err)
	}

	// Get the artifact URL
	artifactURL := helmRepoObj.Status.URL
	if artifactURL == "" {
		return nil, fmt.Errorf("no artifact URL found for HelmRepository %s", helmRepo)
	}

	data, err := httpGetFromSourceController(kcl, artifactURL, useProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to get index file: %w", err)
	}

	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return nil, fmt.Errorf("error unmarshaling chart response: %w, url: %v", err, artifactURL)
	}

	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func (v *valuesFetcher) GetValuesFile(ctx context.Context, cluster cluster.Cluster, helmRepo types.NamespacedName, chartRef Chart, useProxy bool) ([]byte, error) {
	// clients
	cl, err := cluster.GetServerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	kcl, err := cluster.GetServerClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Generate a random name for the HelmChart with a prefix of the chart name
	randomChartName := chartRef.Name + "-" + randString(5)

	// Using a typed object.
	helmChart := &sourcev1beta2.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: helmRepo.Namespace,
			Name:      randomChartName,
		},
		Spec: sourcev1beta2.HelmChartSpec{
			Chart: chartRef.Name,
			SourceRef: sourcev1beta2.LocalHelmChartSourceReference{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Name: helmRepo.Name,
			},
			Version: chartRef.Version,
		},
	}

	err = cl.Create(context.Background(), helmChart)
	if err != nil {
		return nil, fmt.Errorf("failed to create HelmChart: %w", err)
	}
	defer func() {
		err := cl.Delete(context.Background(), helmChart)
		if err != nil {
			// FIXME: log this error
			fmt.Println(fmt.Errorf("failed to delete HelmChart: %w", err))
		}
	}()

	err = waitForReady(ctx, cl, helmChart)

	if err != nil {
		return nil, fmt.Errorf("failed to wait for HelmChart to be ready: %w", err)
	}

	data, err := httpGetFromSourceController(kcl, helmChart.Status.URL, useProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to get values file: %w", err)
	}

	return getValuesYamlFromArchive(data, chartRef.Name)
}

func randString(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getValuesYamlFromArchive(data []byte, chartName string) ([]byte, error) {
	dname, err := os.MkdirTemp("", "helm-chart")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(dname)

	if _, err = untar.Untar(bytes.NewBuffer(data), dname); err != nil {
		return nil, fmt.Errorf("failed to untar helm chart: %w", err)
	}

	return os.ReadFile(filepath.Join(dname, chartName, "values.yaml"))
}

func httpGetFromSourceController(kcl kubernetes.Interface, url string, useProxy bool) ([]byte, error) {
	if !useProxy {
		data, err := httpGetFromSourceControllerLocal(url)
		if err != nil {
			return nil, fmt.Errorf("failed to get values file from local cluster: %w", err)
		}
		return data, nil
	}

	parsed, err := ParseArtifactURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artifact URL: %w", err)
	}

	res := kcl.
		CoreV1().
		Services(parsed.Namespace).
		ProxyGet(parsed.Scheme, parsed.Name, parsed.Port, parsed.Path, nil)

	data, err := res.DoRaw(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact from %+v: %w", parsed, err)
	}

	return data, nil
}

func httpGetFromSourceControllerLocal(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %d from %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}
