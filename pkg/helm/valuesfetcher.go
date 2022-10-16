package helm

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/untar"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type ValuesFetcher interface {
	GetIndexFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName) (*repo.IndexFile, error)
	GetValuesFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, c Chart) ([]byte, error)
}

type MakeClientsFn func(config *rest.Config) (client.Client, kubernetes.Interface, error)

func MakeClients(config *rest.Config) (client.Client, kubernetes.Interface, error) {
	cl, err := getClient(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	kcl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return cl, kcl, nil
}

func getClient(config *rest.Config) (client.Client, error) {
	schema := runtime.NewScheme()
	if err := sourcev1.AddToScheme(schema); err != nil {
		return nil, fmt.Errorf("failed to add sourcev1 to scheme: %w", err)
	}

	return client.New(config, client.Options{Scheme: schema})
}

func waitForReady(ctx context.Context, cl client.Client, helmChart *sourcev1.HelmChart) error {
	// use apimachinery wait package to wait for the HelmChart to be ready
	// then get the values file from the HelmChart
	return util.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		err := cl.Get(ctx, types.NamespacedName{Namespace: helmChart.Namespace, Name: helmChart.Name}, helmChart)
		if err != nil {
			return false, fmt.Errorf("failed to get HelmChart: %w", err)
		}
		return conditions.IsReady(helmChart), nil
	})
}

type valuesFetcher struct {
	makeClients MakeClientsFn
}

func NewValuesFetcher() ValuesFetcher {
	return &valuesFetcher{
		makeClients: MakeClients,
	}
}

func (v *valuesFetcher) GetIndexFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName) (*repo.IndexFile, error) {
	// Get the HelmRepository
	helmRepoObj := &sourcev1.HelmRepository{}
	cl, kcl, err := v.makeClients(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients: %w", err)
	}

	if err := cl.Get(ctx, helmRepo, helmRepoObj); err != nil {
		return nil, fmt.Errorf("failed to get HelmRepository: %w", err)
	}

	// Get the artifact URL
	artifactURL := helmRepoObj.Status.URL
	if artifactURL == "" {
		return nil, fmt.Errorf("no artifact URL found for HelmRepository %s", helmRepo)
	}

	data, err := httpGetFromSourceController(kcl, artifactURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get index file: %w", err)
	}

	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return nil, fmt.Errorf("error unmarshaling chart response: %w", err)
	}

	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func (v *valuesFetcher) GetValuesFile(ctx context.Context, config *rest.Config, helmRepo types.NamespacedName, chartRef Chart) ([]byte, error) {
	// clients
	cl, kcl, err := v.makeClients(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients: %w", err)
	}

	// Generate a random name for the HelmChart with a prefix of the chart name
	randomChartName := chartRef.Name + "-" + randString(5)

	// Using a typed object.
	helmChart := &sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: helmRepo.Namespace,
			Name:      randomChartName,
		},
		Spec: sourcev1.HelmChartSpec{
			Chart: chartRef.Name,
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Kind: sourcev1.HelmRepositoryKind,
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

	data, err := httpGetFromSourceController(kcl, helmChart.Status.URL)
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

func httpGetFromSourceController(kcl kubernetes.Interface, url string) ([]byte, error) {
	parsed, err := ParseArtifactURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artifact URL: %w", err)
	}

	fmt.Printf("Getting artifact from %s\n", kcl)
	res := kcl.
		CoreV1().
		Services(parsed.Namespace).
		ProxyGet(parsed.Scheme, parsed.Name, parsed.Port, parsed.Path, nil)

	data, err := res.DoRaw(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	return data, nil
}
