package helm

import (
	"sort"

	"github.com/Masterminds/semver"
)

// ConvertStringListToSemanticVersionList converts a slice of strings into a slice of semantic version.
func ConvertStringListToSemanticVersionList(versions []string) ([]*semver.Version, error) {
	var result []*semver.Version

	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			return nil, err
		}

		result = append(result, ver)
	}

	return result, nil
}

// SortVersions sorts semver versions in decreasing order.
func SortVersions(versions []*semver.Version) {
	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
}
