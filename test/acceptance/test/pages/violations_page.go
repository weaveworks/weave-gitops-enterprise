package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ViolationsPage struct {
	ViolationHeader *agouti.Selection
	ViolationList   *agouti.Selection
}

type ViolationInformation struct {
	Message         *agouti.Selection
	Cluster         *agouti.Selection
	Application     *agouti.Selection
	Severity        *agouti.Selection
	ValidatedPolicy *agouti.Selection
	Time            *agouti.Selection
}

type ViolationDetailPage struct {
	Header           *agouti.Selection
	ClusterName      *agouti.Selection
	Time             *agouti.Selection
	Severity         *agouti.Selection
	Category         *agouti.Selection
	Application      *agouti.Selection
	OccurrencesCount *agouti.Selection
	Occurrences      *agouti.MultiSelection
	Description      *agouti.Selection
	HowToSolve       *agouti.Selection
	ViolatingEntity  *agouti.Selection
}

func (v ViolationsPage) FindViolationInList(violationMsg string) *ViolationInformation {
	violation := v.ViolationList.FirstByXPath(fmt.Sprintf(`//tr[.//a[contains(@data-violation-message, "%s")]]`, violationMsg))
	return &ViolationInformation{
		Message:         violation.FindByXPath(`td[1]//a`),
		Cluster:         violation.FindByXPath(`td[2]`),
		Application:     violation.FindByXPath(`td[3]`),
		Severity:        violation.FindByXPath(`td[4]`),
		ValidatedPolicy: violation.FindByXPath(`td[5]`),
		Time:            violation.FindByXPath(`td[6]`),
	}
}

func (v ViolationsPage) CountViolations(clusterName ...string) int {
	count := 0

	if len(clusterName) > 0 {
		violations := v.ViolationList.AllByXPath(fmt.Sprintf(`//tr//td[contains(.,"%s")]`, clusterName[0]))

		count, _ = violations.Count()
	} else {
		violations := v.ViolationList.All("tr")
		count, _ = violations.Count()
	}

	return count
}

func GetViolationsPage(webDriver *agouti.Page) *ViolationsPage {
	return &ViolationsPage{
		ViolationHeader: webDriver.Find(`div[role="heading"] a[href="/clusters"]`),
		ViolationList:   webDriver.First(`table tbody`),
	}
}

func GetViolationDetailPage(webDriver *agouti.Page) *ViolationDetailPage {
	return &ViolationDetailPage{
		Header:           webDriver.Find(`div[class*=Page__TopToolBar] span[class*=Breadcrumbs]`),
		ClusterName:      webDriver.FindByXPath(`//div[text()="Cluster"]/following-sibling::*[1]`),
		Time:             webDriver.FindByXPath(`//div/*[text()="Violation Time"]/following-sibling::*[1]`),
		Severity:         webDriver.FindByXPath(`//div[text()="Severity"]/following-sibling::*[1]`),
		Category:         webDriver.FindByXPath(`//div[text()="Category"]/following-sibling::*[1]`),
		Application:      webDriver.FindByXPath(`//div[text()="Application"]/following-sibling::*[1]`),
		OccurrencesCount: webDriver.FindByXPath(`//div[text()="Occurrences"]/span`),
		Occurrences:      webDriver.AllByXPath(`//div[text()="Occurrences"]/following-sibling::*[1]/li`),
		Description:      webDriver.FindByXPath(`//div[text()="Description:"]/following-sibling::*[1]`),
		HowToSolve:       webDriver.FindByXPath(`//div[text()="How to solve:"]/following-sibling::*[1]`),
		ViolatingEntity:  webDriver.FindByXPath(`//div[text()="Violating Entity:"]/following-sibling::*[1]`),
	}
}
