package pages

import (
	"fmt"
	"time"

	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

type ClusterInformation struct {
	Checkbox    *agouti.Selection
	Name        *agouti.Selection
	Dashboards  *agouti.Selection
	Type        *agouti.Selection
	Namespace   *agouti.Selection
	Status      *agouti.Selection
	Message     *agouti.Selection
	EditCluster *agouti.Selection
}

type ClusterStatus struct {
	Phase            *agouti.Selection
	KubeConfigButton *agouti.Selection
}

type ClusterInfrastructure struct {
	Kind       *agouti.Selection
	ApiVersion *agouti.Selection
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

// ClustersPage elements
type ClustersPage struct {
	ClusterHeader         *agouti.Selection
	ConnectClusterButton  *agouti.Selection
	PRDeleteClusterButton *agouti.Selection
	ClustersList          *agouti.Selection
	Tooltip               *agouti.Selection
	SupportEmailLink      *agouti.Selection
	MessageBar            *agouti.Selection
	Version               *agouti.Selection
}

type ClusterDetailPage struct {
	Header       *agouti.Selection
	Applications *agouti.Selection
	Kubeconfig   *agouti.Selection
	Namespace    *agouti.Selection
	Dashboards   *agouti.Selection
	Labels       *agouti.MultiSelection
}

// This function waits for progressbar circle to disappear
func WaitForPageToLoad(webDriver *agouti.Page) {
	gomega.Eventually(func(g gomega.Gomega) bool {
		if pCount, _ := webDriver.All(`[class^=MuiCircularProgress]`).Count(); pCount > 0 {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.BeFalse(), "Page took too long to load")
}

// FindClusterInList finds the cluster with given name
func (c ClustersPage) FindClusterInList(clusterName string) *ClusterInformation {
	cluster := c.ClustersList.FindByXPath(fmt.Sprintf(`//*[@data-cluster-name="%s"]/ancestor::tr`, clusterName))
	return &ClusterInformation{
		Checkbox:    cluster.FindByXPath(`td[1]`).Find("input"),
		Name:        cluster.FindByXPath(`td[2]`),
		Dashboards:  cluster.FindByXPath(`td[3]`),
		Type:        cluster.FindByXPath(`td[4]//*[@role="img"]`),
		Namespace:   cluster.FindByXPath(`td[5]`),
		Status:      cluster.FindByXPath(`td[6]//div/*[last()][name()="div"]`),
		Message:     cluster.FindByXPath(`td[7]`),
		EditCluster: cluster.FindByXPath(`td[8]//button`),
	}
}

func (c ClusterInformation) GetDashboard(dashboard string) *agouti.Selection {
	return c.Dashboards.FindByXPath(fmt.Sprintf(`//li/a[.="%s"]`, dashboard))
}

func (c ClustersPage) CountClusters() int {
	clusters := c.ClustersList.All("[data-cluster-name]")
	count, _ := clusters.Count()
	return count
}

func GetClusterStatus(webDriver *agouti.Page) *ClusterStatus {
	clusterStatus := ClusterStatus{
		Phase:            webDriver.FindByXPath(`//tr/th[.="phase"]/following-sibling::td`),
		KubeConfigButton: webDriver.FindByButton(`Kubeconfig`),
	}

	return &clusterStatus
}

func GetClusterInfrastructure(webDriver *agouti.Page) *ClusterInfrastructure {
	return &ClusterInfrastructure{
		Kind:       webDriver.FindByXPath(`//tr/td[.="Kind:"]/following-sibling::td`),
		ApiVersion: webDriver.FindByButton(`//tr/td[.="APIVersion:"]/following-sibling::td`),
	}
}

func GetDeletePRPopup(webDriver *agouti.Page) *DeletePullRequestPopup {
	deletePRPopup := DeletePullRequestPopup{
		Title:               webDriver.Find(`#delete-popup h5`),
		PRDescription:       webDriver.FindByID(`PULL REQUEST DESCRIPTION-input`),
		ClosePopup:          webDriver.Find(`#delete-popup > div > button[type=button]`),
		DeleteClusterButton: webDriver.Find(`#delete-popup button#delete-cluster`),
		ConfirmDelete:       webDriver.Find(`#confirm-disconnect-cluster-dialog button:first-child`),
		CancelDelete:        webDriver.Find(`#confirm-disconnect-cluster-dialog button:last-child`),
		GitCredentials:      webDriver.Find(`div.auth-message`),
	}

	return &deletePRPopup
}

// GetClustersPage initialises the webDriver object
func GetClustersPage(webDriver *agouti.Page) *ClustersPage {
	clustersPage := ClustersPage{
		ClusterHeader:         webDriver.Find(`span[title="Clusters"]`),
		ConnectClusterButton:  webDriver.Find(`#connect-cluster`),
		PRDeleteClusterButton: webDriver.Find(`#delete-cluster`),
		ClustersList:          webDriver.First(`.clusters-list table tbody`),
		Tooltip:               webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink:      webDriver.FindByLink(`support ticket`),
		MessageBar:            webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:               webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}

	return &clustersPage
}

func GetClusterDetailPage(webDriver *agouti.Page) *ClusterDetailPage {
	infoList := webDriver.Find(`table[class*="InfoList"]`)
	return &ClusterDetailPage{
		Header:       webDriver.Find(`div[class*=Page__TopToolBar] span[class*=Breadcrumbs]`),
		Applications: infoList.FindByButton(`GO TO APPLICATIONS`),
		Kubeconfig:   infoList.FindByXPath(`//td[.="kubeconfig:"]/following-sibling::td`),
		Namespace:    webDriver.FindByXPath(`//td[.="Namespace:"]/following-sibling::td`),
		Dashboards:   webDriver.FindByXPath(`//div[.="Dashboards"]//following-sibling::ul`),
		Labels:       webDriver.AllByXPath(`//div[.="Labels"]//following-sibling::div`),
	}
}

func (c ClusterDetailPage) GetDashboard(dashboard string) *agouti.Selection {
	return c.Dashboards.FindByXPath(fmt.Sprintf(`//li/a[.="%s"]`, dashboard))
}

func (c ClusterDetailPage) GetLabels() []string {
	labels := []string{}
	tCount, _ := c.Labels.Count()

	for i := 0; i < tCount; i++ {
		label, _ := c.Labels.At(i).Text()
		labels = append(labels, label)
	}
	return labels
}
