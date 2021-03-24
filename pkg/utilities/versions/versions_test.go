package versions

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckUpgradeConstraints(t *testing.T) {
	assert.Equal(t, fmt.Errorf("1.12.1 is not a valid Kubernetes version; must be 1.16.x-1.20.x"), CheckUpgradeConstraints("1.17.1", "1.12.1", false))

	assert.Equal(t, fmt.Errorf("2.15.1 is not a valid Kubernetes version; must be 1.16.x-1.20.x"), CheckUpgradeConstraints("1.17.1", "2.15.1", false))

	assert.NoError(t, CheckUpgradeConstraints("1.16.3", "1.17.7", false))

	assert.NoError(t, CheckUpgradeConstraints("1.16.3", "1.17.10", false))

	assert.NoError(t, CheckUpgradeConstraints("1.16.8", "1.17.5", false))

	assert.NoError(t, CheckUpgradeConstraints("1.16.3", "1.17.99", true))

	assert.NoError(t, CheckUpgradeConstraints("1.18.1", "1.19.0", false))

	assert.NoError(t, CheckUpgradeConstraints("1.19.0", "1.20.0", false))

	assert.Equal(t, fmt.Errorf("downgrade not supported"), CheckUpgradeConstraints("1.20.0", "1.19.0", false))

	assert.Equal(t, fmt.Errorf("downgrade not supported"), CheckUpgradeConstraints("1.20.4", "1.20.2", false))
}

func TestGetSupportedUpgradeForVersion(t *testing.T) {
	upgradeVersions, err := GetSupportedUpgradesForVersion("1.14.1")
	expectedVersions := VersionList{"1.14.10", "1.15.7", "1.15.8", "1.15.9", "1.15.10", "1.15.11"}
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, upgradeVersions)

	expectedVersions = VersionList{"1.17.7"}
	upgradeVersions, err = GetSupportedUpgradesForVersion("1.17.5")
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, upgradeVersions)

	upgradeVersions, err = GetSupportedUpgradesForVersion("1.15.7")
	expectedVersions = VersionList{"1.15.8", "1.15.9", "1.15.10", "1.15.11",
		"1.16.3", "1.16.4", "1.16.5", "1.16.6", "1.16.7", "1.16.8", "1.16.11"}
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, upgradeVersions)
}

func TestGetVersionParts(t *testing.T) {
	parts, err := GetVersionParts("1.15.3")
	expectedParts := []int{1, 15, 3}
	assert.NoError(t, err)
	assert.Equal(t, expectedParts, parts)

	parts, err = GetVersionParts("1.15.10")
	expectedParts = []int{1, 15, 10}
	assert.NoError(t, err)
	assert.Equal(t, expectedParts, parts)

	_, err = GetVersionParts("1.test.10")
	assert.Error(t, err)

	_, err = GetVersionParts("1.16")
	assert.Error(t, err)

	_, err = GetVersionParts("")
	assert.Error(t, err)
}

func TestVersionList(t *testing.T) {
	unsortedVersionList := VersionList{}
	sortedVersionList := VersionList{}
	sort.Sort(unsortedVersionList)
	assert.Equal(t, sortedVersionList, unsortedVersionList)

	unsortedVersionList = VersionList{"1.15.3"}
	sortedVersionList = VersionList{"1.15.3"}
	sort.Sort(unsortedVersionList)
	assert.Equal(t, sortedVersionList, unsortedVersionList)

	unsortedVersionList = VersionList{"1.15.3", "1.14.1", "1.14.10"}
	sortedVersionList = VersionList{"1.14.1", "1.14.10", "1.15.3"}
	sort.Sort(unsortedVersionList)
	assert.Equal(t, sortedVersionList, unsortedVersionList)

	unsortedVersionList = VersionList{"1.15.3", "1.14.1", "1.14.10", "1.13.15"}
	sortedVersionList = VersionList{"1.13.15", "1.14.1", "1.14.10", "1.15.3"}
	sort.Sort(unsortedVersionList)
	assert.Equal(t, sortedVersionList, unsortedVersionList)
}

func TestCheckVersionExists(t *testing.T) {
	err := CheckVersionExists("")
	assert.Error(t, err)

	err = CheckVersionExists("foo")
	assert.Error(t, err)

	// Should succeed without 'v' prefix
	err = CheckVersionExists("1.17.9")
	assert.NoError(t, err)

	err = CheckVersionExists("v1.17.9")
	assert.NoError(t, err)

	err = CheckVersionExists("v1.17.99")
	assert.Error(t, err)

	err = CheckVersionExists("v1.16.13")
	assert.NoError(t, err)

	err = CheckVersionExists("v1.16.99")
	assert.Error(t, err)

	err = CheckVersionExists("v1.15.12")
	assert.NoError(t, err)

	err = CheckVersionExists("v1.15.99")
	assert.Error(t, err)

	err = CheckVersionExists("v1.14.10")
	assert.NoError(t, err)

	err = CheckVersionExists("v1.14.99")
	assert.Error(t, err)
}
