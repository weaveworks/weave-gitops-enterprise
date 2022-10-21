package helm

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// ProfileAnnotation is the annotation that Helm charts must have to indicate
// that they provide a Profile.
const ProfileAnnotation = "weave.works/profile"

// RepositoryProfilesAnnotation is the annotation that Helm Repositories must
// have to indicate that all charts are to be considered as Profiles.
const RepositoryProfilesAnnotation = "weave.works/profiles"

// LayerAnnotation specifies profile application order.
// Profiles are sorted by layer and those at a higher "layer" are only installed after
// lower layers have successfully installed and started.
const LayerAnnotation = "weave.works/layer"

// ChartReference is a Helm chart reference
type ChartReference struct {
	Chart   string
	Version string
}

// DefaultChartGetter provides default ways to get a chart index.yaml based on
// the URL scheme.
var defaultChartGetters = getter.Providers{
	getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	},
}

// ChartPredicate is used to filter charts coming from a HelmRepository.
type ChartPredicate func(*sourcev1.HelmRepository, *repo.ChartVersion) bool

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(hr *sourcev1.HelmRepository, v *repo.ChartVersion) bool {
	return hasAnnotation(v.Metadata.Annotations, ProfileAnnotation) ||
		hasAnnotation(hr.ObjectMeta.Annotations, RepositoryProfilesAnnotation)
}

func ReverseSemVerSort(versions []string) ([]string, error) {
	vs := make([]*semver.Version, len(versions))

	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r, err)
		}

		vs[i] = v
	}

	sort.Sort(sort.Reverse(semver.Collection(vs)))

	result := make([]string, len(versions))
	for i := range vs {
		result[i] = vs[i].String()
	}

	return result, nil
}

func credsForRepository(ctx context.Context, kc client.Client, hr *sourcev1.HelmRepository) (string, string, error) {
	var secret corev1.Secret
	if err := kc.Get(ctx, types.NamespacedName{Name: hr.Spec.SecretRef.Name, Namespace: hr.Namespace}, &secret); err != nil {
		return "", "", fmt.Errorf("repository authentication: %w", err)
	}

	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func fetchIndexFile(chartURL string) (*repo.IndexFile, error) {
	if hostname := os.Getenv("SOURCE_CONTROLLER_LOCALHOST"); hostname != "" {
		u, err := url.Parse(chartURL)
		if err != nil {
			return nil, err
		}

		u.Host = hostname
		chartURL = u.String()
	}

	u, err := url.Parse(chartURL)

	if err != nil {
		return nil, fmt.Errorf("error parsing URL %q: %w", chartURL, err)
	}

	c, err := defaultChartGetters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("no provider for scheme %q: %w", u.Scheme, err)
	}

	res, err := c.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching index file: %w", err)
	}

	b, err := io.ReadAll(res)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return nil, fmt.Errorf("error unmarshaling chart response: %w", err)
	}

	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func getLayer(annotations map[string]string) string {
	return annotations[LayerAnnotation]
}

func hasAnnotation(cm map[string]string, name string) bool {
	for k := range cm {
		if k == name {
			return true
		}
	}

	return false
}
