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

type WorkspaceDetailsPage struct {
	Header                    *agouti.Selection
	GoToTenantApplicationsBtn *agouti.Selection
	WorkspaceName             *agouti.Selection
	Namespaces                *agouti.Selection
	ServiceAccountsTab        *agouti.Selection
	RolesTab                  *agouti.Selection
	RoleBindingsTab           *agouti.Selection
	PoliciesTab               *agouti.Selection
}

type ServiceAccounts struct {
	Name      *agouti.Selection
	Namespace *agouti.Selection
	Age       *agouti.Selection
}

type Roles struct {
	Name      *agouti.Selection
	Namespace *agouti.Selection
	Rules     *agouti.Selection
	Age       *agouti.Selection
}

type RoleBindings struct {
	Name      *agouti.Selection
	Namespace *agouti.Selection
	Bindings  *agouti.Selection
	Role      *agouti.Selection
	Age       *agouti.Selection
}

type Policies struct {
	Name     *agouti.Selection
	Category *agouti.Selection
	Severity *agouti.Selection
	Age      *agouti.Selection
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

func GetWorkspaceDetailsPage(webDriver *agouti.Page) *WorkspaceDetailsPage {
	return &WorkspaceDetailsPage{
		Header:                    webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/workspaces"]/parent::node()/parent::node()/following-sibling::div`),
		GoToTenantApplicationsBtn: webDriver.FindByXPath(`//span[normalize-space()='go to TENANT applications']`),
		WorkspaceName:             webDriver.FindByXPath(`//div[@data-testid='Workspace Name']`),
		Namespaces:                webDriver.FindByXPath(`//div[@data-testid='Namespaces']`),
		ServiceAccountsTab:        webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/serviceAccounts?"]`),
		RolesTab:                  webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/roles?"]`),
		RoleBindingsTab:           webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/roleBindings?"]`),
		PoliciesTab:               webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/policies?"]`),
	}
}

// func GetWorkspaceServiceAccounts(webDriver *agouti.Page) *ServiceAccounts {
// 	return &ServiceAccounts{
// 		Name:      webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="GitRepository"]/parent::node()`),
// 		Namespace: webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="Kustomization"]/parent::node()`),
// 		Age:       webDriver.FirstByXPath(`//div[contains(@class, "GraphNode__NodeText")]/div[contains(@class, "GraphNode__Kinds")][.="HelmRepository"]/parent::node()`),
// 	}
// }

func GetWorkspaceServiceAccounts(webDriver *agouti.Page) *ServiceAccounts {
	return &ServiceAccounts{
		Name:      webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[1]`),
		Namespace: webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[2]`),
		Age:       webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[3]`),
	}
}
