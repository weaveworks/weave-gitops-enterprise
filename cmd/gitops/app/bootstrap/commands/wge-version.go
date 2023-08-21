package commands

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	if err != nil {
		fmt.Printf("An error occurred\n%v", err)
		os.Exit(1)
	}

	username, password := string(entitlementSecret.Data["username"]), string(entitlementSecret.Data["password"])
	if err != nil {
		fmt.Printf("An error occurred\n%v", err)
		os.Exit(1)
	}

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
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}

	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error performing request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}

	var chart HelmChart
	err = yaml.Unmarshal(bodyBytes, &chart)
	if err != nil {
		log.Fatalf("error parsing yaml: %v", err)
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
	return versions
}
