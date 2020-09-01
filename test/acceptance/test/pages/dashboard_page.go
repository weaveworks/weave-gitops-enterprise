package pages

import (
	"github.com/sclevine/agouti"
)

//DashboardwebDriver elements
type DashboardwebDriver struct {
	WKPTitle          *agouti.Selection
	WKPDocLink        *agouti.Selection
	ClusterName       *agouti.Selection
	K8SVersion        *agouti.Selection
	AlertInfo         *agouti.Selection
	GrafanaLink       *agouti.Selection
	AddComponentsLink *agouti.Selection
	OpenGitRepoLink   *agouti.Selection
}

//Dashboard initialises the webDriver object
func Dashboard(webDriver *agouti.Page) *DashboardwebDriver {
	dashboard := DashboardwebDriver{
		WKPTitle:          webDriver.FindByXPath(`//*[@id="app"]/div/div[1]/div/a[1]`),
		WKPDocLink:        webDriver.FindByXPath(`//*[@id="app"]/div/div[1]/div/a[2]`),
		ClusterName:       webDriver.FindByXPath(`//*[@id="app"]/div/div[2]/div/div[1]/div[1]/div/div[1]/span[2]`),
		K8SVersion:        webDriver.Find("#wkp-ui-cluster-version"),
		AlertInfo:         webDriver.FindByXPath(`//*[@id="app"]/div/div[2]/div/div[1]/div[2]/i`),
		GrafanaLink:       webDriver.FindByLink(`View Grafana dashboards`),
		AddComponentsLink: webDriver.FindByLink(`Add components`),
		OpenGitRepoLink:   webDriver.FindByLink(`Open git repo`)}

	return &dashboard
}
