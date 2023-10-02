package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"gopkg.in/yaml.v2"
)

const (
	versionMsg = "Please select a version for WGE to be installed"
)

// SelectWgeVersion ask user to select wge version from the latest 3 versions.
func (c *Config) SelectWgeVersion() error {
	entitlementSecret, err := utils.GetSecret(c.KubernetesClient, entitlementSecretName, WGEDefaultNamespace)
	if err != nil {
		return err
	}

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])

	chartUrl := fmt.Sprintf("%s/index.yaml", wgeChartUrl)
	versions, err := fetchHelmChartVersions(chartUrl, username, password)
	if err != nil {
		return err
	}

	if c.WGEVersion == "" {
		c.WGEVersion, err = utils.GetSelectInput(versionMsg, versions)
		if err != nil {
			return err
		}
		c.Logger.Successf("Selected version %s", c.WGEVersion)
	}

	return nil
}

// fetchHelmChartVersions helper method to fetch wge helm chart versions.
func fetchHelmChartVersions(chartUrl, username, password string) ([]string, error) {
	bodyBytes, err := doBasicAuthGetRequest(chartUrl, username, password)
	if err != nil {
		return []string{}, err
	}

	var chart helmChartResponse
	err = yaml.Unmarshal(bodyBytes, &chart)
	if err != nil {
		return []string{}, err
	}
	entries := chart.Entries[wgeChartName]
	var versions []string
	for _, entry := range entries {
		if entry.Name == wgeChartName {
			versions = append(versions, entry.Version)
			if len(versions) == 3 {
				break
			}
		}
	}

	return versions, nil
}

func doBasicAuthGetRequest(url, username, password string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err

	}
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return bodyBytes, err
}
