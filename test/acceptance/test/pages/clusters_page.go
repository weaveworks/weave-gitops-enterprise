package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ClusterInformation struct {
	Checkbox         *agouti.Selection
	ShowStatusDetail *agouti.Selection
	Name             *agouti.Selection
	Status *agouti.Selection
}

type ClusterStatus struct {
	Phase            *agouti.Selection
	KubeConfigButton *agouti.Selection
}

type DeletePullRequestPopup struct {
	Title               *agouti.Selection
	ClosePopup          *agouti.Selection
	PRDescription       *agouti.Selection
	DeleteClusterButton *agouti.Selection
	ConfirmDelete       *agouti.Selection
	CancelDelete        *agouti.Selection
	GitCredentials      *agouti.Selection
}

//ClustersPage elements
type ClustersPage struct {
	ClusterCount                                *agouti.Selection
	ConnectClusterButton                        *agouti.Selection
	PRDeleteClusterButton                       *agouti.Selection
	ClustersListSection                         *agouti.Selection
	ClustersListHeader                          *agouti.Selection
	FiringAlertsPerPage                         *agouti.Selection
	FiringAlerts                                *agouti.MultiSelection
	HeaderCheckbox                              *agouti.Selection
	HeaderName                                  *agouti.Selection
	HeaderIcon                                  *agouti.Selection
	HeaderStatus                                *agouti.Selection
	ClustersList                                *agouti.Selection
	Tooltip                                     *agouti.Selection
	SupportEmailLink                            *agouti.Selection
	MessageBar                                  *agouti.Selection
	Version                                     *agouti.Selection
}

// This function waits for cluster to appear in the cluste table (become visible)
func (c ClustersPage) WaitForClusterToAppear(webDriver *agouti.Page, clusterName string) {
	Eventually(webDriver.Find(fmt.Sprintf(`#clusters-list > div > div[2] > div[1] > div[1] > table > tbody > tr.summary[data-cluster-name="%s"]`, clusterName))).Should(BeFound())
}

// FindClusterInList finds the cluster with given name
func FindClusterInList(clustersPage *ClustersPage, clusterName string) *ClusterInformation {
	cluster := clustersPage.ClustersList.Find(fmt.Sprintf(`tr.summary[data-cluster-name="%s"]`, clusterName))
	return &ClusterInformation{
		Checkbox:         cluster.FindByXPath(`td[1]`),
		ShowStatusDetail: cluster.FindByXPath(`td[2]`).Find(`svg`),
		Name:             cluster.FindByXPath(`td[2]`),
		Status: cluster.FindByXPath(`td[5]`),
	}
}

func GetClusterStatus(webDriver *agouti.Page) *ClusterStatus {
	clusterStatus := ClusterStatus{
		Phase:            webDriver.FindByXPath(`//tr/th[.="phase"]/following-sibling::td`),
		KubeConfigButton: webDriver.FindByXPath(`//button[.="Download the kubeconfig here"]`),
	}

	return &clusterStatus
}

func GetDeletePRPopup(webDriver *agouti.Page) *DeletePullRequestPopup {
	deletePRPopup := DeletePullRequestPopup{
		Title:               webDriver.Find(`#delete-popup h5`),
		PRDescription:       webDriver.Find(`#delete-popup textarea`),
		ClosePopup:          webDriver.Find(`#delete-popup > div > button[type=button]`),
		DeleteClusterButton: webDriver.Find(`#delete-popup button#delete-cluster`),
		ConfirmDelete:       webDriver.Find(`#confirm-disconnect-cluster-dialog button:first-child`),
		CancelDelete:        webDriver.Find(`#confirm-disconnect-cluster-dialog button:last-child`),
		GitCredentials:      webDriver.Find(`div.auth-message`),
	}

	return &deletePRPopup
}

//GetClustersPage initialises the webDriver object
func GetClustersPage(webDriver *agouti.Page) *ClustersPage {
	clustersPage := ClustersPage{
		ClusterCount:          webDriver.Find(`.count-header .section-header-count`),
		ConnectClusterButton:  webDriver.Find(`#connect-cluster`),
		PRDeleteClusterButton: webDriver.Find(`#delete-cluster`),
		ClustersListSection:   webDriver.Find(`#clusters-list`),
		ClustersListHeader:    webDriver.FindByXPath(`//*[@id="clusters-list"]/div/div[2]/div[1]/div[1]/table/thead`),
		HeaderCheckbox:        webDriver.FindByXPath(`//*[@id="clusters-list"]/div/div[2]/div[1]/div[1]/table/thead/tr/th[1]/span`),
		HeaderName:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/div[2]/div[1]/div[1]/table/thead/tr/th[2]/span`),
		HeaderIcon:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/div[2]/div[1]/div[1]/table/thead/tr/th[3]/span`),
		HeaderStatus:          webDriver.FindByXPath(`//*[@id="clusters-list"]/div/div[2]/div[1]/div[1]/table/thead/tr/th[4]/span`),
		ClustersList:     webDriver.Find(`#clusters-list > div > div[2] > div[1] > div[1] > table > tbody`),
		Tooltip:          webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink: webDriver.FindByLink(`support@weave.works`),
		MessageBar: webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:    webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}

	return &clustersPage
}
