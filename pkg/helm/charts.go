package helm

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
)

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
