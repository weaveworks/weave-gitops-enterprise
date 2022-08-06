package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type ViolationsPage struct {
	ViolationHeader *agouti.Selection
	ViolationCount  *agouti.Selection
	ViolationList   *agouti.Selection
}

type ViolationInformation struct {
	Message     *agouti.Selection
	Severity    *agouti.Selection
	Cluster     *agouti.Selection
	Time        *agouti.Selection
	Application *agouti.Selection
}

type ViolationDetailPage struct {
	Header           *agouti.Selection
	Title            *agouti.Selection
	Message          *agouti.Selection
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
	violation := v.ViolationList.FirstByXPath(fmt.Sprintf(`//tr[.//a[contains(text(), "%s")]]`, violationMsg))
	return &ViolationInformation{
		Message:     violation.FindByXPath(`td[1]`),
		Severity:    violation.FindByXPath(`td[2]`),
		Cluster:     violation.FindByXPath(`td[3]`),
		Time:        violation.FindByXPath(`td[4]`),
		Application: violation.FindByXPath(`td[5]`),
	}
}

func (v ViolationsPage) CountViolations() int {
	violations := v.ViolationList.All("tr")
	count, _ := violations.Count()
	return count
}

func GetViolationsPage(webDriver *agouti.Page) *ViolationsPage {
	return &ViolationsPage{
		ViolationHeader: webDriver.Find(`div[role="heading"] a[href="/clusters/violations"]`),
		ViolationCount:  webDriver.First(`.section-header-count`),
		ViolationList:   webDriver.First(`table tbody`),
	}
}

func GetViolationDetailPage(webDriver *agouti.Page) *ViolationDetailPage {
	return &ViolationDetailPage{
		Header:           webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/clusters/violations"]/parent::node()/parent::node()/following-sibling::div`),
		Title:            webDriver.First(`h2`),
		Message:          webDriver.FindByXPath(`//div[text()="Message:"]/following-sibling::*[1]`),
		ClusterName:      webDriver.FindByXPath(`//div[text()="Cluster Name:"]/following-sibling::*[1]`),
		Time:             webDriver.FindByXPath(`//div/*[text()="Violation Time:"]/following-sibling::*[1]`),
		Severity:         webDriver.FindByXPath(`//div[text()="Severity:"]/following-sibling::*[1]`),
		Category:         webDriver.FindByXPath(`//div[text()="Category:"]/following-sibling::*[1]`),
		Application:      webDriver.FindByXPath(`//div[text()="Application:"]/following-sibling::*[1]`),
		OccurrencesCount: webDriver.FindByXPath(`//div[text()="Occurences"]/span`),
		Occurrences:      webDriver.AllByXPath(`//div[text()="Occurences"]/following-sibling::*[1]/li`),
		Description:      webDriver.FindByXPath(`//div[text()="Description:"]/following-sibling::*[1]`),
		HowToSolve:       webDriver.FindByXPath(`//div[text()="How to solve:"]/following-sibling::*[1]`),
		ViolatingEntity:  webDriver.FindByXPath(`//div[text()="Violating Entity:"]/following-sibling::*[1]`),
	}
}
