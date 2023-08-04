package acceptance

import (
	"fmt"
	"strings"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

const (
	AdminUserName    = "wego-admin"
	ClusterUserLogin = "cluster-user"
	OidcUserLogin    = "oidc"
)

type UserCredentials struct {
	UserType            string
	UserName            string
	UserPassword        string
	UserKubeconfig      string
	ClusterUserName     string
	ClusterUserPassword string
}

func initUserCredentials() UserCredentials {
	userCredentials := UserCredentials{
		UserType:            GetEnv("LOGIN_USER_TYPE", ClusterUserLogin),
		UserName:            AdminUserName,
		UserPassword:        GetEnv("CLUSTER_ADMIN_PASSWORD", "dev"),
		UserKubeconfig:      GetEnv("OIDC_KUBECONFIG", ""),
		ClusterUserName:     AdminUserName,
		ClusterUserPassword: GetEnv("CLUSTER_ADMIN_PASSWORD", "dev"),
	}

	if userCredentials.UserType == OidcUserLogin {
		userCredentials.UserName = gitProviderEnv.Username
		userCredentials.UserPassword = gitProviderEnv.Password
	}
	return userCredentials
}

func loginUser() {
	loginUserFlow(userCredentials)
}

func loginUserFlow(uc UserCredentials) {
	loginPage := pages.GetLoginPage(webDriver)

	if pages.ElementExist(loginPage.LoginOIDC, 10) {
		gomega.Eventually(loginPage.LoginOIDC).Should(matchers.BeVisible())

		switch uc.UserType {
		case ClusterUserLogin:
			// Login via cluster user account
			gomega.Expect(loginPage.Username.SendKeys(uc.UserName)).To(gomega.Succeed())
			gomega.Expect(loginPage.Password.SendKeys(uc.UserPassword)).To(gomega.Succeed())
			gomega.Expect(loginPage.Continue.Click()).To(gomega.Succeed())
		case OidcUserLogin:
			// Login via OIDC provider
			gomega.Eventually(loginPage.LoginOIDC.Click).Should(gomega.Succeed())

			dexLogin := pages.GetDexLoginPage(webDriver)
			switch gitProviderEnv.Type {
			case GitProviderGitHub:
				gomega.Eventually(dexLogin.Github.Click).Should(gomega.Succeed())

				authenticate := pages.ActivateDeviceGithub(webDriver)

				if pages.ElementExist(authenticate.Username) {
					gomega.Eventually(authenticate.Username).Should(matchers.BeVisible())
					gomega.Expect(authenticate.Username.SendKeys(uc.UserName)).To(gomega.Succeed())
					gomega.Expect(authenticate.Password.SendKeys(uc.UserPassword)).To(gomega.Succeed())
					gomega.Expect(authenticate.Signin.Click()).To(gomega.Succeed())
				}

				if pages.ElementExist(authenticate.AuthCode) {
					gomega.Eventually(authenticate.AuthCode).Should(matchers.BeVisible())
					// Generate 6 digit authentication OTP for MFA
					authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
					gomega.Expect(authenticate.AuthCode.SendKeys(authCode)).To(gomega.Succeed())
				}
				gomega.Eventually(dexLogin.GrantAccess.Click).Should(gomega.Succeed())

			case GitProviderGitLab:
				gomega.Eventually(dexLogin.GitlabOnPrem.Click).Should(gomega.Succeed())

				var authenticate *pages.AuthenticateGitlab
				if gitProviderEnv.Hostname == gitlab.DefaultDomain {
					authenticate = pages.AuthenticateWithGitlab(webDriver)
				} else {
					authenticate = pages.AuthenticateWithOnPremGitlab(webDriver)
				}
				if pages.ElementExist(authenticate.Username) {
					gomega.Eventually(authenticate.Username).Should(matchers.BeVisible())
					gomega.Expect(authenticate.Username.SendKeys(uc.UserName)).To(gomega.Succeed())
					gomega.Expect(authenticate.Password.SendKeys(uc.UserPassword)).To(gomega.Succeed())
					gomega.Expect(authenticate.Signin.Click()).To(gomega.Succeed())
				}
				gomega.Eventually(dexLogin.GrantAccess.Click).Should(gomega.Succeed())
			default:
				gomega.Expect(fmt.Errorf("error: Provided oidc issuer '%s' is not supported", gitProviderEnv.Type))
			}
		default:
			gomega.Expect(fmt.Errorf("error: Provided login type '%s' is not supported", uc.UserType))
		}

		gomega.Eventually(loginPage.AccountSettings.Click).Should(gomega.Succeed())
		account := pages.GetAccount(webDriver)
		gomega.Eventually(account.User.Click).Should(gomega.Succeed())
	} else {
		logger.Infof("%s user already logged in", uc.UserType)
	}
}

func logoutUser() {
	loginPage := pages.GetLoginPage(webDriver)

	if pages.ElementExist(loginPage.AccountSettings, 5) {
		gomega.Eventually(loginPage.AccountSettings.Click).Should(gomega.Succeed())
		account := pages.GetAccount(webDriver)
		gomega.Eventually(account.Logout.Click).Should(gomega.Succeed())
		gomega.Eventually(loginPage.LoginOIDC).Should(matchers.BeVisible(), "Failed to logout")
	}
}

func cliOidcLogin() {
	switch mgmtClusterKind {
	case EKSMgmtCluster:
		go func() {
			_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl get pods --kubeconfig=%s", userCredentials.UserKubeconfig))
		}()

		redirectUrl := "http://localhost:8000"
		pages.OpenWindowInBg(webDriver, redirectUrl, "cli-oidc-auth")
		gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), "Failed to switch to 'cli-oidc-auth' window")
		cliOidcAuthFlow(userCredentials)
		gomega.Eventually(webDriver.CloseWindow).Should(gomega.Succeed(), "Failed to close 'cli-oidc-auth' dashboard window")
		gomega.Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to weave gitops enterprise dashboard")

	case GKEMgmtCluster:
		logger.Info("GKE cli oidc auth is not implemented yet")
		// kubectl oidc login --cluster=CLUSTER_NAME --login-config=client-config.yaml --kubeconfig=${OIDC_KUBECONFIG}
	}
}

