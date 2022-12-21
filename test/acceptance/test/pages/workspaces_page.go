package pages

import (
	"fmt"

	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

type WorkspacesPage struct {
	WorkspaceHeader     *agouti.Selection
	WorkspaceHeaderLink *agouti.Selection
	WorkspacesList      *agouti.Selection
	AlertError          *agouti.Selection
}

type WorkspaceInformation struct {
	Name       *agouti.Selection
	Namespaces *agouti.Selection
	Cluster    *agouti.Selection
}

func (w WorkspacesPage) FindWorkspaceInList(workspaceName string) *WorkspaceInformation {
	workspace := w.WorkspacesList.FindByXPath(fmt.Sprintf(`//tr[.//a[.="%s"]]`, workspaceName))
	return &WorkspaceInformation{
		Name:       workspace.FindByXPath(`td[1]//a`),
		Namespaces: workspace.FindByXPath(`td[2]`),
		Cluster:    workspace.FindByXPath(`td[3]`),
	}
}

func (w WorkspacesPage) CountWorkspaces() int {
	workspaces := w.WorkspacesList.All("tr")
	count, err := workspaces.Count()
	gomega.Expect(err).Should(gomega.BeNil(), "Failed to get the number of workspaces records in the list")
	return count
}

func GetWorkspacesPage(webDriver *agouti.Page) *WorkspacesPage {
	workspacePage := WorkspacesPage{
		WorkspaceHeader:     webDriver.Find(`span[title="Workspaces"]`),
		WorkspaceHeaderLink: webDriver.Find(`div[role="heading"] a[href="/workspaces"]`),
		WorkspacesList:      webDriver.First(`table tbody`),
		AlertError:          webDriver.Find(`#alert-list-errors`),
	}
	return &workspacePage
}
