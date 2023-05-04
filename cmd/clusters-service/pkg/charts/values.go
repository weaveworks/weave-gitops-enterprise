package charts

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
)

const (
	// LayerAnnotation is the annotation that Helm charts can have to indicate which
	// layer they should be in, the HelmRelease DependsOn is calculated from this.
	LayerAnnotation = "weave.works/layer"

	// LayerLabel is applied to created HelmReleases which makes it possible to
	// query for HelmReleases that are applied in a layer.
	LayerLabel = "weave.works/applied-layer"
)

// ChartReference is a Helm chart reference, the SourceRef is a Flux
// SourceReference for the Helm chart.
type ChartReference struct {
	Chart     string
	Version   string
	SourceRef helmv2.CrossNamespaceObjectReference
}

// HelmChartClient implements ChartClient using the Helm library packages.
type HelmChartClient struct {
	client.Client
	Namespace  string
	Repository *sourcev1beta2.HelmRepository
	CacheDir   string
}

// WithCacheDir configures the HelmChartClient to use the directory for the Helm
// repository cache.
func WithCacheDir(dir string) func(*HelmChartClient) {
	return func(h *HelmChartClient) {
		h.CacheDir = dir
	}
}

// NewHelmChartClient creates and returns a new HelmChartClient.
func NewHelmChartClient(kc client.Client, ns string, hr *sourcev1beta2.HelmRepository, opts ...func(*HelmChartClient)) *HelmChartClient {
	h := &HelmChartClient{
		Client:     kc,
		Namespace:  ns,
		Repository: hr,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// UpdateCache must be called before any calls to fetch charts.
//
// If the cache dir is empty, then it will use the default Helm cache directory
// for the repo cache.
func (h *HelmChartClient) UpdateCache(ctx context.Context) error {
	entry, err := h.entryForRepository(ctx)
	if err != nil {
		return err
	}
	r, err := repo.NewChartRepository(entry, DefaultChartGetters)
	if err != nil {
		return err
	}
	r.CachePath = h.CacheDir
	_, err = r.DownloadIndexFile()
	return err
}

// ValuesForChart fetches the values.yaml file for a ChartReference.
func (h HelmChartClient) ValuesForChart(ctx context.Context, c *ChartReference) (map[string]interface{}, error) {
	chart, err := h.loadChart(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("loading chart values: %w", err)
	}
	return chart.Values, nil
}

func (h HelmChartClient) loadChart(ctx context.Context, c *ChartReference) (*chart.Chart, error) {
	o, err := h.chartPathOptionsFromRepository(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to configure client: %w", err)
	}

	chartLocation, err := o.LocateChart(c.Chart, h.envSettings())
	if err != nil {
		return nil, fmt.Errorf("locating chart %q: %w", c.Chart, err)
	}
	chart, err := loader.Load(chartLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %q: %w", c.Chart, err)
	}
	return chart, nil
}

// FileFromChart fetches the named file from a chart.
func (h HelmChartClient) FileFromChart(ctx context.Context, c *ChartReference, filename string) ([]byte, error) {
	chart, err := h.loadChart(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("loading %s from chart: %w", filename, err)
	}
	for _, v := range chart.Raw {
		if v.Name == filename {
			return v.Data, nil
		}
	}
	return nil, fmt.Errorf("failed to find file: %s", filename)
}

func credsForRepository(ctx context.Context, kc client.Client, ns string, hr *sourcev1beta2.HelmRepository) (string, string, error) {
	var secret corev1.Secret
	if err := kc.Get(ctx, types.NamespacedName{Name: hr.Spec.SecretRef.Name, Namespace: ns}, &secret); err != nil {
		return "", "", fmt.Errorf("repository authentication: %w", err)
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func (h HelmChartClient) chartPathOptionsFromRepository(ctx context.Context, c *ChartReference) (*action.ChartPathOptions, error) {
	// TODO: This should probably use Verify: true
	co := &action.ChartPathOptions{
		RepoURL: h.Repository.Spec.URL,
		Version: c.Version,
	}

	if h.Repository.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, h.Namespace, h.Repository)
		if err != nil {
			return nil, err
		}
		co.Username = username
		co.Password = password
	}
	return co, nil
}

func (h HelmChartClient) entryForRepository(ctx context.Context) (*repo.Entry, error) {
	entry := &repo.Entry{
		Name: h.Repository.GetName() + "-" + h.Repository.GetNamespace(),
		URL:  h.Repository.Spec.URL,
	}
	if h.Repository.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, h.Namespace, h.Repository)
		if err != nil {
			return nil, err
		}
		entry.Username = username
		entry.Password = password
	}
	return entry, nil
}

