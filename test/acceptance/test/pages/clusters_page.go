package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ClusterInformation struct {
	Checkbox         *agouti.Selection
	Name             *agouti.Selection
	Type             *agouti.Selection
	Namespace        *agouti.Selection
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
	ClusterHeader         *agouti.Selection
	ClusterCount          *agouti.Selection
	ConnectClusterButton  *agouti.Selection
	PRDeleteClusterButton *agouti.Selection
	ClustersList          *agouti.Selection
	Tooltip               *agouti.Selection
	SupportEmailLink      *agouti.Selection
	MessageBar            *agouti.Selection
	Version               *agouti.Selection
}

// This function waits for progressbar circle to disappear
func (t ClustersPage) WaitForPageToLoad(webDriver *agouti.Page) {
	Eventually(webDriver.Find(`[class^=MuiCircularProgress]`)).ShouldNot(BeFound())
}

// FindClusterInList finds the cluster with given name
func FindClusterInList(clustersPage *ClustersPage, clusterName string) *ClusterInformation {
	cluster := clustersPage.ClustersList.FindByXPath(fmt.Sprintf(`//*[@data-cluster-name="%s"]/ancestor::tr`, clusterName))
	return &ClusterInformation{
		Checkbox:         cluster.FindByXPath(`td[1]`).Find("input"),
		Name:             cluster.FindByXPath(`td[2]`),
		Type:             cluster.FindByXPath(`td[4]`),
		Namespace:        cluster.FindByXPath(`td[5]`),
		Status:           cluster.FindByXPath(`td[6]//div/*[last()][name()="div"]`),
	}
}

func CountClusters(clustersPage *ClustersPage) int {
	clusters := clustersPage.ClustersList.All("div[data-cluster-name]")
	count, _ := clusters.Count()
	return count
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
		ClusterHeader:         webDriver.Find(`div[role="heading"] a[href="/clusters"]`),
		ClusterCount:          webDriver.Find(`.count-header .section-header-count`),
		ConnectClusterButton:  webDriver.Find(`#connect-cluster`),
		PRDeleteClusterButton: webDriver.Find(`#delete-cluster`),
		ClustersList:          webDriver.First(`#clusters-list table tbody`),
		Tooltip:               webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink:      webDriver.FindByLink(`support@weave.works`),
		MessageBar:            webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:               webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}

	return &clustersPage
}
