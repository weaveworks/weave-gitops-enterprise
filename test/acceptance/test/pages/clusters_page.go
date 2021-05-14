package pages

import (
	"fmt"
	"strings"

	"github.com/sclevine/agouti"
)

type ClusterInformation struct {
	Name           *agouti.Selection
	Icon           *agouti.Selection
	Status         *agouti.Selection
	GitActivity    *agouti.Selection
	NodesVersions  *agouti.Selection
	TeamWorkspaces *agouti.Selection
	GitRepoURL     *agouti.Selection
	EditCluster    *agouti.Selection
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
	NoFiringAlertMessage                        *agouti.Selection
	FiringAlertsSection                         *agouti.Selection
	FiringAlertsHeader                          *agouti.Selection
	FiringAlertsNavCtl                          *agouti.Selection
	ClustersListSection                         *agouti.Selection
	ClustersListHeader                          *agouti.Selection
	FiringAlertsPerPage                         *agouti.Selection
	FiringAlerts                                *agouti.MultiSelection
	HeaderName                                  *agouti.Selection
	HeaderIcon                                  *agouti.Selection
	HeaderStatus                                *agouti.Selection
	HeaderGitActivity                           *agouti.Selection
	HeaderWorkspaces                            *agouti.Selection
	HeaderNodeVersion                           *agouti.Selection
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
}

// FindClusterInList finds the cluster with given name
func FindClusterInList(clustersPage *ClustersPage, clusterName string) *ClusterInformation {
	cluster := clustersPage.ClustersList.Find(fmt.Sprintf(`tr[data-cluster-name="%s"]`, clusterName))
	return &ClusterInformation{
		Name:           cluster.FindByXPath(`td[1]`),
		Icon:           cluster.FindByXPath(`td[2]`),
		Status:         cluster.FindByXPath(`td[3]`),
		GitActivity:    cluster.FindByXPath(`td[4]`),
		NodesVersions:  cluster.FindByXPath(`td[5]`),
		TeamWorkspaces: cluster.FindByXPath(`td[6]`),
		GitRepoURL:     cluster.FindByXPath(`td[7]`),
		EditCluster:    cluster.FindByXPath(`td[8]`),
	}
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
		ClusterCount:                          webDriver.FindByXPath(`//*[@id="count-header"]/div/div[2]`),
		ConnectClusterButton:                  webDriver.Find(`#connect-cluster`),
		NoFiringAlertMessage:                  webDriver.FindByXPath(`//*[@id="app"]/div/div[2]/div[1]/i`),
		FiringAlertsSection:                   webDriver.Find(`#firing-alerts`),
		FiringAlertsHeader:                    webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[1]/div`),
		FiringAlertsNavCtl:                    webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[2]/div/p/span/span[2]`),
		FiringAlertsPerPage:                   webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[3]`),
		FiringAlerts:                          webDriver.All(`#firing-alerts > div > table > tbody > tr`),
		ClustersListSection:                   webDriver.Find(`#clusters-list`),
		ClustersListHeader:                    webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead`),
		HeaderName:                            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[1]/span`),
		HeaderIcon:                            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[2]/span`),
		HeaderStatus:                          webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[3]/span`),
		HeaderGitActivity:                     webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[4]/span`),
		HeaderNodeVersion:                     webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[5]/span`),
		HeaderWorkspaces:                      webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[6]/span`),
		NoClusterConfigured:                   webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/caption`),
		ClustersList:                          webDriver.Find(`#clusters-list > div > table > tbody`),
		Tooltip:                               webDriver.Find(`div[role="tooltip"]`),
		SupportEmailLink:                      webDriver.FindByLink(`support@weave.works`),
		ClustersListPaginationNext:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[3]`),
		ClustersListPaginationPrevious:        webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[2]`),
		ClustersListPaginationLast:            webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[4]`),
		ClustersListPaginationFirst:           webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[3]/button[1]`),
		ClustersListPaginationPerPageDropdown: webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/tfoot/tr/td/div/div[2]/div`),
		ClustersListPaginationPerPageDropdownSecond: webDriver.FindByXPath(`//*[@id="menu-"]/div[3]/ul/li[2]`),
	}

	return &clustersPage
}