func (h HelmChartClient) envSettings() *cli.EnvSettings {
	conf := cli.New()
	conf.Debug = true
	if h.CacheDir != "" {
		conf.RepositoryCache = h.CacheDir
		conf.RepositoryConfig = path.Join(h.CacheDir, "/repository.yaml")
	}
	return conf
}

// MakeHelmReleasesInLayers accepts a set of ChartInstall requests and
// returns a set of HelmReleases that are configured with appropriate
// dependencies.
//
// If the Charts are annotated with a layer, the charts will be installed in the
// layer order.
//
// For charts without a layer, these will be configured to depend on the highest
// layer.
func MakeHelmReleasesInLayers(namespace string, installs []ChartInstall) ([]*helmv2.HelmRelease, error) {
	layerInstalls := map[string][]ChartInstall{}
	for _, v := range installs {
		current, ok := layerInstalls[v.Layer]
		if !ok {
			current = []ChartInstall{}
		}
		current = append(current, v)
		layerInstalls[v.Layer] = current
	}

	var layerNames []string
	for k := range layerInstalls {
		layerNames = append(layerNames, k)
	}

	layerDependencies := pairLayers(layerNames)
	var releases []*helmv2.HelmRelease
	for _, layer := range layerDependencies {
		for _, install := range layerInstalls[layer.name] {
			jsonValues, err := json.Marshal(install.Values)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal values for chart %s: %w", install.Ref.Chart, err)
			}
			hr := helmv2.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      install.Ref.Chart,
					Namespace: namespace,
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: helmv2.GroupVersion.Identifier(),
					Kind:       helmv2.HelmReleaseKind,
				},
				Spec: helmv2.HelmReleaseSpec{
					Chart: helmv2.HelmChartTemplate{
						Spec: helmv2.HelmChartTemplateSpec{
							Chart:   install.Ref.Chart,
							Version: install.Ref.Version,
							SourceRef: helmv2.CrossNamespaceObjectReference{
								APIVersion: sourcev1beta2.GroupVersion.Identifier(),
								Kind:       sourcev1beta2.HelmRepositoryKind,
								Name:       install.Ref.SourceRef.Name,
								Namespace:  install.Ref.SourceRef.Namespace,
							},
						},
					},
					Interval: metav1.Duration{Duration: time.Minute},
					Values:   &apiextensionsv1.JSON{Raw: jsonValues},
					Install: &helmv2.Install{
						CRDs: helmv2.CreateReplace,
					},
					Upgrade: &helmv2.Upgrade{
						CRDs: helmv2.CreateReplace,
					},
				},
			}
			if install.Namespace != "" {
				hr.Spec.TargetNamespace = install.Namespace
				hr.Spec.Install.CreateNamespace = true
			}
			if layer.dependsOn != "" {
				for _, v := range layerInstalls[layer.dependsOn] {
					hr.Spec.DependsOn = append(hr.Spec.DependsOn,
						meta.NamespacedObjectReference{
							Name: v.Ref.Chart,
						})
				}
			}
			if layer.name != "" {
				hr.Labels = map[string]string{
					LayerLabel: layer.name,
				}
			}
			if install.ProfileTemplate != "" {
				err := yaml.Unmarshal([]byte(install.ProfileTemplate), &hr)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal spec for chart %s: %w", install.Ref.Chart, err)
				}
			}

			releases = append(releases, &hr)
		}
	}

	sort.Slice(releases, func(i, j int) bool { return releases[i].GetName() < releases[j].GetName() })
	return releases, nil
}

type layerDependency struct {
	name      string
	dependsOn string
}

// iterate over a slice returning slice where element 1 will be configured to
// depend on layer 0.
//
// The sorting is determined lexicographically.
func pairLayers(names []string) []layerDependency {
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	deps := []layerDependency{}
	for i := range names {
		if i < len(names)-1 {
			deps = append(deps, layerDependency{name: names[i], dependsOn: names[i+1]})
			continue
		}
		dep := layerDependency{name: names[i]}
		if names[i] == "" && len(names) > 0 {
			dep.dependsOn = names[0]
		}
		deps = append(deps, dep)
	}
	return deps
}

// ChartInstall configures the installation of a specific chart into a
// cluster.
type ChartInstall struct {
	Ref       ChartReference
	Layer     string
	Values    map[string]interface{}
	Namespace string
	// Spec is a RawExtension.Raw field that contains the raw JSON data
	ProfileTemplate string
}
