package acceptance

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/gitlab"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

const (
	AdminUserName    = "admin"
	ClusterUserLogin = "cluster-user"
	OidcUserLogin    = "oidc"
)

type UserCredentials struct {
	UserType     string
	UserName     string
	UserPassword string
}

func initUserCredentials() UserCredentials {
	userCredentials := UserCredentials{
		UserType:     GetEnv("LOGIN_USER_TYPE", ClusterUserLogin),
		UserName:     AdminUserName,
		UserPassword: GetEnv("CLUSTER_ADMIN_PASSWORD", ""),
	}

	if userCredentials.UserType == OidcUserLogin {
		userCredentials.UserName = gitProviderEnv.Username
		userCredentials.UserPassword = gitProviderEnv.Password
	}
	return userCredentials
}

func loginUser() {
	loginPage := pages.GetLoginPage(webDriver)

	if pages.ElementExist(loginPage.LoginOIDC, 10) {
		Eventually(loginPage.LoginOIDC).Should(BeVisible())

		switch userCredentials.UserType {
		case ClusterUserLogin:
			// Login via cluster user account
			Expect(loginPage.Password.SendKeys(userCredentials.UserPassword)).To(Succeed())
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
					Expect(authenticate.Username.SendKeys(userCredentials.UserName)).To(Succeed())
					Expect(authenticate.Password.SendKeys(userCredentials.UserPassword)).To(Succeed())
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

				var authenticate *pages.AuthenticateGitlab
				if gitProviderEnv.Hostname == gitlab.DefaultDomain {
					authenticate = pages.AuthenticateWithGitlab(webDriver)
				} else {
					authenticate = pages.AuthenticateWithOnPremGitlab(webDriver)
				}
				if pages.ElementExist(authenticate.Username) {
					Eventually(authenticate.Username).Should(BeVisible())
					Expect(authenticate.Username.SendKeys(userCredentials.UserName)).To(Succeed())
					Expect(authenticate.Password.SendKeys(userCredentials.UserPassword)).To(Succeed())
					Expect(authenticate.Signin.Click()).To(Succeed())
				}
				Eventually(dexLogin.GrantAccess.Click).Should(Succeed())
			default:
				Expect(fmt.Errorf("error: Provided oidc issuer '%s' is not supported", gitProviderEnv.Type))
			}
		default:
			Expect(fmt.Errorf("error: Provided login type '%s' is not supported", userCredentials.UserType))
		}

		Eventually(loginPage.AccountSettings.Click).Should(Succeed())
		account := pages.GetAccount(webDriver)
		Eventually(account.User.Click).Should(Succeed())
	} else {
		logger.Infof("%s user already logged in", userCredentials.UserType)
	}
}

func logoutUser() {
	loginPage := pages.GetLoginPage(webDriver)

	if pages.ElementExist(loginPage.AccountSettings, 10) {
		Eventually(loginPage.AccountSettings.Click).Should(Succeed())
		account := pages.GetAccount(webDriver)
		Eventually(account.Logout.Click).Should(Succeed())
	}
}
