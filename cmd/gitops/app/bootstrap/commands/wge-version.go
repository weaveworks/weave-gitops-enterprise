package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"gopkg.in/yaml.v2"
)

type HelmChartResponse struct {
	ApiVersion string
	Entries    map[string][]ChartEntry
	Generated  string
}

type ChartEntry struct {
	ApiVersion string
	Name       string
	Version    string
}

func SelectWgeVersion() (string, error) {

	entitlementSecret, err := utils.GetSecret(WGE_DEFAULT_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	versions, err := fetchHelmChart(username, password)
	if err != nil {
		return "", utils.CheckIfError(err)
	}

	versionSelectorPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please select a version for WGE to be installed",
		DefaultValue: "",
	}
	return utils.GetPromptSelect(versionSelectorPrompt, versions)
}

func fetchHelmChart(username, password string) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/index.yaml", WGE_CHART_URL), nil)
	if err != nil {
		return []string{}, utils.CheckIfError(err)
	}

	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return []string{}, utils.CheckIfError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}, utils.CheckIfError(err)
	}

	var chart HelmChartResponse
	err = yaml.Unmarshal(bodyBytes, &chart)
	if err != nil {
		return []string{}, utils.CheckIfError(err)
	}

	entries := chart.Entries["mccp"]
	var versions []string
	for _, entry := range entries {
		if entry.Name == "mccp" {
			versions = append(versions, entry.Version)
			if len(versions) == 3 {
				break
			}
		}
	}
	return versions, nil
}
