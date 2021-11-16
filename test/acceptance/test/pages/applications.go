package pages

import (
	"fmt"
	"strconv"

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

type AuthenticateGithub struct {
	AuthenticateGithub *agouti.Selection
	AccessCode         *agouti.Selection
	AuthroizeButton    *agouti.Selection
}

type DeviceActivation struct {
	Username            *agouti.Selection
	Password            *agouti.Selection
	Signin              *agouti.Selection
	UserCode            *agouti.MultiSelection
	Verify              *agouti.Selection
	Continue            *agouti.Selection
	AuthroizeWeaveworks *agouti.Selection
	ConfirmPassword     *agouti.Selection
	ConnectedMessage    *agouti.Selection
}

type Commits struct {
	SHA     *agouti.Selection
	Date    *agouti.Selection
	Message *agouti.Selection
	Author  *agouti.Selection
}

func WaitForAuthenticationAlert(webDriver *agouti.Page) {
	Eventually(webDriver.FindByXPath(`//div[@class="MuiAlert-message"][.="Authentication Successful"]`)).Should(BeVisible())
}

// This function waits for application graph to be rendered
func (a ApplicationDetails) WaitForPageToLoad(webDriver *agouti.Page) {
	Eventually(webDriver.Find(` svg g.output`)).Should(BeVisible(), "Application details failed to load/render as expected")
}

// This function waits for main application page to be rendered
func (a ApplicationPage) WaitForPageToLoad(webDriver *agouti.Page, appCount int) {
	Eventually(a.ApplicationCount).Should(BeVisible())
	strCount, _ := a.ApplicationCount.Text()
	count, _ := strconv.Atoi(strCount)
	Expect(count).Should(BeNumerically(">=", appCount), "Application page failed to load/render as expected")
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

func AuthenticateWithGithub(webDriver *agouti.Page) *AuthenticateGithub {
	return &AuthenticateGithub{
		AuthenticateGithub: webDriver.FindByButton(`Authenticate with Github`),
		AccessCode:         webDriver.Find(`div[class*=GithubDeviceAuthModal__P]:first-child`),
		AuthroizeButton:    webDriver.FindByButton(`Authorize Github Access`),
	}
}

func ActivateDevice(webDriver *agouti.Page) *DeviceActivation {
	return &DeviceActivation{
		Username:            webDriver.Find(`input[type=text][name=login]`),
		Password:            webDriver.Find(`input[type=password][name*=password]`),
		Signin:              webDriver.Find(`input[type=submit][value="Sign in"]`),
		UserCode:            webDriver.All(`input[type=text][name^=user-code-]`),
		Verify:              webDriver.FindByButton(`Verify`),
		Continue:            webDriver.Find(`[type=submit][name=commit]`),
		AuthroizeWeaveworks: webDriver.FindByButton(`Authorize weaveworks`),
		ConfirmPassword:     webDriver.FindByButton(`password`),
		ConnectedMessage:    webDriver.FindByXPath(`//p[contains(text(), "device is now connected")]`),
	}
}

func GetCommits(webDriver *agouti.Page) []Commits {
	Eventually(webDriver.All(`div[class^=CommitsTable] thead tr`)).Should(BeVisible())

	commits := webDriver.All(`div[class^=CommitsTable] tbody > tr`)
	cCnt, _ := commits.Count()

	retCommits := []Commits{}

	for i := 0; i < cCnt; i++ {
		retCommits = append(retCommits, Commits{
			SHA:     commits.At(i).FindByXPath(`td[1]`),
			Date:    commits.At(i).FindByXPath(`td[2]`),
			Message: commits.At(i).FindByXPath(`td[3]`),
			Author:  commits.At(i).FindByXPath(`td[4]`),
		})
	}
	return retCommits
}
