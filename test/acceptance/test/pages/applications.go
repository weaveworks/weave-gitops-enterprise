package pages

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/gomega"

	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type ApplicationPage struct {
	ApplicationHeader *agouti.Selection
	ApplicationCount  *agouti.Selection
	AddApplication    *agouti.Selection
	ApplicationTable  *agouti.MultiSelection
}

type Applicationrow struct {
	Application *agouti.Selection
}

type Conditions struct {
	Type    *agouti.Selection
	Status  *agouti.Selection
	Reason  *agouti.Selection
	Message *agouti.Selection
}

type ApplicationDetails struct {
	Name           *agouti.Selection
	DeploymentType *agouti.Selection
	URL            *agouti.Selection
	Path           *agouti.Selection
}

// This function waits for application graph to be rendered
func (a ApplicationDetails) WaitForPageToLoad(webDriver *agouti.Page) {
	Eventually(webDriver.Find(` svg g.output`)).Should(BeVisible(), "Application details failed to load/render as expected")
}

// This function waits for main application page to be rendered
func (a ApplicationPage) WaitForPageToLoad(webDriver *agouti.Page, appCount int) {
	Eventually(a.ApplicationCount).Should(BeVisible())

	isLoaded := func() bool {
		_ = webDriver.Refresh()
		if aCnt, _ := a.ApplicationCount.Text(); aCnt >= strconv.Itoa(appCount) {
			return true
		}
		return false
	}
	Eventually(isLoaded, 2*time.Minute, 5*time.Second).Should(BeTrue(), "Application page failed to load/render as expected")
}

func GetApplicationPage(webDriver *agouti.Page) *ApplicationPage {
	applicationPage := ApplicationPage{
		ApplicationHeader: webDriver.Find(`div[role="heading"] a[href="/applications"]`),
		ApplicationCount:  webDriver.FindByXPath(`//*[@href="/applications"]/parent::div[@role="heading"]/following-sibling::div`),
		AddApplication:    webDriver.FindByButton(`Add Application`),
		ApplicationTable:  webDriver.All(`tr.MuiTableRow-root a`),
	}

	return &applicationPage
}

func GetApplicationRow(applicationPage *ApplicationPage, applicationeName string) *Applicationrow {
	aCnt, _ := applicationPage.ApplicationTable.Count()
	for i := 0; i < aCnt; i++ {
		aName, _ := applicationPage.ApplicationTable.At(i).Text()
		if applicationeName == aName {
			return &Applicationrow{
				Application: applicationPage.ApplicationTable.At(i),
			}
		}
	}
	return nil
}

func GetApplicationDetails(webDriver *agouti.Page) *ApplicationDetails {
	appDetalis := webDriver.Find(`div[role="list"] > table tr`)

	return &ApplicationDetails{
		Name:           appDetalis.FindByXPath(`td[1]`),
		DeploymentType: appDetalis.FindByXPath(`td[2]`),
		URL:            appDetalis.FindByXPath(`td[3]`),
		Path:           appDetalis.FindByXPath(`td[4]`),
	}
}

func GetApplicationConditions(webDriver *agouti.Page, condition string) *Conditions {
	sourceConditions := webDriver.FindByXPath(fmt.Sprintf(`//h3[.="%s"]/following-sibling::div[1]//tbody/tr`, condition))

	return &Conditions{
		Type:    sourceConditions.FindByXPath(`td[1]`),
		Status:  sourceConditions.FindByXPath(`td[2]`),
		Reason:  sourceConditions.FindByXPath(`td[3]`),
		Message: sourceConditions.FindByXPath(`td[4]`),
	}
}
