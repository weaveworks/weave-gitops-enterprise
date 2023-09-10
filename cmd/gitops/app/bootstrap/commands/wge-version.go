package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"gopkg.in/yaml.v2"
)

const (
	versionMsg = "Please select a version for WGE to be installed"
)

// SelectWgeVersion ask user to select wge version from the latest 3 versions.
func SelectWgeVersion(opts config.Options) (string, error) {
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return "", err
	}
	entitlementSecret, err := utils.GetSecret(entitlementSecretName, wgeDefaultNamespace, kubernetesClient)
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

	return utils.GetSelectInput(versionMsg, versions)
}

// fetchHelmChart helper method to fetch wge helm chart detauls.
func fetchHelmChart(username, password string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/index.yaml", wgeChartUrl), nil)
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
