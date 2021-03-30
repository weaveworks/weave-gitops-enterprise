package pages

import (
	"github.com/sclevine/agouti"
)

//ClusterConnectionPage elements
type ClusterConnectionPage struct {
	ClusterConnectionPopup   *agouti.Selection
	ClusterName              *agouti.Selection
	ClusterIngressURL        *agouti.Selection
	ClusterSaveAndNext       *agouti.Selection
	ButtonClusterSaveAndNext *agouti.Selection
	ConnectionInstructions   *agouti.Selection
	ConnectionStatus         *agouti.Selection
	ButtonClose              *agouti.Selection
	DisconnectTab            *agouti.Selection
	ButtonRemoveCluster      *agouti.Selection
}

//GetClusterConnectionPage initialises the webDriver object
func GetClusterConnectionPage(webDriver *agouti.Page) *ClusterConnectionPage {
	clusterConnPage := ClusterConnectionPage{
		ClusterConnectionPopup:   webDriver.Find(`#connection-popup`),
		ClusterName:              webDriver.Find(`#Name-input`),
		ClusterIngressURL:        webDriver.FindByXPath(`//*[@id="Ingress URL-input"]`),
		ClusterSaveAndNext:       webDriver.FindByXPath(`//*[@id="connection-popup"]/div[2]/div/form/div[2]/button/span`),
		ButtonClusterSaveAndNext: webDriver.FindByXPath(`//*[@id="connection-popup"]/div[2]/div/form/div[2]/button`),
		ConnectionInstructions:   webDriver.Find(`#instructions`),
		ConnectionStatus:         webDriver.FindByXPath(`//*[@id="connection-popup"]/div[2]/div/form/div[1]/div/div[2]/div[2]`),
		ButtonClose:              webDriver.Find(`#connection-popup button[type="submit"]`),
		DisconnectTab:            webDriver.FindByXPath(`//*[@id="connection-popup"]/div[2]/div/div/div[3]`),
		ButtonRemoveCluster:      webDriver.FindByXPath(`//*[@id="connection-popup"]/div[2]/div/form/div[1]/div/div[3]/button/span`),
	}

	return &clusterConnPage
}
