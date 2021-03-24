package pages

import (
	"github.com/sclevine/agouti"
)

type clusterInformation struct {
	Name          *agouti.Selection
	Icon          *agouti.Selection
	Status        *agouti.Selection
	GitMessage    *agouti.Selection
	NodesVersions *agouti.Selection
	GitRepoURL    *agouti.Selection
	EditCluster   *agouti.Selection
}

//ClustersPage elements
type ClustersPage struct {
	ClusterCount         *agouti.Selection
	ConnectClusterButton *agouti.Selection
	NoFiringAlertMessage *agouti.Selection
	FiringAlertsSection  *agouti.Selection
	FiringAlertsHeader   *agouti.Selection
	FiringAlertsNavCtl   *agouti.Selection
	ClustersListSection  *agouti.Selection
	ClustersListHeader   *agouti.Selection
	HeaderName           *agouti.Selection
	HeaderIcon           *agouti.Selection
	HeaderStatus         *agouti.Selection
	HeaderGitActivity    *agouti.Selection
	HeaderNodeVersion    *agouti.Selection
	NoClusterConfigured  *agouti.Selection
	ClustersList         *agouti.MultiSelection
	SupportEmailLink     *agouti.Selection
}

// FindClusterInList finds the cluster with given name
func FindClusterInList(clustersPage *ClustersPage, clusterName string) *clusterInformation {
	count, _ := clustersPage.ClustersList.Count()
	for i := 0; i < count; i++ {
		cluster := clustersPage.ClustersList.At(i)
		name, _ := cluster.FindByXPath(`td[1]`).Text()

		if name == clusterName {
			return &clusterInformation{
				Name:          cluster.FindByXPath(`td[1]`),
				Icon:          cluster.FindByXPath(`td[2]`),
				Status:        cluster.FindByXPath(`td[3]`),
				GitMessage:    cluster.FindByXPath(`td[4]`),
				NodesVersions: cluster.FindByXPath(`td[5]`),
				GitRepoURL:    cluster.FindByXPath(`td[6]`),
				EditCluster:   cluster.FindByXPath(`td[7]`),
			}
		}
	}

	// So you can return something to Eventually that won't panic
	nullSelector := clustersPage.HeaderName.Find("#null-selector-foo")
	return &clusterInformation{
		Name:          nullSelector,
		Icon:          nullSelector,
		Status:        nullSelector,
		GitMessage:    nullSelector,
		NodesVersions: nullSelector,
		GitRepoURL:    nullSelector,
		EditCluster:   nullSelector,
	}
}

//GetClustersPage initialises the webDriver object
func GetClustersPage(webDriver *agouti.Page) *ClustersPage {
	clustersPage := ClustersPage{
		ClusterCount:         webDriver.FindByXPath(`//*[@id="count-header"]/div/div[2]`),
		ConnectClusterButton: webDriver.Find(`#connect-cluster`),
		NoFiringAlertMessage: webDriver.FindByXPath(`//*[@id="app"]/div/div[2]/div[1]/i`),
		FiringAlertsSection:  webDriver.Find(`#firing-alerts`),
		FiringAlertsHeader:   webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[1]/div`),
		FiringAlertsNavCtl:   webDriver.FindByXPath(`//*[@id="firing-alerts"]/div/div/div[2]/div/p/span/span[2]`),
		ClustersListSection:  webDriver.Find(`#clusters-list`),
		ClustersListHeader:   webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead`),
		HeaderName:           webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[1]/span`),
		HeaderIcon:           webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[2]/span`),
		HeaderStatus:         webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[3]/span`),
		HeaderGitActivity:    webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[4]/span`),
		HeaderNodeVersion:    webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/thead/tr/th[5]/span`),
		NoClusterConfigured:  webDriver.FindByXPath(`//*[@id="clusters-list"]/div/table/caption`),
		ClustersList:         webDriver.All(`#clusters-list > div > table > tbody > tr`),
		SupportEmailLink:     webDriver.FindByLink(`support@weave.works`)}

	return &clustersPage
}
