package steps

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"gopkg.in/yaml.v2"
)

func getWgeVersions(input []StepInput, c *Config) (interface{}, error) {
	entitlementSecret, err := utils.GetSecret(c.KubernetesClient, entitlementSecretName, WGEDefaultNamespace)
	if err != nil {
		return []string{}, err
	}

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])

	chartUrl := c.ChartURL
	if chartUrl == "" {
		chartUrl = fmt.Sprintf("%s/index.yaml", wgeChartUrl)
	}

	versions, err := fetchHelmChartVersions(chartUrl, username, password)
	if err != nil {
		return []string{}, err
	}
	return versions, nil
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
