package acceptance

import (
	"fmt"

	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var loginUserName string

const (
	EnableUserLogin   = false
	AdminUserName     = "admin"
	AdminUserPassword = "wego-enterprise"
	ClusterUserLogin  = "cluster-user"
	OidcUserLogin     = "oidc"
)

func loginUser(userType string) {
	loginPage := pages.GetLoginPage(webDriver)
	Eventually(loginPage.LoginOIDC).Should(BeVisible())

	switch userType {
	case ClusterUserLogin:
		// Login via cluster user account
		Expect(loginPage.Password.SendKeys(AdminUserPassword)).To(Succeed())
		Expect(loginPage.Continue.Click()).To(Succeed())
	case OidcUserLogin:
		// Login via OIDC provider
		Eventually(loginPage.LoginOIDC.Click).Should(Succeed())

		dexLogin := pages.GetDexLoginPage(webDriver)
		switch gitProviderEnv.Type {
		case GitProviderGitHub:
			Eventually(dexLogin.Github.Click).Should(Succeed())

			authenticate := pages.ActivateDeviceGithub(webDriver)

			if pages.ElementExist(authenticate.Username) {
				Eventually(authenticate.Username).Should(BeVisible())
				Expect(authenticate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
				Expect(authenticate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
				Expect(authenticate.Signin.Click()).To(Succeed())
			}

			if pages.ElementExist(authenticate.AuthCode) {
				Eventually(authenticate.AuthCode).Should(BeVisible())
				// Generate 6 digit authentication OTP for MFA
				authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
				Expect(authenticate.AuthCode.SendKeys(authCode)).To(Succeed())
			}
			Eventually(dexLogin.GrantAccess.Click).Should(Succeed())

		case GitProviderGitLab:
			Eventually(dexLogin.GitlabOnPrem.Click).Should(Succeed())

			authenticate := pages.AuthenticateWithGitlab(webDriver)
			if pages.ElementExist(authenticate.Username) {
				Eventually(authenticate.Username).Should(BeVisible())
				Expect(authenticate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
				Expect(authenticate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
				Expect(authenticate.Signin.Click()).To(Succeed())
			}

			Eventually(dexLogin.GrantAccess.Click).Should(Succeed())
		default:
			Expect(fmt.Errorf("error: Provided oidc issuer '%s' is not supported", gitProviderEnv.Type))
		}
	default:
		Expect(fmt.Errorf("error: Provided login type '%s' is not supported", userType))
	}

	Eventually(loginPage.AccountSettings.Click).Should(Succeed())
	account := pages.GetAccount(webDriver)
	Eventually(account.User.Click).Should(Succeed())
}

func logoutUser(authenticationType string) {
	loginPage := pages.GetLoginPage(webDriver)

	if pages.ElementExist(loginPage.AccountSettings) {
		Eventually(loginPage.AccountSettings.Click).Should(Succeed())
		account := pages.GetAccount(webDriver)
		Eventually(account.Logout.Click).Should(Succeed())
	}
}
