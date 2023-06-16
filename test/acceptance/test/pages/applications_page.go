package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ApplicationsPage struct {
	ApplicationHeader     *agouti.Selection
	ApplicationHeaderLink *agouti.Selection
	AddApplication        *agouti.Selection
	ApplicationsList      *agouti.Selection
	SupportEmailLink      *agouti.Selection
	MessageBar            *agouti.Selection
	Version               *agouti.Selection
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
	Header     *agouti.Selection
	Title      *agouti.Selection
	Sync       *agouti.Selection
	Details    *agouti.Selection
	Events     *agouti.Selection
	Graph      *agouti.Selection
	Violations *agouti.Selection
	Edit       *agouti.Selection
}

type ApplicationDetail struct {
	Kind              *agouti.Selection
	Source            *agouti.Selection
	Chart             *agouti.Selection
	ChartVersion      *agouti.Selection
	AppliedRevision   *agouti.Selection
	AttemptedRevision *agouti.Selection
	Cluster           *agouti.Selection
	Tenant            *agouti.Selection
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

type ApplicationViolationsList struct {
	ViolationList         *agouti.Selection
	ViolationMessage      *agouti.Selection
	ViolationMessageValue *agouti.Selection
	Severity              *agouti.Selection
	SeverityValue         *agouti.Selection
	SeverityIcon          *agouti.Selection
	ViolatedPolicy        *agouti.Selection
	ViolatedPolicyValue   *agouti.Selection
	ViolationTime         *agouti.Selection
	ViolationTimeValue    *agouti.Selection
}

// ApplicationViolationsDetailsPage contains all the fields in the app violations details page
type ApplicationViolationsDetailsPage struct {
	ViolationHeader        *agouti.Selection
	PolicyName             *agouti.Selection
	PolicyNameValue        *agouti.Selection
	ClusterName            *agouti.Selection
	ClusterNameValue       *agouti.Selection
	ViolationTime          *agouti.Selection
	ViolationTimeValue     *agouti.Selection
	Severity               *agouti.Selection
	SeverityIcon           *agouti.Selection
	SeverityValue          *agouti.Selection
	Category               *agouti.Selection
	CategoryValue          *agouti.Selection
	Occurrences            *agouti.Selection
	OccurrencesCount       *agouti.Selection
	OccurrencesValue       *agouti.MultiSelection
	Description            *agouti.Selection
	DescriptionValue       *agouti.Selection
	HowToSolve             *agouti.Selection
	HowToSolveValue        *agouti.Selection
	ViolatingEntity        *agouti.Selection
	ViolatingEntityValue   *agouti.Selection
	PolicyConfigParameters *agouti.MultiSelection
}

type PolicyConfigViolationsParametersFields struct {
	ParameterName    *agouti.Selection
	ParameterValue   *agouti.Selection
	PolicyConfigName *agouti.Selection
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
	applications := a.ApplicationsList.AllByXPath(`tr[.!="No data"]`)
	count, _ := applications.Count()
	return count
}

func CountUnfilteredAppViolations(webDriver *agouti.Page, filterKey, filterValue string) int {
	violations := webDriver.All("table > tbody > tr")
	count, _ := violations.Count()
	return count
}

func GetApplicationsPage(webDriver *agouti.Page) *ApplicationsPage {
	return &ApplicationsPage{
		ApplicationHeader:     webDriver.Find(`span[title="Applications"]`),
		ApplicationHeaderLink: webDriver.Find(`div[role="heading"] a[href="/applications"]`),
		AddApplication:        webDriver.FindByButton("ADD AN APPLICATION"),
		ApplicationsList:      webDriver.First(`table tbody`),
		SupportEmailLink:      webDriver.FindByLink(`support ticket`),
		MessageBar:            webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
		Version:               webDriver.FindByXPath(`//div[starts-with(text(), "Weave GitOps Enterprise")]`),
	}
}

type ViolationsList struct {
	ViolationsList *agouti.Selection
}

func GetApplicationsDetailPage(webDriver *agouti.Page, appType string) *ApplicationDetailPage {
	return &ApplicationDetailPage{
		Header:     webDriver.Find(`div[class*=Page__TopToolBar] span`),
		Title:      webDriver.First(`div[class*="AutomationDetail"]`),
		Sync:       webDriver.FindByButton(`Sync`),
		Details:    webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/detail"`, appType)),
		Events:     webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/event"`, appType)),
		Graph:      webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/graph"`, appType)),
		Violations: webDriver.First(fmt.Sprintf(`div[role="tablist"] a[href*="/%s/violations"`, appType)),
		Edit:       webDriver.First(`a[href*="resources/edit"] button`),
	}
}

func GetApplicationDetail(webDriver *agouti.Page) *ApplicationDetail {
	autoDetails := webDriver.FirstByXPath(`//table[contains(@class, "InfoList")]/tbody`)
	reconcileDetails := webDriver.FindByXPath(`//div[contains(@class, "ReconciledObjectsTable")]//table/tbody//td[2][.="Deployment"]/ancestor::tr`)

	return &ApplicationDetail{
		Kind:              autoDetails.FindByXPath(`tr[contains(.,"Kind:")]/td[2]`),
		Source:            autoDetails.FindByXPath(`tr[contains(.,"Source:")]/td[2]`),
		Chart:             autoDetails.FindByXPath(`tr[contains(.,"Chart:")]/td[2]`),
		ChartVersion:      autoDetails.FindByXPath(`tr[contains(.,"Chart Version")]/td[2]`),
		AppliedRevision:   autoDetails.FindByXPath(`tr[contains(.,"Applied Revision")]/td[2]`),
		AttemptedRevision: autoDetails.FindByXPath(`tr[contains(.,"Attempted Revision")]/td[2]`),
		Cluster:           autoDetails.FindByXPath(`tr[contains(.,"Cluster")]/td[2]`),
		Tenant:            autoDetails.FindByXPath(`tr[contains(.,"Tenant")]/td[2]`),
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
		GitRepository:  webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="GitRepository"]/parent::node()`),
		Kustomization:  webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="Kustomization"]/parent::node()`),
		HelmRepository: webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="HelmRepository"]/parent::node()`),
		HelmRelease:    webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="HelmRelease"]/parent::node()`),
		Deployment:     webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="Deployment"]/parent::node()`),
		ReplicaSet:     webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="ReplicaSet"]/parent::node()`),
		Pod:            webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="Pod"]/parent::node()`),
	}
}

func CountAppViolations(webDriver *agouti.Page) int {
	violations := webDriver.All("table > tbody > tr")
	count, _ := violations.Count()
	return count
}

func AppViolationOccurrances(webDriver *agouti.Page, filterKey, filterValue string) int {

	var violations *agouti.MultiSelection
	switch filterKey {
	case "severity":
		violations = webDriver.AllByXPath(fmt.Sprintf(`//table/tbody/tr/td[2][normalize-space("%s")]`, filterValue))
	case "message":
		violations = webDriver.AllByXPath(fmt.Sprintf(`//table/tbody/tr[.//a[contains(@data-violation-message, "%s")]]`, filterValue))
	}
	count, _ := violations.Count()
	return count
}

