package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type AuthenticateGithub struct {
	AuthenticateGithub *agouti.Selection
	AccessCode         *agouti.Selection
	AuthroizeButton    *agouti.Selection
	AuthorizationError *agouti.Selection
	Close              *agouti.Selection
}

type DeviceActivationGitHub struct {
	Username            *agouti.Selection
	Password            *agouti.Selection
	Signin              *agouti.Selection
	UserCode            *agouti.MultiSelection
	AuthCode            *agouti.Selection
	Verify              *agouti.Selection
	Continue            *agouti.Selection
	AuthroizeWeaveworks *agouti.Selection
	ConfirmPassword     *agouti.Selection
	ConnectedMessage    *agouti.Selection
}

type AuthenticateGitlab struct {
	AuthenticateGitlab *agouti.Selection

	Username      *agouti.Selection
	Password      *agouti.Selection
	Authorize     *agouti.Selection
	Signin        *agouti.Selection
	CheckBrowser  *agouti.Selection
	AcceptCookies *agouti.Selection
}

type LoginPage struct {
	Username        *agouti.Selection
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

func WaitForAuthenticationAlert(webDriver *agouti.Page, alert_success_msg string) {
	Eventually(webDriver.FindByXPath(fmt.Sprintf(`//div[@class="MuiAlert-message"][.="%s"]`, alert_success_msg))).Should(BeVisible())
}

func AuthenticateWithGithub(webDriver *agouti.Page) *AuthenticateGithub {
	return &AuthenticateGithub{
		AuthenticateGithub: webDriver.FindByButton(`Authenticate with GitHub`),
		// FIXME: bit brittle
		AccessCode:         webDriver.FindByXPath(`//button[contains(.,'Authorize Github Access')]/../../preceding-sibling::div/span`),
		AuthroizeButton:    webDriver.FindByButton(`Authorize Github Access`),
		AuthorizationError: webDriver.FindByXPath(`//div[@role="alert"]//div[.="Error"]`),
		Close:              webDriver.FindByButton(`Close`),
	}
}

func ActivateDeviceGithub(webDriver *agouti.Page) *DeviceActivationGitHub {
	return &DeviceActivationGitHub{
		Username:            webDriver.Find(`input[type=text][name=login]`),
		Password:            webDriver.Find(`input[type=password][name*=password]`),
		Signin:              webDriver.Find(`input[type=submit][value="Sign in"]`),
		UserCode:            webDriver.All(`input[type=text][name^=user-code-]`),
		AuthCode:            webDriver.Find(`input#otp`),
		Verify:              webDriver.FindByButton(`Verify`),
		Continue:            webDriver.Find(`[type=submit][name=commit]`),
		AuthroizeWeaveworks: webDriver.FindByButton(`Authorize weaveworks`),
		ConfirmPassword:     webDriver.FindByButton(`password`),
		ConnectedMessage:    webDriver.FindByXPath(`//p[contains(text(), "device is now connected")]`),
	}
}

func AuthenticateWithGitlab(webDriver *agouti.Page) *AuthenticateGitlab {
	return &AuthenticateGitlab{
		AuthenticateGitlab: webDriver.FindByButton(`Authenticate with GitLab`),
		Authorize:          webDriver.Find(`input[name="commit"][value="Authorize"]`),
		Username:           webDriver.Find(`#user_login`),
		Password:           webDriver.Find(`#user_password`),
		Signin:             webDriver.Find(`button[data-qa-selector=sign_in_button]`),
		AcceptCookies:      webDriver.Find(`#onetrust-accept-btn-handler`),
		CheckBrowser:       webDriver.Find(`span[data-translate=checking_browser]`),
	}
}

func AuthenticateWithOnPremGitlab(webDriver *agouti.Page) *AuthenticateGitlab {
	return &AuthenticateGitlab{
		AuthenticateGitlab: webDriver.FindByButton(`Authenticate with GitLab`),
		Authorize:          webDriver.Find(`input[name="commit"][value="Authorize"]`),
		Username:           webDriver.Find(`#user_login`),
		Password:           webDriver.Find(`#user_password`),
		Signin:             webDriver.Find(`input[name=commit]`),
		AcceptCookies:      webDriver.Find(`#onetrust-accept-btn-handler`),
		CheckBrowser:       webDriver.Find(`span[data-translate=checking_browser]`),
	}
}

func GetLoginPage(webDriver *agouti.Page) *LoginPage {
	loginPage := LoginPage{
		Username:        webDriver.Find(`input#email`),
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
