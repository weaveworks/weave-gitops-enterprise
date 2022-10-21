package helm

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"helm.sh/helm/v3/pkg/repo"
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

func hasAnnotation(cm map[string]string, name string) bool {
	for k := range cm {
		if k == name {
			return true
		}
	}

	return false
}
