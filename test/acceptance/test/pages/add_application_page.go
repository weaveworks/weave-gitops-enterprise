package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type AddApplicationPage struct {
	ApplicationHeader *agouti.Selection
}

type AddApplication struct {
	Name                  *agouti.Selection
	Namespace             *agouti.Selection
	TargetNamespace       *agouti.Selection
	Path                  *agouti.Selection
	Source                *agouti.Selection
	Cluster               *agouti.Selection
	RemoveApplication     *agouti.Selection
	SourceHref            *agouti.Selection
	CreateTargetNamespace *agouti.Selection
}

type GitOps struct {
	GitOpsLabel     *agouti.Selection
	BranchName      *agouti.Selection
	PullRequestTile *agouti.Selection
	CommitMessage   *agouti.Selection
	PullRequestDesc *agouti.Selection
	GitCredentials  *agouti.Selection
	CreatePR        *agouti.Selection
	SuccessBar      *agouti.Selection
	PRLinkBar       *agouti.Selection
	ErrorBar        *agouti.Selection
}

func GetAddApplicationsPage(webDriver *agouti.Page) *AddApplicationPage {
	return &AddApplicationPage{
		ApplicationHeader: webDriver.Find(`div[role="heading"] a[href="/applications"]`),
	}
}

func GetAddApplication(webDriver *agouti.Page, appNo ...int) *AddApplication {
	app := webDriver.FirstByXPath(`//div/form`)
	if len(appNo) > 0 {
		app = webDriver.FindByXPath(fmt.Sprintf(`//h3[.="Application No.%d"]/parent::div`, appNo[0]))
	}

	return &AddApplication{
		Name:                  app.Find(`[id="KUSTOMIZATION NAME-input"]`),
		Namespace:             app.Find(`[id="KUSTOMIZATION NAMESPACE-input"]`),
		TargetNamespace:       app.Find(`[id="TARGET NAMESPACE-input"]`),
		Path:                  app.Find(`[id="SELECT PATH-input"]`),
		Source:                app.Find(`[id="SELECT SOURCE-input"]`),
		Cluster:               app.Find(`[id="SELECT CLUSTER-input"]`),
		RemoveApplication:     app.Find(`button#remove-application`),
		CreateTargetNamespace: app.First(`input[type="checkbox"]`),
		SourceHref:            app.FindByXPath(`div[contains(@class, "MuiGrid-container")]/div[2]`),
	}
}

func (a AddApplication) SelectListItem(webDriver *agouti.Page, itemName string) *agouti.Selection {
	return webDriver.FindByXPath(fmt.Sprintf(`//ul/li[contains(., "%s")]`, itemName))
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	return GitOps{
		GitOpsLabel:     webDriver.FindByXPath(`//h2[.="GitOps"]`),
		BranchName:      webDriver.FindByID(`CREATE BRANCH-input`),
		PullRequestTile: webDriver.FindByID(`PULL REQUEST TITLE-input`),
		CommitMessage:   webDriver.FindByID(`COMMIT MESSAGE-input`),
		PullRequestDesc: webDriver.FindByID(`PULL REQUEST DESCRIPTION-input`),
		GitCredentials:  webDriver.Find(`div.auth-message`),
		CreatePR:        webDriver.FindByButton(`CREATE PULL REQUEST`),
		SuccessBar:      webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Success")]/parent::node()`),
		PRLinkBar:       webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "PR created")]/parent::node()`),
		ErrorBar:        webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Error")]/parent::node()`),
	}
}
