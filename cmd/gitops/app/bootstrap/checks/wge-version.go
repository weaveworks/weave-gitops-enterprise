package checks

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
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
	//get secret from getSecret()
	entitlementSecret, err := getSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil {
		fmt.Printf("An error occurred %v\n", err)
		os.Exit(1)
	}

	//get username and password from entitlementSecret
	username, password, err := getSecretUsernamePassword(entitlementSecret)
	if err != nil {
		fmt.Printf("An error occurred %v\n", err)
		os.Exit(1)
	}

	versions := fetchHelmChart(username, password)

	versionSelectorPrompt := promptContent{
		"",
		"Please select a version for WGE to be installed",
	}
	return promptGetSelect(versionSelectorPrompt, versions)
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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

func getSecretUsernamePassword(secret *corev1.Secret) (string, string, error) {

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])

	// If the username and password are base64 encoded, decode them.
	if isValidBase64(username) {
		decodedUsername, err := base64.StdEncoding.DecodeString(username)
		if err != nil {
			return "", "", err
		}
		username = string(decodedUsername)
	}

	if isValidBase64(password) {
		decodedPassword, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			return "", "", err
		}
		password = string(decodedPassword)
	}

	return username, password, nil
}