// GetApplicationViolationsList will have all the locators for App Violations List page for a specific violation message.
func GetApplicationViolationsList(webDriver *agouti.Page, violationMsg string) *ApplicationViolationsList {
	violationHeader := webDriver.FirstByXPath(`//table/thead/tr`)
	violationRow := webDriver.FirstByXPath(fmt.Sprintf(`//tr[.//a[contains(@data-violation-message, "%s")]]`, violationMsg))
	applicationViolationsList := ApplicationViolationsList{
		ViolationMessage:      violationHeader.FindByXPath(`th[1]`),
		ViolationMessageValue: violationRow.FindByXPath(`td[1]//a`),
		Severity:              violationHeader.FindByXPath(`th[2]`),
		SeverityValue:         violationRow.FindByXPath(`td[2]`),
		SeverityIcon:          violationRow.FindByXPath(`td[2]//*[name()="svg"]`),
		ViolatedPolicy:        violationHeader.FindByXPath(`th[3]`),
		ViolatedPolicyValue:   violationRow.FindByXPath(`td[3]`),
		ViolationTime:         violationHeader.FindByXPath(`th[4]`),
		ViolationTimeValue:    violationRow.FindByXPath(`td[4]`),
	}
	return &applicationViolationsList
}

// GetApplicationViolationsDetailsPage returns all the locators for the app violations details page
func GetApplicationViolationsDetailsPage(webDriver *agouti.Page) *ApplicationViolationsDetailsPage {
	return &ApplicationViolationsDetailsPage{
		ViolationHeader:        webDriver.Find(`div[class*=Page__TopToolBar] span`),
		PolicyName:             webDriver.FindByXPath(`//div[text()="Policy Name"]`),
		PolicyNameValue:        webDriver.FindByXPath(`//a[contains(@href,"/policies/details?")]`),
		ClusterName:            webDriver.FindByXPath(`//div[text()="Cluster"]`),
		ClusterNameValue:       webDriver.FindByXPath(`//div[text()="Cluster"]/following-sibling::*[1]`),
		ViolationTime:          webDriver.FindByXPath(`//div/*[text()="Violation Time"]`),
		ViolationTimeValue:     webDriver.FindByXPath(`//div/*[text()="Violation Time"]/following-sibling::*[1]`),
		Severity:               webDriver.FindByXPath(`//div[text()="Severity"]`),
		SeverityIcon:           webDriver.AllByXPath(`//*[name()='svg' and contains(@class,'MuiSvgIcon')]`).At(3),
		SeverityValue:          webDriver.FindByXPath(`//div[text()="Severity"]/following-sibling::*[1]`),
		Category:               webDriver.FindByXPath(`//div[text()="Category"]`),
		CategoryValue:          webDriver.FindByXPath(`//div[text()="Category"]/following-sibling::*[1]`),
		Occurrences:            webDriver.FindByXPath(`//div[text()="Occurrences"]`),
		OccurrencesCount:       webDriver.FindByXPath(`//div[text()="Occurrences"]/span`),
		OccurrencesValue:       webDriver.AllByXPath(`//div[text()="Occurrences"]/following-sibling::*[1]/li`),
		Description:            webDriver.FindByXPath(`//div[text()="Description:"]`),
		DescriptionValue:       webDriver.FindByXPath(`//div[text()="Description:"]/following-sibling::*[1]`),
		HowToSolve:             webDriver.FindByXPath(`//div[text()="How to solve:"]`),
		HowToSolveValue:        webDriver.FindByXPath(`//div[text()="How to solve:"]/following-sibling::*[1]`),
		ViolatingEntity:        webDriver.FindByXPath(`//div[text()="Violating Entity:"]`),
		ViolatingEntityValue:   webDriver.FindByXPath(`//div[text()="Violating Entity:"]/following-sibling::*[1]`),
		PolicyConfigParameters: webDriver.AllByXPath(`//div/*[text()="Parameters Values:"]/following-sibling::*`),
	}
}

func (a ApplicationViolationsDetailsPage) GetPolicyConfigViolationsParameters(parameterName string) *PolicyConfigViolationsParametersFields {
	parametersCount, _ := a.PolicyConfigParameters.Count()
	parameterFields := PolicyConfigViolationsParametersFields{}

	for i := 0; i < parametersCount; i++ {
		if paramName, _ := a.PolicyConfigParameters.At(i).FindByXPath(`div[1]/span[2]`).Text(); paramName == parameterName {
			parameterFields = PolicyConfigViolationsParametersFields{
				ParameterName:    a.PolicyConfigParameters.At(i).FindByXPath(`div[1]/span[2]`),
				ParameterValue:   a.PolicyConfigParameters.At(i).FindByXPath(`div[2]/span[2]`),
				PolicyConfigName: a.PolicyConfigParameters.At(i).FindByXPath(`div[3]/span[2]`),
			}
		}
	}
	return &parameterFields
}
