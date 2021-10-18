package charts

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"helm.sh/helm/pkg/repo"
	"helm.sh/helm/v3/pkg/getter"
	hapichart "k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/yaml"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
)

// ProfileAnnotation is the annotation that Helm charts must have to indicate
// that they provide a Profile.
const ProfileAnnotation = "weave.works/profile"

// DefaultChartGetter provides default ways to get a chart index.yaml based on
// the URL scheme.
var DefaultChartGetters = getter.Providers{
	getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	},
}

type chartPredicate func(*repo.ChartVersion) bool

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(v *repo.ChartVersion) bool {
	return hasAnnotation(v.Metadata, ProfileAnnotation)
}

// ScanCharts filters charts using the provided predicate.
//
// TODO: Add caching based on the Status Artifact Revision.
func ScanCharts(ctx context.Context, hr *sourcev1beta1.HelmRepository, pred chartPredicate) (map[string][]string, error) {
	chartRepo, err := fetchIndexFile(hr.Status.URL)
	if err != nil {
		return nil, fmt.Errorf("fetching profiles from HelmRepository %s/%s %q: %w",
			hr.GetName(), hr.GetNamespace(), hr.Spec.URL, err)
	}

	profiles := map[string][]string{}
	for name, versions := range chartRepo.Entries {
		for _, v := range versions {
			if pred(v) {
				current, ok := profiles[name]
				if !ok {
					current = []string{}
				}
				current = append(current, v.Metadata.Version)
				profiles[name] = current
			}
		}
	}
	return profiles, nil
}

func fetchIndexFile(chartURL string) (*repo.IndexFile, error) {
	u, err := url.Parse(chartURL)
	if err != nil {
		return nil, err
	}
	c, err := DefaultChartGetters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("no provider for scheme: %s", u.Scheme)
	}

	res, err := c.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("get chart URL: %w", err)
	}

	b, err := ioutil.ReadAll(res)
	if err != nil {
		return nil, fmt.Errorf("read chart response: %w", err)
	}
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return nil, fmt.Errorf("unmarshaling chart response: %w", err)
	}
	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func hasAnnotation(cm *hapichart.Metadata, name string) bool {
	for k := range cm.Annotations {
		if k == name {
			return true
		}
	}
	return false
}
