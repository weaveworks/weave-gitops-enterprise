package versions

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var kubectlDownloadURL = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/amd64/kubectl"

var k8sVersionRegexp = regexp.MustCompile(`^([1][.](14|15|16|17)[.][0-9][0-9]?)$`)

// ValidVersions maps the k8s versions WKP supports
var ValidVersions = map[string]bool{
	"1.14.1":  true,
	"1.14.10": true,
	"1.15.7":  true,
	"1.15.8":  true,
	"1.15.9":  true,
	"1.15.10": true,
	"1.15.11": true,
	"1.16.3":  true,
	"1.16.4":  true,
	"1.16.5":  true,
	"1.16.6":  true,
	"1.16.7":  true,
	"1.16.8":  true,
	"1.16.11": true,
	"1.17.5":  true,
	"1.17.7":  true,
}

// VersionList is a list to contain versions in string format,
// with a custom comparator for sorting them
type VersionList []string

func (vl VersionList) Len() int      { return len(vl) }
func (vl VersionList) Swap(i, j int) { vl[i], vl[j] = vl[j], vl[i] }
func (vl VersionList) Less(i, j int) bool {
	iParts, _ := GetVersionParts(vl[i])
	jParts, _ := GetVersionParts(vl[j])

	for versionPart := range []int{0, 1, 2} {
		if iParts[versionPart] < jParts[versionPart] {
			return true
		} else if iParts[versionPart] > jParts[versionPart] {
			return false
		}
	}
	return false
}

// CheckUpgradeConstraints verifies that the versions for the upgrade are valid
func CheckUpgradeConstraints(oldVersion, newVersion string, skipExistenceCheck bool) error {
	err := CheckValidVersion(oldVersion)
	if err != nil {
		return err
	}
	err = CheckValidVersion(newVersion)
	if err != nil {
		return err
	}
	if !skipExistenceCheck {
		err = CheckVersionExists(newVersion)
		if err != nil {
			return err
		}
	}
	err = CheckVersionRange(oldVersion, newVersion)
	if err != nil {
		return err
	}
	return nil
}

// CheckValidVersion checks that a string is a correctly formatted k8s version that WKP supports
func CheckValidVersion(version string) error {
	if !k8sVersionRegexp.MatchString(version) {
		return fmt.Errorf(
			"%s is not a valid Kubernetes version; must be 1.14.x-1.17.x",
			version)
	}
	return nil
}

// CheckVersionExists checks if a k8s version exists by getting the status code of the equivalent kubectl download URL
func CheckVersionExists(version string) error {
	// Check string has min length of 5: a.b.c
	if len(version) < 5 {
		return fmt.Errorf("version string %s is not valid", version)
	}

	// Prepend 'v' if not there
	versionString := version
	if string(version[0]) != "v" {
		versionString = "v" + version
	}

	renderedURL := fmt.Sprintf(kubectlDownloadURL, versionString)

	resp, err := http.Head(renderedURL)

	logrus.Debugf("URL %s returned status: %v", renderedURL, resp.StatusCode)
	if err == nil && resp.StatusCode == 200 {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("version %s was not found by checking %s and is probably not a valid kubernetes version.\n"+
		"You can skip this check with --skip-check-exists", versionString, renderedURL)
}

// CheckVersionRange checks if an upgrade from one version to another is valid
func CheckVersionRange(oldVersion, newVersion string) error {
	oldParts := strings.Split(oldVersion, ".")
	newParts := strings.Split(newVersion, ".")

	if newParts[0] != oldParts[0] {
		return fmt.Errorf("cannot upgrade across major versions")
	}

	// We know the versions are valid so Atoi can't fail
	oldMinor, _ := strconv.Atoi(oldParts[1])
	newMinor, _ := strconv.Atoi(newParts[1])
	oldMicro, _ := strconv.Atoi(oldParts[2])
	newMicro, _ := strconv.Atoi(newParts[2])

	if oldMinor > newMinor || (oldMinor == newMinor && oldMicro > newMicro) {
		return fmt.Errorf("downgrade not supported")
	}

	if newMinor-oldMinor > 1 {
		return fmt.Errorf("cannot upgrade across more than one minor version")
	}
	return nil
}

// GetSupportedUpgrades returns all the supported direct K8s version bumps for the cluster
func GetSupportedUpgrades() map[string][]string {
	supportedUpgrades := map[string][]string{}
	for validVersion := range ValidVersions {
		supportedVersions, err := GetSupportedUpgradesForVersion(validVersion)
		if err != nil {
			return nil
		}
		supportedUpgrades[validVersion] = supportedVersions
		if err != nil {
			return nil
		}
	}
	return supportedUpgrades
}

// GetSupportedVersions returns a list of supported K8s versions to be
// installed on the clusters (based on the map of supported upgrades)
func GetSupportedVersions() VersionList {
	versions := VersionList{}
	for validVersion := range ValidVersions {
		versions = append(versions, validVersion)
	}
	sort.Sort(versions)
	return versions
}

// GetVersionParts splits a version string to a slice of 3 elements: major, minor, patch
func GetVersionParts(version string) ([]int, error) {
	var parts []int

	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("version is not valid: %v\nPlease specify all 3 version parts: major.minor.patch", version)
	}
	for _, part := range versionParts {
		intPart, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version from %s\nerr: %s", version, err)
		}
		parts = append(parts, intPart)
	}
	return parts, nil
}

// GetSupportedUpgradesForVersion returns all the valid upgrade target versions for a given version
func GetSupportedUpgradesForVersion(version string) (VersionList, error) {
	supportedUpgrades := VersionList{}

	versionParts, err := GetVersionParts(version)
	if err != nil {
		return nil, err
	}

	validVersions := GetSupportedVersions()

	for _, validVersion := range validVersions {
		validVersionParts, err := GetVersionParts(validVersion)
		if err != nil {
			return nil, err
		}
		if (validVersionParts[1] == versionParts[1] && validVersionParts[2] > versionParts[2]) ||
			validVersionParts[1] == versionParts[1]+1 {
			supportedUpgrades = append(supportedUpgrades, validVersion)
		}
	}
	sort.Sort(supportedUpgrades)
	return supportedUpgrades, nil
}