func cliOidcAuthFlow(uc UserCredentials) {
	dexLogin := pages.GetDexLoginPage(webDriver)

	if pages.ElementExist(dexLogin.Title, 5) {
		gomega.Eventually(dexLogin.Title).Should(matchers.BeVisible())

		switch gitProviderEnv.Type {
		case GitProviderGitHub:
			gomega.Eventually(dexLogin.Github.Click).Should(gomega.Succeed())

			authenticate := pages.ActivateDeviceGithub(webDriver)

			if pages.ElementExist(authenticate.Username) {
				gomega.Eventually(authenticate.Username).Should(matchers.BeVisible())
				gomega.Expect(authenticate.Username.SendKeys(uc.UserName)).To(gomega.Succeed())
				gomega.Expect(authenticate.Password.SendKeys(uc.UserPassword)).To(gomega.Succeed())
				gomega.Expect(authenticate.Signin.Click()).To(gomega.Succeed())
			}

			if pages.ElementExist(authenticate.AuthCode) {
				gomega.Eventually(authenticate.AuthCode).Should(matchers.BeVisible())
				// Generate 6 digit authentication OTP for MFA
				authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
				gomega.Expect(authenticate.AuthCode.SendKeys(authCode)).To(gomega.Succeed())
			}
			gomega.Eventually(dexLogin.GrantAccess.Click).Should(gomega.Succeed())

		case GitProviderGitLab:
			gomega.Eventually(dexLogin.GitlabOnPrem.Click).Should(gomega.Succeed())

			var authenticate *pages.AuthenticateGitlab
			if gitProviderEnv.Hostname == gitlab.DefaultDomain {
				authenticate = pages.AuthenticateWithGitlab(webDriver)
			} else {
				authenticate = pages.AuthenticateWithOnPremGitlab(webDriver)
			}
			if pages.ElementExist(authenticate.Username) {
				gomega.Eventually(authenticate.Username).Should(matchers.BeVisible())
				gomega.Expect(authenticate.Username.SendKeys(uc.UserName)).To(gomega.Succeed())
				gomega.Expect(authenticate.Password.SendKeys(uc.UserPassword)).To(gomega.Succeed())
				gomega.Expect(authenticate.Signin.Click()).To(gomega.Succeed())
			}
			gomega.Eventually(dexLogin.GrantAccess.Click).Should(gomega.Succeed())
		default:
			gomega.Expect(fmt.Errorf("error: Provided oidc issuer '%s' is not supported", gitProviderEnv.Type))
		}
	} else {
		logger.Infof("%s user already logged in", uc.UserType)
	}
}

