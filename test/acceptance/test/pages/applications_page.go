package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ApplicationsPage struct {
	ApplicationHeader *agouti.Selection
	ApplicationCount  *agouti.Selection
	AddApplication    *agouti.Selection
	ApplicationsList  *agouti.Selection
	SupportEmailLink  *agouti.Selection
	MessageBar        *agouti.Selection
	Version           *agouti.Selection
}

type ApplicationInformation struct {
	Name        *agouti.Selection
	Type        *agouti.Selection
	Namespace   *agouti.Selection
	Tenant      *agouti.Selection
	Cluster     *agouti.Selection
	Source      *agouti.Selection
	Status      *agouti.Selection
	Message     *agouti.Selection
	Revision    *agouti.Selection
	LastUpdated *agouti.Selection
}

type ApplicationDetailPage struct {
	Header  *agouti.Selection
	Title   *agouti.Selection
	Sync    *agouti.Selection
	Details *agouti.Selection
	Events  *agouti.Selection
	Graph   *agouti.Selection
}

type ApplicationDetail struct {
	Source            *agouti.Selection
	Chart             *agouti.Selection
	ChartVersion      *agouti.Selection
	AppliedRevision   *agouti.Selection
	AttemptedRevision *agouti.Selection
	Cluster           *agouti.Selection
	Path              *agouti.Selection
	Interval          *agouti.Selection
	LastUpdated       *agouti.Selection
	Metadata          *agouti.Selection
	Name              *agouti.Selection
	Type              *agouti.Selection
	Namespace         *agouti.Selection
	Status            *agouti.Selection
	Message           *agouti.Selection
}

type ApplicationEvent struct {
	Reason    *agouti.Selection
	Message   *agouti.Selection
	Component *agouti.Selection
	TimeStamp *agouti.Selection
}

type ApplicationGraph struct {
	GitRepository  *agouti.Selection
	Kustomization  *agouti.Selection
	HelmRepository *agouti.Selection
	HelmRelease    *agouti.Selection
	Deployment     *agouti.Selection
	ReplicaSet     *agouti.Selection
	Pod            *agouti.Selection
}

func (a ApplicationsPage) FindApplicationInList(applicationName string) *ApplicationInformation {
	application := a.ApplicationsList.FindByXPath(fmt.Sprintf(`//tr[.//a[.="%s"]]`, applicationName))
	return &ApplicationInformation{
		Name:        application.FindByXPath(`td[2]//a`),
		Type:        application.FindByXPath(`td[3]`),
		Namespace:   application.FindByXPath(`td[4]`),
		Tenant:      application.FindByXPath(`td[5]`),
		Cluster:     application.FindByXPath(`td[6]`),
		Source:      application.FindByXPath(`td[7]//a`),
		Status:      application.FindByXPath(`td[8]`),
		Message:     application.FindByXPath(`td[9]`),
		Revision:    application.FindByXPath(`td[10]`),
		LastUpdated: application.FindByXPath(`td[11]`),
	}
}

func (a ApplicationsPage) CountApplications() int {
	applications := a.ApplicationsList.All("tr")
	count, _ := applications.Count()
	return count
}

func GetApplicationsPage(webDriver *agouti.Page) *ApplicationsPage {
	return &ApplicationsPage{
		ApplicationHeader: webDriver.Find(`div[role="heading"] a[href="/applications"]`),
		ApplicationCount:  webDriver.Find(`.section-header-count`),
		AddApplication:    webDriver.FindByButton("ADD AN APPLICATION"),
		ApplicationsList:  webDriver.First(`table tbody`),
		SupportEmailLink:  webDriver.FindByLink(`support ticket`),
		MessageBar:        webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:           webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}
}

func GetApplicationsDetailPage(webDriver *agouti.Page, appType string) *ApplicationDetailPage {
	return &ApplicationDetailPage{
		Header:  webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/applications"]/parent::node()/parent::node()/following-sibling::div`),
		Title:   webDriver.First(`div[class*="AutomationDetail"]`),
		Sync:    webDriver.FindByButton(`Sync`),
		Details: webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/detail"`, appType)),
		Events:  webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/event"`, appType)),
		Graph:   webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/graph"`, appType)),
	}
}

func GetApplicationDetail(webDriver *agouti.Page) *ApplicationDetail {
	autoDetails := webDriver.FirstByXPath(`//table[contains(@class, "InfoList")]/tbody`)
	reconcileDetails := webDriver.FindByXPath(`//div[contains(@class, "ReconciledObjectsTable")]//table/tbody//td[2][.="Deployment"]/ancestor::tr`)

	return &ApplicationDetail{
		Source:            autoDetails.FindByXPath(`tr[1]/td[2]`),
		Chart:             autoDetails.FindByXPath(`tr[contains(.,"Chart:")]/td[2]`),
		ChartVersion:      autoDetails.FindByXPath(`tr[contains(.,"Chart Version")]/td[2]`),
		AppliedRevision:   autoDetails.FindByXPath(`tr[contains(.,"Applied Revision")]/td[2]`),
		AttemptedRevision: autoDetails.FindByXPath(`tr[contains(.,"Attempted Revision")]/td[2]`),
		Cluster:           autoDetails.FindByXPath(`tr[contains(.,"Cluster")]/td[2]`),
		Path:              autoDetails.FindByXPath(`tr[contains(.,"Path:")]/td[2]`),
		Interval:          autoDetails.FindByXPath(`tr[contains(.,"Interval")]/td[2]`),
		LastUpdated:       autoDetails.FindByXPath(`tr[contains(.,"Last Updated")]/td[2]`),
		Metadata:          webDriver.Find(`div[class*=Metadata] table tbody`),
		Name:              reconcileDetails.FindByXPath(`td[1]`),
		Type:              reconcileDetails.FindByXPath(`td[2]`),
		Namespace:         reconcileDetails.FindByXPath(`td[3]`),
		Status:            reconcileDetails.FindByXPath(`td[4]`),
		Message:           reconcileDetails.FindByXPath(`td[5]`),
	}
}

func (a ApplicationDetail) GetMetadata(name string) *agouti.Selection {
	return a.Metadata.FindByXPath(fmt.Sprintf(`tr/td[.="%s:"]/following-sibling::td`, name))
}

func GetApplicationEvent(webDriver *agouti.Page, reason string) *ApplicationEvent {
	events := webDriver.FirstByXPath(fmt.Sprintf(`//div[contains(@class,"EventsTable")]//table/tbody//td[1][.="%s"]/ancestor::tr`, reason))

	return &ApplicationEvent{
		Reason:    events.FindByXPath(`td[1]`),
		Message:   events.FindByXPath(`td[2]`),
		Component: events.FindByXPath(`td[3]`),
		TimeStamp: events.FindByXPath(`td[4]`),
	}
}

func GetApplicationGraph(webDriver *agouti.Page) *ApplicationGraph {
	return &ApplicationGraph{
		GitRepository:  webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="GitRepository"]/parent::node()`),
		Kustomization:  webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="Kustomization"]/parent::node()`),
		HelmRepository: webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="HelmRepository"]/parent::node()`),
		HelmRelease:    webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="HelmRelease"]/parent::node()`),
		Deployment:     webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="Deployment"]/parent::node()`),
		ReplicaSet:     webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="ReplicaSet"]/parent::node()`),
		Pod:            webDriver.FirstByXPath(`//div[contains(@class, "GraphNode")]/following-sibling::div[contains(@class, "GraphNode")][.="Pod"]/parent::node()`),
	}
}
