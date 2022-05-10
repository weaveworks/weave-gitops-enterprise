package pages

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ClusterInformation struct {
	Checkbox         *agouti.Selection
	ShowStatusDetail *agouti.Selection
	Name             *agouti.Selection
	// Icon             *agouti.Selection
	Status           *agouti.Selection
	// EditCluster      *agouti.Selection
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

type AlertInformation struct {
	Severity    *agouti.Selection
	Message     *agouti.Selection
	ClusterName *agouti.Selection
	TimeStamp   *agouti.Selection
}

//ClustersPage elements
type ClustersPage struct {
	ClusterCount                                *agouti.Selection
	ConnectClusterButton                        *agouti.Selection
	PRDeleteClusterButton                       *agouti.Selection
	NoFiringAlertMessage                        *agouti.Selection
	FiringAlertsSection                         *agouti.Selection
	FiringAlertsHeader                          *agouti.Selection
	FiringAlertsNavCtl                          *agouti.Selection
	ClustersListSection                         *agouti.Selection
	ClustersListHeader                          *agouti.Selection
	FiringAlertsPerPage                         *agouti.Selection
	FiringAlerts                                *agouti.MultiSelection
	HeaderCheckbox                              *agouti.Selection
	HeaderName                                  *agouti.Selection
	HeaderIcon                                  *agouti.Selection
	HeaderStatus                                *agouti.Selection
	NoClusterConfigured                         *agouti.Selection
	ClustersList                                *agouti.Selection
	Tooltip                                     *agouti.Selection
	SupportEmailLink                            *agouti.Selection
	ClustersListPaginationNext                  *agouti.Selection
	ClustersListPaginationPrevious              *agouti.Selection
	ClustersListPaginationLast                  *agouti.Selection
	ClustersListPaginationFirst                 *agouti.Selection
	ClustersListPaginationPerPageDropdown       *agouti.Selection
	ClustersListPaginationPerPageDropdownSecond *agouti.Selection
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
		// Icon:             cluster.FindByXPath(`td[3]`).Find(`svg`),
		Status:           cluster.FindByXPath(`td[5]`),
		// EditCluster:      cluster.FindByXPath(`td[5]`).Find("button"),
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

func FindAlertInFiringAlertsWidget(clustersPage *ClustersPage, alertName string) *AlertInformation {
	for _, a := range AlertsFiringInAlertsWidget(clustersPage) {
		text, _ := a.Message.Text()
		if strings.Contains(text, alertName) {
			return a
		}
	}
	return nil
}

func AlertsFiringInAlertsWidget(clustersPage *ClustersPage) []*AlertInformation {
	alertInfos := []*AlertInformation{}
	count, _ := clustersPage.FiringAlerts.Count()
	for i := 0; i < count; i++ {
		alert := clustersPage.FiringAlerts.At(i)
		alertInfos = append(alertInfos, &AlertInformation{
			Severity:    alert.FindByXPath(`td[1]`),
			Message:     alert.FindByXPath(`td[2]`),
			ClusterName: alert.FindByXPath(`td[3]`),
			TimeStamp:   alert.FindByXPath(`td[4]`),
		})

	}
	return alertInfos
}

//GetClustersPage initialises the webDriver object
func GetClustersPage(webDriver *agouti.Page) *ClustersPage {
	clustersPage := ClustersPage{
		ClusterCount:                          webDriver.Find(`.count-header .section-header-count`),
		ConnectClusterButton:                  webDriver.Find(`#connect-cluster`),
		PRDeleteClusterButton:                 webDriver.Find(`#delete-cluster`),
		NoFiringAlertMessage:                  webDriver.Find(`#firing-alerts caption`),
		FiringAlertsSection:                   webDriver.Find(`#firing-alerts`),
		FiringAlertsHeader:                    webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[1]/div`),
		FiringAlertsNavCtl:                    webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[2]/div/p/span/span[2]`),
		FiringAlertsPerPage:                   webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[3]`),
		FiringAlerts:                          webDriver.All(`#firing-alerts tbody tr`),
		ClustersListSection:                   webDriver.Find(`#clusters-list`),
		ClustersListHeader:                    webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead`),
		HeaderCheckbox:                        webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[1]/span`),
		HeaderName:                            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[2]/span`),
		HeaderIcon:                            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[3]/span`),
		HeaderStatus:                          webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[4]/span`),
		// NoClusterConfigured:                   webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/caption`),
		ClustersList:                          webDriver.Find(`#clusters-list > div > table > tbody`),
		Tooltip:                               webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink:                      webDriver.FindByLink(`support@weave.works`),
		// ClustersListPaginationNext:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[3]`),
		// ClustersListPaginationPrevious:        webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[2]`),
		// ClustersListPaginationLast:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[4]`),
		// ClustersListPaginationFirst:           webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[1]`),
		// ClustersListPaginationPerPageDropdown: webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[2]/div`),
		// ClustersListPaginationPerPageDropdownSecond: webDriver.FindByXPath(`//*[@id="menu-"]/div[3]/ul/li[2]`),
		MessageBar: webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:    webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}

	return &clustersPage
}
