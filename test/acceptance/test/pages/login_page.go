package pages

import (
	"github.com/sclevine/agouti"
)

type LoginPage struct {
	Password        *agouti.Selection
	LoginOIDC       *agouti.Selection
	Continue        *agouti.Selection
	AccountSettings *agouti.Selection
}

type Account struct {
	User   *agouti.Selection
	Logout *agouti.Selection
}

type DexLoginPage struct {
	Github       *agouti.Selection
	GitlabOnPrem *agouti.Selection
	GrantAccess  *agouti.Selection
	Cancel       *agouti.Selection
}

func GetLoginPage(webDriver *agouti.Page) *LoginPage {
	loginPage := LoginPage{
		Password:        webDriver.Find(`input#password`),
		LoginOIDC:       webDriver.FindByButton(`LOGIN WITH OIDC PROVIDER`),
		Continue:        webDriver.FindByButton(`CONTINUE`),
		AccountSettings: webDriver.Find(`button[title="Account settings"]`),
	}

	return &loginPage
}

func GetAccount(webDriver *agouti.Page) *Account {
	account := Account{

		User:   webDriver.FindByXPath(`//ul/li[@role="menuitem"][contains(., "Hello")]`),
		Logout: webDriver.FindByXPath(`//ul/li[@role="menuitem"][contains(., "Logout")]`),
	}

	return &account
}

func GetDexLoginPage(webDriver *agouti.Page) *DexLoginPage {
	loginPage := DexLoginPage{

		Github:       webDriver.Find(`[class*=github] + .dex-btn-text`),
		GitlabOnPrem: webDriver.Find(`[class*=gitlab] + .dex-btn-text`),
		GrantAccess:  webDriver.FindByXPath(`//button[contains(., "Grant Access")]`),
		Cancel:       webDriver.FindByXPath(`//button[contains(., "Cancel")]`),
	}

	return &loginPage
}
