package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ApplicationsPage struct {
	ApplicationHeader *agouti.Selection
	ApplicationCount  *agouti.Selection
	ApplicationsList  *agouti.Selection
	SupportEmailLink  *agouti.Selection
	MessageBar        *agouti.Selection
	Version           *agouti.Selection
}

type ApplicationInformation struct {
	Name        *agouti.Selection
	Type        *agouti.Selection
	Namespace   *agouti.Selection
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
	Source          *agouti.Selection
	AppliedRevision *agouti.Selection
	Cluster         *agouti.Selection
	Path            *agouti.Selection
	Interval        *agouti.Selection
	LastUpdated     *agouti.Selection
	Name            *agouti.Selection
	Type            *agouti.Selection
	Namespace       *agouti.Selection
	Status          *agouti.Selection
	Message         *agouti.Selection
}

type ApplicationEvent struct {
	Reason    *agouti.Selection
	Message   *agouti.Selection
	Component *agouti.Selection
	TimeStamp *agouti.Selection
}

type ApplicationGraph struct {
	SourceGit     *agouti.Selection
	Kustomization *agouti.Selection
	Deployment    *agouti.Selection
	ReplicaSet    *agouti.Selection
	Pod           *agouti.MultiSelection
}

func (a ApplicationsPage) FindApplicationInList(applicationName string) *ApplicationInformation {
	application := a.ApplicationsList.FindByXPath(fmt.Sprintf(`//tr[.//a[.="%s"]]`, applicationName))
	return &ApplicationInformation{
		Name:        application.FindByXPath(`td[1]`),
		Type:        application.FindByXPath(`td[2]`),
		Namespace:   application.FindByXPath(`td[3]`),
		Cluster:     application.FindByXPath(`td[4]`),
		Source:      application.FindByXPath(`td[5]`),
		Status:      application.FindByXPath(`td[6]`),
		Message:     application.FindByXPath(`td[7]`),
		Revision:    application.FindByXPath(`td[8]`),
		LastUpdated: application.FindByXPath(`td[9]`),
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
		ApplicationsList:  webDriver.First(`table tbody`),
		SupportEmailLink:  webDriver.FindByLink(`support@weave.works`),
		MessageBar:        webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:           webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}
}

func GetApplicationsDetailPage(webDriver *agouti.Page) *ApplicationDetailPage {
	return &ApplicationDetailPage{
		Header:  webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/applications"]/parent::node()/parent::node()/following-sibling::div`),
		Title:   webDriver.Find(`[class*=DetailTitle]`),
		Sync:    webDriver.FindByButton(`Sync`),
		Details: webDriver.First(`div[role="tablist"] a[href*="/kustomization/detail"`),
		Events:  webDriver.First(`div[role="tablist"] a[href*="/kustomization/event"`),
		Graph:   webDriver.First(`div[role="tablist"] a[href*="/kustomization/graph"`),
	}
}

func GetApplicationDetail(webDriver *agouti.Page) *ApplicationDetail {
	autoDetails := webDriver.FirstByXPath(`//table[contains(@class, "InfoList")]/tbody`)
	reconcileDetails := webDriver.FindByXPath(`//div[contains(@class, "ReconciledObjectsTable")]//table/tbody//td[2][.="Deployment"]/ancestor::tr`)

	return &ApplicationDetail{
		Source:          autoDetails.FindByXPath(`tr[1]/td[2]`),
		AppliedRevision: autoDetails.FindByXPath(`tr[2]/td[2]`),
		Cluster:         autoDetails.FindByXPath(`tr[3]/td[2]`),
		Path:            autoDetails.FindByXPath(`tr[4]/td[2]`),
		Interval:        autoDetails.FindByXPath(`tr[5]/td[2]`),
		LastUpdated:     autoDetails.FindByXPath(`tr[6]/td[2]`),
		Name:            reconcileDetails.FindByXPath(`td[1]`),
		Type:            reconcileDetails.FindByXPath(`td[2]`),
		Namespace:       reconcileDetails.FindByXPath(`td[3]`),
		Status:          reconcileDetails.FindByXPath(`td[4]`),
		Message:         reconcileDetails.FindByXPath(`td[5]`),
	}
}

func GetApplicationEvent(webDriver *agouti.Page, reason string) *ApplicationEvent {
	events := webDriver.AllByXPath(fmt.Sprintf(`//div[contains(@class,"EventsTable")]//table/tbody//td[1][.="%s"]/ancestor::tr`, reason))

	return &ApplicationEvent{
		Reason:    events.At(0).FindByXPath(`td[1]`),
		Message:   events.At(0).FindByXPath(`td[2]`),
		Component: events.At(0).FindByXPath(`td[3]`),
		TimeStamp: events.At(0).FindByXPath(`td[4]`),
	}
}

func GetApplicationGraph(webDriver *agouti.Page, namespace string, targetNamespace string) *ApplicationGraph {
	return &ApplicationGraph{
		SourceGit:     webDriver.FindByXPath(fmt.Sprintf(`//div[@class="kind-text"][.="GitRepository"]/parent::node()/following-sibling::div/div[@class="kind-text"][.="%s"]`, namespace)),
		Kustomization: webDriver.FindByXPath(fmt.Sprintf(`//div[@class="kind-text"][.="Kustomization"]/parent::node()/following-sibling::div/div[@class="kind-text"][.="%s"]`, namespace)),
		Deployment:    webDriver.FindByXPath(fmt.Sprintf(`//div[@class="kind-text"][.="Deployment"]/parent::node()/following-sibling::div/div[@class="kind-text"][.="%s"]`, targetNamespace)),
		ReplicaSet:    webDriver.FindByXPath(fmt.Sprintf(`//div[@class="kind-text"][.="ReplicaSet"]/parent::node()/following-sibling::div/div[@class="kind-text"][.="%s"]`, targetNamespace)),
		Pod:           webDriver.AllByXPath(fmt.Sprintf(`//div[@class="kind-text"][.="Pod"]/parent::node()/following-sibling::div/div[@class="kind-text"][.="%s"]`, targetNamespace)),
	}
}
