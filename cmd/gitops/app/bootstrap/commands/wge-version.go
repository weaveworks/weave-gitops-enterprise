package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"gopkg.in/yaml.v2"
)

const (
	VERSION_MSG = "Please select a version for WGE to be installed"
)

// SelectWgeVersion ask user to select wge version from the latest 3 versions
func SelectWgeVersion() (string, error) {
	entitlementSecret, err := utils.GetSecret(WGE_DEFAULT_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil {
		return "", err
	}

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])
	if err != nil {
		return "", err
	}

	versions, err := fetchHelmChart(username, password)
	if err != nil {
		return "", err
	}

	return utils.GetSelectInput(VERSION_MSG, versions)
}

// fetchHelmChart helper method to fetch wge helm chart detauls
func fetchHelmChart(username, password string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/index.yaml", WGE_CHART_URL), nil)
	if err != nil {
		return []string{}, err

	}
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	var chart domain.HelmChartResponse
	err = yaml.Unmarshal(bodyBytes, &chart)
	if err != nil {
		return []string{}, err
	}

	entries := chart.Entries[WGE_CHART_NAME]
	var versions []string
	for _, entry := range entries {
		if entry.Name == WGE_CHART_NAME {
			versions = append(versions, entry.Version)
			if len(versions) == 3 {
				break
			}
		}
	}

	return versions, nil
}
