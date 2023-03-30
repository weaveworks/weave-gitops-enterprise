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
	Name                    *agouti.Selection
	ServiceAccountsManifest *agouti.Selection
	ManifestCloseBtn        *agouti.Selection
	Namespace               *agouti.Selection
	Age                     *agouti.Selection
}

type Roles struct {
	Name             *agouti.Selection
	RoleApi          *agouti.Selection
	ManifestCloseBtn *agouti.Selection
	Namespace        *agouti.Selection
	Rules            *agouti.Selection
	RulesBtn         *agouti.Selection
	ViewRules        *agouti.Selection
	CloseBtn         *agouti.Selection
	Age              *agouti.Selection
}

type RoleBindings struct {
	Name             *agouti.Selection
	RoleBindingApi   *agouti.Selection
	ManifestCloseBtn *agouti.Selection
	Namespace        *agouti.Selection
	Bindings         *agouti.Selection
	Role             *agouti.Selection
	Age              *agouti.Selection
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
		Header:                    webDriver.Find(".test-id-breadcrumbs > :last-child"),
		GoToTenantApplicationsBtn: webDriver.FindByXPath(`//span[normalize-space()='go to TENANT applications']`),
		WorkspaceName:             webDriver.FindByXPath(`//div[@data-testid='Workspace Name']`),
		Namespaces:                webDriver.FindByXPath(`//div[@data-testid='Namespaces']`),
		ServiceAccountsTab:        webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/serviceAccounts?"]`),
		RolesTab:                  webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/roles?"]`),
		RoleBindingsTab:           webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/roleBindings?"]`),
		PoliciesTab:               webDriver.First(`div[role="tablist"] a[href*="/workspaces/details/policies?"]`),
	}
}

func GetWorkspaceServiceAccounts(webDriver *agouti.Page) *ServiceAccounts {
	return &ServiceAccounts{
		Name:                    webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[1]`),
		ServiceAccountsManifest: webDriver.FindByXPath(`(//span[@class='token key'][normalize-space()='apiVersion'])[1]`),
		ManifestCloseBtn:        webDriver.FindByXPath(`(//*[name()='svg'][@class='MuiSvgIcon-root'])[3]`),
		Namespace:               webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[2]`),
		Age:                     webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[3]`),
	}
}

func GetWorkspaceRoles(webDriver *agouti.Page) *Roles {
	return &Roles{
		Name:             webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[1]`),
		RoleApi:          webDriver.FindByXPath(`(//span[@class='token key'][normalize-space()='apiVersion'])[1]`),
		ManifestCloseBtn: webDriver.FindByXPath(`(//*[name()='svg'][@class='MuiSvgIcon-root'])[3]`),
		Namespace:        webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[2]`),
		Rules:            webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[3]`),
		RulesBtn:         webDriver.FindByXPath(`(//span[contains(text(),'View Rules')])[1]`),
		ViewRules:        webDriver.FindByXPath(`(//label[normalize-space()='Resources:'])[1]`),
		CloseBtn:         webDriver.FindByXPath(`(//*[name()='svg'][@class='MuiSvgIcon-root'])[3]`),
		Age:              webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[4]`),
	}
}

func GetWorkspaceRoleBindings(webDriver *agouti.Page) *RoleBindings {
	return &RoleBindings{
		Name:             webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[1]`),
		RoleBindingApi:   webDriver.FindByXPath(`(//span[@class='token key'][normalize-space()='apiVersion'])[1]`),
		ManifestCloseBtn: webDriver.FindByXPath(`(//*[name()='svg'][@class='MuiSvgIcon-root'])[3]`),
		Namespace:        webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[2]`),
		Bindings:         webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[3]`),
		Role:             webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[4]`),
		Age:              webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[5]`),
	}
}

func GetWorkspacePolicies(webDriver *agouti.Page) *Policies {
	return &Policies{
		Name:     webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[1]`),
		Category: webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[2]`),
		Severity: webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[3]`),
		Age:      webDriver.FindByXPath(`(//td[@class='MuiTableCell-root MuiTableCell-body'])[4]`),
	}
}
