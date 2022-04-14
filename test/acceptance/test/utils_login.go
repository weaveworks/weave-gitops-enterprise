package acceptance

import (
	"fmt"
	"strings"

	"github.com/fluxcd/go-git-providers/gitlab"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
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
			Expect(loginPage.Username.SendKeys(userCredentials.UserName)).To(Succeed())
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

func AuthenticateWithGitProvider(webDriver *agouti.Page, gitProvider, gitProviderHostname string) {
	if gitProvider == GitProviderGitHub {
		authenticate := pages.AuthenticateWithGithub(webDriver)

		if pages.ElementExist(authenticate.AuthenticateGithub) {
			Expect(authenticate.AuthenticateGithub.Click()).To(Succeed())
			AuthenticateWithGitHub(webDriver)

			// Sometimes authentication failed to get the github device code, it may require revalidation with new access code
			if pages.ElementExist(authenticate.AuthorizationError) {
				logger.Info("Error getting github device code, requires revalidating...")
				Expect(authenticate.Close.Click()).To(Succeed())
				Eventually(authenticate.AuthenticateGithub.Click).Should(Succeed())
				AuthenticateWithGitHub(webDriver)
			}

			Eventually(authenticate.AuthroizeButton).ShouldNot(BeFound())
		}
	} else if gitProvider == GitProviderGitLab {
		var authenticate *pages.AuthenticateGitlab
		if gitProviderHostname == gitlab.DefaultDomain {
			authenticate = pages.AuthenticateWithGitlab(webDriver)
		} else {
			authenticate = pages.AuthenticateWithOnPremGitlab(webDriver)
		}

		if pages.ElementExist(authenticate.AuthenticateGitlab) {
			Expect(authenticate.AuthenticateGitlab.Click()).To(Succeed())

			if !pages.ElementExist(authenticate.Username) {
				if pages.ElementExist(authenticate.CheckBrowser) {
					setGitlabBrowserCompatibility(webDriver)
					Eventually(authenticate.CheckBrowser, ASSERTION_30SECONDS_TIME_OUT).ShouldNot(BeFound())
					TakeScreenShot("gitlab_browser_compatibility")
				}

				if pages.ElementExist(authenticate.AcceptCookies, 10) {
					Eventually(authenticate.AcceptCookies.Click).Should(Succeed())
				}
			}

			TakeScreenShot("gitlab_cookies_accepted")
			if pages.ElementExist(authenticate.Username) {
				Eventually(authenticate.Username).Should(BeVisible())
				Expect(authenticate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
				Expect(authenticate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
				Expect(authenticate.Signin.Submit()).To(Succeed())
			} else {
				logger.Info("Login not found, assuming already logged in")
			}

			if pages.ElementExist(authenticate.Authorize) {
				Expect(authenticate.Authorize.Click()).To(Succeed())
			}
		}
	}
}

func setGitlabBrowserCompatibility(webDriver *agouti.Page) {
	// opening the gitlab in a separate window not controlled by webdriver seems to redirect gitlab to login
	pages.OpenNewWindow(webDriver, `http://`+gitProviderEnv.Hostname+`/users/sign_in`, "gitlab")
	// Make sure weave-gitops-enterprise application window is still active window
	Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to wego application window")
}

func AuthenticateWithGitHub(webDriver *agouti.Page) {

	authenticate := pages.AuthenticateWithGithub(webDriver)

	Eventually(authenticate.AccessCode).Should(BeVisible())
	accessCode, _ := authenticate.AccessCode.Text()
	Expect(authenticate.AuthroizeButton.Click()).To(Succeed())
	accessCode = strings.Replace(accessCode, "-", "", 1)
	logger.Info(accessCode)

	// Move to device activation window
	TakeScreenShot("application_authentication")
	Expect(webDriver.NextWindow()).ShouldNot(HaveOccurred(), "Failed to switch to github authentication window")
	TakeScreenShot("github_authentication")

	activate := pages.ActivateDeviceGithub(webDriver)

	if pages.ElementExist(activate.Username) {
		Eventually(activate.Username).Should(BeVisible())
		Expect(activate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
		Expect(activate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
		Expect(activate.Signin.Click()).To(Succeed())
	} else {
		logger.Info("Login not found, assuming already logged in")
		TakeScreenShot("login_skipped")
	}

	if pages.ElementExist(activate.AuthCode) {
		Eventually(activate.AuthCode).Should(BeVisible())
		// Generate 6 digit authentication OTP for MFA
		authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
		Expect(activate.AuthCode.SendKeys(authCode)).To(Succeed())
	} else {
		logger.Info("OTP not found, assuming already logged in")
		TakeScreenShot("otp_skipped")
	}

	Eventually(activate.Continue).Should(BeVisible())
	Expect(activate.UserCode.At(0).SendKeys(accessCode)).To(Succeed())
	Expect(activate.Continue.Click()).To(Succeed())

	Eventually(activate.AuthroizeWeaveworks).Should(BeEnabled())
	Expect(activate.AuthroizeWeaveworks.Click()).To(Succeed())

	Eventually(activate.ConnectedMessage).Should(BeVisible())
	Expect(webDriver.CloseWindow()).ShouldNot(HaveOccurred())

	// Device is connected, now move back to application window
	Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to wego application window")
}
