package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"gopkg.in/yaml.v2"
)

const CHART_URL string = "https://charts.dev.wkp.weave.works/releases/charts-v3"

type HelmChart struct {
	ApiVersion string
	Entries    map[string][]ChartEntry
	Generated  string
}

type ChartEntry struct {
	ApiVersion string
	Name       string
	Version    string
}

func SelectWgeVersion() string {

	entitlementSecret, err := utils.GetSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	utils.CheckIfError(err)

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])
	utils.CheckIfError(err)

	versions := fetchHelmChart(username, password)

	versionSelectorPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please select a version for WGE to be installed",
		DefaultValue: "",
	}
	return utils.GetPromptSelect(versionSelectorPrompt, versions)
}

func fetchHelmChart(username, password string) []string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/index.yaml", CHART_URL), nil)
	utils.CheckIfError(err)

	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	utils.CheckIfError(err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	utils.CheckIfError(err)

	var chart HelmChart
	err = yaml.Unmarshal(bodyBytes, &chart)
	utils.CheckIfError(err)

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
	return versions
}
