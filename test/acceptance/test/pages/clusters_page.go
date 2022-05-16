package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ClusterInformation struct {
	Checkbox         *agouti.Selection
	ShowStatusDetail *agouti.Selection
	Name             *agouti.Selection
	Status           *agouti.Selection
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
	ClusterCount          *agouti.Selection
	ConnectClusterButton  *agouti.Selection
	PRDeleteClusterButton *agouti.Selection
	ClustersList          *agouti.Selection
	Tooltip               *agouti.Selection
	SupportEmailLink      *agouti.Selection
	MessageBar            *agouti.Selection
	Version               *agouti.Selection
}

// FindClusterInList finds the cluster with given name
func FindClusterInList(clustersPage *ClustersPage, clusterName string) *ClusterInformation {
	cluster := clustersPage.ClustersList.FindByXPath(fmt.Sprintf(`//*[@data-cluster-name="%s"]/ancestor::tr`, clusterName))
	return &ClusterInformation{
		Checkbox:         cluster.FindByXPath(`td[1]`),
		ShowStatusDetail: cluster.FindByXPath(`td[2]`).Find(`svg`),
		Name:             cluster.FindByXPath(`td[2]`),
		Status:           cluster.FindByXPath(`td[5]`),
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
		ClustersList:          webDriver.FirstByXPath(`#clusters-list table tbody`),
		Tooltip:               webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink:      webDriver.FindByLink(`support@weave.works`),
		MessageBar:            webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:               webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}

	return &clustersPage
}