func authenticateWithGitProvider(webDriver *agouti.Page, gitProvider, gitProviderHostname string) {
	if gitProvider == GitProviderGitHub {
		authenticate := pages.AuthenticateWithGithub(webDriver)

		if pages.ElementExist(authenticate.AuthenticateGithub) {
			gomega.Expect(authenticate.AuthenticateGithub.Click()).To(gomega.Succeed())
			authenticateWithGitHub(webDriver)

			// Sometimes authentication failed to get the github device code, it may require revalidation with new access code
			if pages.ElementExist(authenticate.AuthorizationError) {
				logger.Info("Error getting github device code, requires revalidating...")
				gomega.Expect(authenticate.Close.Click()).To(gomega.Succeed())
				gomega.Eventually(authenticate.AuthenticateGithub.Click).Should(gomega.Succeed())
				authenticateWithGitHub(webDriver)
			}

			gomega.Eventually(authenticate.AuthroizeButton).ShouldNot(matchers.BeFound())
		}
	} else if gitProvider == GitProviderGitLab {
		var authenticate *pages.AuthenticateGitlab
		if gitProviderHostname == gitlab.DefaultDomain {
			authenticate = pages.AuthenticateWithGitlab(webDriver)
		} else {
			authenticate = pages.AuthenticateWithOnPremGitlab(webDriver)
		}

		if pages.ElementExist(authenticate.AuthenticateGitlab) {
			gomega.Expect(authenticate.AuthenticateGitlab.Click()).To(gomega.Succeed())

			browserCompatibility := false
			if !pages.ElementExist(authenticate.Username) {
				logger.Info("Username field not found, checking for browser compatibility...")
				if pages.ElementExist(authenticate.CheckBrowser) {
					logger.Info("Found browser compatibility check, opening gitlab in a separate window...")
					// Opening the gitlab in a separate window not controlled by webdriver redirects gitlab to login page (DDOS workaround for gitlab)
					pages.OpenWindowInBg(webDriver, `http://`+gitProviderEnv.Hostname+`/users/sign_in`, "gitlab")
					gomega.Eventually(authenticate.CheckBrowser, ASSERTION_30SECONDS_TIME_OUT).ShouldNot(matchers.BeFound())
					browserCompatibility = true
				} else {
					logger.Info("Browser compatibility check not found")
				}

				if pages.ElementExist(authenticate.AcceptCookies, 10) {
					gomega.Eventually(authenticate.AcceptCookies.Click).Should(gomega.Succeed())
				}
			} else {
				logger.Info("Username field found")
			}

			if pages.ElementExist(authenticate.Username) {
				logger.Info("Username field re-found")
				gomega.Eventually(authenticate.Username).Should(matchers.BeVisible())
				gomega.Expect(authenticate.Username.SendKeys(gitProviderEnv.Username)).To(gomega.Succeed())
				gomega.Expect(authenticate.Password.SendKeys(gitProviderEnv.Password)).To(gomega.Succeed())
				gomega.Expect(authenticate.Signin.Submit()).To(gomega.Succeed())

			} else {
				logger.Info("Login not found, assuming already logged in")
			}

			if pages.ElementExist(authenticate.Authorize) {
				gomega.Expect(authenticate.Authorize.Click()).To(gomega.Succeed())
			}

			if browserCompatibility {
				pages.CloseWindow(webDriver, pages.GetNextWindow(webDriver))
			}
		}
	}
}

func authenticateWithGitHub(webDriver *agouti.Page) {

	authenticate := pages.AuthenticateWithGithub(webDriver)

	gomega.Eventually(authenticate.AccessCode).Should(matchers.BeVisible())
	accessCode, _ := authenticate.AccessCode.Text()
	gomega.Expect(authenticate.AuthroizeButton.Click()).To(gomega.Succeed())
	accessCode = strings.Replace(accessCode, "-", "", 1)
	logger.Info(accessCode)

	// Move to device activation window
	gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), "Failed to switch to github authentication window")

	activate := pages.ActivateDeviceGithub(webDriver)

	if pages.ElementExist(activate.Username) {
		gomega.Eventually(activate.Username).Should(matchers.BeVisible())
		gomega.Expect(activate.Username.SendKeys(gitProviderEnv.Username)).To(gomega.Succeed())
		gomega.Expect(activate.Password.SendKeys(gitProviderEnv.Password)).To(gomega.Succeed())
		gomega.Expect(activate.Signin.Click()).To(gomega.Succeed())
	} else {
		logger.Info("Login not found, assuming already logged in")
		takeScreenShot("login_skipped")
	}

	if pages.ElementExist(activate.AuthCode) {
		gomega.Eventually(activate.AuthCode).Should(matchers.BeVisible())
		// Generate 6 digit authentication OTP for MFA
		authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
		gomega.Expect(activate.AuthCode.SendKeys(authCode)).To(gomega.Succeed())
	} else {
		logger.Info("OTP not found, assuming already logged in")
		takeScreenShot("otp_skipped")
	}

	gomega.Eventually(activate.Continue).Should(matchers.BeVisible())
	gomega.Expect(activate.UserCode.At(0).SendKeys(accessCode)).To(gomega.Succeed())
	gomega.Expect(activate.Continue.Click()).To(gomega.Succeed())

	gomega.Eventually(activate.AuthroizeWeaveworks).Should(matchers.BeEnabled())
	gomega.Expect(activate.AuthroizeWeaveworks.Click()).To(gomega.Succeed())

	gomega.Eventually(activate.ConnectedMessage).Should(matchers.BeVisible())
	gomega.Expect(webDriver.CloseWindow()).ShouldNot(gomega.HaveOccurred())

	// Device is connected, now move back to application window
	gomega.Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to wego application window")
}
