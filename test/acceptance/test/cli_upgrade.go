package acceptance

import (
	"fmt"
	"path"
	"regexp"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var _ = ginkgo.Describe("Gitops upgrade Tests", ginkgo.Label("cli", "upgrade"), func() {
	var upgradeMgmtClusterHostname string
	var upgradeWgeEndpointUrl string
	var upgradeTestUiUrl string
	const UI_NODEPORT = "30081"
	var stdOut string
	var stdErr string

	ginkgo.BeforeEach(func() {

	})

	ginkgo.Context("[CLI] When gitops upgrade command is available", func() {
		ginkgo.It("Verify gitops upgrade command in --dry-run mode", func() {
			repositoryURL := fmt.Sprintf(`https://%s/%s/%s`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)
			prBranch := "wego-enterprise-dry-run"
			version := "1.4.20"

			ginkgo.By("And I run gitops upgrade command", func() {
				upgradeCommand := fmt.Sprintf(" %s upgrade --version %s --branch %s --config-repo %s --path=./clusters/management/clusters --set 'service.nodePorts.https=%s' --set 'service.type=NodePort' --dry-run", gitopsBinPath, version, prBranch, repositoryURL, UI_NODEPORT)
				logger.Infof("Upgrade command: '%s'", upgradeCommand)
				stdOut, stdErr = runCommandAndReturnStringOutput(upgradeCommand)
				gomega.Expect(stdErr).Should(gomega.BeEmpty())
			})

			ginkgo.By("And verify kind 'HelmRepository' in upgrade manifest", func() {
				gomega.Expect(stdOut).Should(gomega.MatchRegexp(`kind: HelmRepository[\s\w\d./:-]*name: weave-gitops-enterprise-charts[\s\w\d./:-]*secretRef:[\s]*name: weave-gitops-enterprise-credentials[\s]*url: https://charts.dev.wkp.weave.works/releases/charts-v3`))
			})

			ginkgo.By("And verify kind 'HelmRelease' in upgrade manifest", func() {
				gomega.Expect(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*sourceRef:[\s]*kind: HelmRepository[\s]*name: weave-gitops-enterprise-charts[\s\w\d./:-]*version: %s`, version)))
				gomega.Expect(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d.@/:-]*service[\s\w\d.@/:-]*nodePorts:[\s]*https: %s`, UI_NODEPORT)))
				gomega.Expect(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*repositoryURL: %s`, repositoryURL)))
			})

			ginkgo.By("And verify kind 'Kustomization' in upgrade manifest", func() {
				gomega.Expect(stdOut).Should(gomega.MatchRegexp(`kind: Kustomization[\s\w\d./:-]*sourceRef:[\s]*kind: GitRepository[\s]*name: wego-`))
			})

		})
	})

	ginkgo.Context("[CLI] When Wego core is installed in the cluster", ginkgo.Label("cluster"), func() {
		var currentConfigRepo string
		var currentContext string
		clusterPath := "./clusters/management/clusters"
		kind_upgrade_cluster_name := "test-upgrade"

		ginkgo.JustBeforeEach(func() {
			currentContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			currentConfigRepo = gitProviderEnv.Repo
			gitProviderEnv.Repo = "upgrade-" + currentConfigRepo

			upgradeMgmtClusterHostname = GetEnv("UPGRADE_MANAGEMENT_CLUSTER_CNAME", "localhost")
			upgradeWgeEndpointUrl = fmt.Sprintf(`https://%s:%s`, upgradeMgmtClusterHostname, UI_NODEPORT)
			upgradeTestUiUrl = fmt.Sprintf(`https://%s:%s`, upgradeMgmtClusterHostname, UI_NODEPORT)

			// Create vanilla cluster for WGE upgrade
			createCluster("kind", kind_upgrade_cluster_name, "kind/extra-port-mapping-kind-config.yaml")

		})

		ginkgo.JustAfterEach(func() {
			deleteRepo(gitProviderEnv)              // Delete the upgrade config repository to keep the org clean
			gitProviderEnv.Repo = currentConfigRepo // Revert to original config repository for subsequent tests
			err := runCommandPassThrough("kubectl", "config", "use-context", currentContext)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

			deleteCluster("kind", kind_upgrade_cluster_name, "")

			// Login to management cluster console, in case it has been logged out
			initializeWebdriver(testUiUrl)
			loginUser()

		})

		ginkgo.It("Verify wego core can be upgraded to wego enterprise", ginkgo.Label("kind"), func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

			ginkgo.By("When I create a private repository for cluster configs", func() {
				initAndCreateEmptyRepo(gitProviderEnv, true)
			})

			ginkgo.By("When I install gitops/wego to my active cluster", func() {
				logger.Info("Bootstrapping the cluster to install flux")
				bootstrapAndVerifyFlux(gitProviderEnv, GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))

				logger.Info("Installing Weave gitops...")
				_ = runCommandPassThrough("sh", "-c", "helm repo add ww-gitops https://helm.gitops.weave.works && helm repo update")
				wegoInstallCmd := fmt.Sprintf("helm install weave-gitops ww-gitops/weave-gitops --namespace %s --set 'adminUser.create=true' --set 'adminUser.username=%s' --set 'adminUser.passwordHash=%s'", GITOPS_DEFAULT_NAMESPACE, AdminUserName, GetEnv("CLUSTER_ADMIN_PASSWORD_HASH", ""))
				logger.Info("Weave gitops install command: " + wegoInstallCmd)
				_ = runCommandPassThrough("sh", "-c", wegoInstallCmd)

				verifyCoreControllers(GITOPS_DEFAULT_NAMESPACE)
			})

			ginkgo.By("And I install Profile (HelmRepository chart)", func() {
				sourceURL := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				addSource("helm", "weaveworks-charts", GITOPS_DEFAULT_NAMESPACE, sourceURL, "", "")
			})

			ginkgo.By("And I install the entitlement for cluster upgrade", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "entitlement/entitlement-secret.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("And secure access to dashboard for dex users", func() {
				logger.Info("Create client credential secret for OIDC (dex)")
				_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl create secret generic client-credentials --namespace %s --from-literal=clientID=%s --from-literal=clientSecret=%s", GITOPS_DEFAULT_NAMESPACE, GetEnv("DEX_CLIENT_ID", ""), GetEnv("DEX_CLIENT_SECRET", "")))
			})

			ginkgo.By("And I install the git repository secret for cluster service", func() {
				var cmd string
				switch gitProviderEnv.Type {
				case GitProviderGitHub:
					cmd = fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"`, GITOPS_DEFAULT_NAMESPACE, gitProviderEnv.Token)
				case GitProviderGitLab:
					if gitProviderEnv.Hostname == gitlab.DefaultDomain {
						cmd = fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"  --from-literal="GITLAB_CLIENT_ID=%s" --from-literal="GITLAB_CLIENT_SECRET=%s"`,
							GITOPS_DEFAULT_NAMESPACE, gitProviderEnv.Token, gitProviderEnv.ClientId, gitProviderEnv.ClientSecret)
					} else {
						stdOut, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`ssh-keyscan %s > known_hosts`, gitProviderEnv.Hostname) + " && " + fmt.Sprintf(`kubectl create configmap ssh-config --namespace %s --from-file=./known_hosts`, GITOPS_DEFAULT_NAMESPACE))
						gomega.Expect(stdOut).Should(gomega.MatchRegexp(`configmap/ssh-config created`), "Failed to create on-prem known hosts 'ssh-config' configmap")

						cmd = fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"  --from-literal="GITLAB_CLIENT_ID=%s" --from-literal="GITLAB_CLIENT_SECRET=%s"  --from-literal="GITLAB_HOSTNAME=%s" --from-literal="GIT_HOST_TYPES=%s"`,
							GITOPS_DEFAULT_NAMESPACE, gitProviderEnv.Token, gitProviderEnv.ClientId, gitProviderEnv.ClientSecret, gitProviderEnv.GitlabHostname, gitProviderEnv.HostTypes)
					}
				}
				stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to create git repository secret for cluster service")
			})

			ginkgo.By("And install cert-manager for tls certificate creation", func() {
				_ = runCommandPassThrough("sh", "-c", "helm upgrade --install cert-manager cert-manager/cert-manager --namespace cert-manager --create-namespace --version v1.10.0 --wait --set installCRDs=true")
				_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cat %s | sed s,{{HOST_NAME}},"%s",g | kubectl apply -f -`, path.Join(testDataPath, "ingress/certificate-issuer.yaml"), upgradeMgmtClusterHostname))
				_ = runCommandPassThrough("sh", "-c", "kubectl wait --for=condition=Ready --timeout=60s -n flux-system --all certificate")

			})

			ginkgo.By("And install ingress-nginx for tls termination", func() {
				command := fmt.Sprintf("helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace --version 4.4.0 --wait --set controller.service.type=NodePort --set controller.service.nodePorts.https=%s --set controller.extraArgs.v=4", UI_NODEPORT)
				_ = runCommandPassThrough("sh", "-c", command)
				_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cat %s | sed s,{{HOST_NAME}},"%s",g | kubectl apply -f -`, path.Join(testDataPath, "ingress/ingress.yaml"), upgradeMgmtClusterHostname))
			})

			prBranch := "wego-upgrade-enterprise"
			version := "0.13.0"
			ginkgo.By(fmt.Sprintf("And I run gitops upgrade command from directory %s", repoAbsolutePath), func() {
				gitRepositoryURL := fmt.Sprintf(`https://%s/%s/%s`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)
				// Explicitly setting the gitprovider type, hostname and repository path url scheme in configmap, the default is github and ssh url scheme which is not supported for capi cluster PR creation.
				upgradeCommand := fmt.Sprintf("upgrade --version %s --branch %s --config-repo %s --path=%s --set 'config.capi.repositoryClustersPath=./clusters'  --set 'config.capi.repositoryURL=%s' --set 'config.git.type=%s' --set 'config.git.hostname=%s' --set tls.enabled=false --set config.oidc.enabled=true --set config.oidc.clientCredentialsSecret=client-credentials --set config.oidc.issuerURL=https://dex-01.wge.dev.weave.works --set config.oidc.redirectURL=https://weave.gitops.upgrade.enterprise.com:%s/oauth2/callback ",
					version, prBranch, gitRepositoryURL, clusterPath, gitRepositoryURL, gitProviderEnv.Type, gitProviderEnv.Hostname, UI_NODEPORT)

				if gitProviderEnv.HostTypes != "" {
					upgradeCommand += ` --set "config.extraVolumes[0].name=ssh-config" --set "config.extraVolumes[0].configMap.name=ssh-config" --set "config.extraVolumeMounts[0].name=ssh-config" --set "config.extraVolumeMounts[0].mountPath=/root/.ssh"`
				}

				logger.Infof("Upgrade command: '%s'", upgradeCommand)
				// stdOut, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && %s", repoAbsolutePath, upgradeCommand), ASSERTION_1MINUTE_TIME_OUT)
				stdOut, stdErr = runGitopsCommand(upgradeCommand)
				gomega.Expect(stdErr).Should(gomega.BeEmpty())
			})

			ginkgo.By("Then I should see pull request created to management cluster", func() {
				re := regexp.MustCompile(`Pull Request created.*:[\s\w\d]+(?P<URL>https.*\/\d+)`)
				match := re.FindSubmatch([]byte(stdOut))
				gomega.Eventually(match[1]).ShouldNot(gomega.BeNil(), "Failed to Create pull request")
			})

			ginkgo.By("Then I should merge the pull request to start weave gitops enterprise upgrade", func() {
				upgradePRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, upgradePRUrl)
			})

			ginkgo.By("And I should see cluster upgraded from 'wego core' to 'wego enterprise'", func() {
				verifyEnterpriseControllers("weave-gitops-enterprise", "mccp-", GITOPS_DEFAULT_NAMESPACE)
			})

			ginkgo.By("And I should install rolebindings for default enterprise roles'", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "rbac/user-role-bindings.yaml")), "Failed to install rolbindings for enterprise default roles")
			})

			initializeWebdriver(upgradeTestUiUrl) // Initilize web driver for whole test suite run

			ginkgo.By(fmt.Sprintf("Login as a %s user", userCredentials.UserType), func() {
				loginUser() // Login to the weaveworks enterprise dashboard
			})

			ginkgo.By("And the Cluster service is healthy", func() {
				checkClusterService(upgradeWgeEndpointUrl)
			})

			ginkgo.By("Then I should run enterprise CLI commands", func() {
				testGetCommand := func(subCommand, expectedOutput string) {
					// Using self signed certs, all `gitops get clusters` etc commands should use insecure tls connections
					insecureFlag := "--insecure-skip-tls-verify"
					// Login via cluster user account (basic authentication). Kind cluster doesn't support CLI OIDC authentication
					authFlag := fmt.Sprintf("--username %s --password %s", userCredentials.ClusterUserName, userCredentials.ClusterUserPassword)

					cmd := fmt.Sprintf(`%s --endpoint %s %s %s get %s`, gitopsBinPath, upgradeWgeEndpointUrl, insecureFlag, authFlag, subCommand)
					ginkgo.By(fmt.Sprintf(`And I run '%s'`, cmd), func() {
						stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
						gomega.Expect(stdErr).Should(gomega.BeEmpty(), fmt.Sprintf("'%s get %s' command failed", gitopsBinPath, subCommand))
						gomega.Expect(stdOut).Should(gomega.MatchRegexp(expectedOutput), fmt.Sprintf("'%s get %s' command failed", gitopsBinPath, subCommand))
					})
				}

				testGetCommand("templates", "No templates were found")
				testGetCommand("credentials", "No credentials were found")
				testGetCommand("clusters", `kind-test-upgrade\s+Ready`)
			})

			// Namespace for some GitOpsTemplates
			existingSourceCount := getSourceCount()
			templateNamespaces := []string{"dev-system"}
			createNamespace(templateNamespaces)

			templateName := "helm-repository-template"
			templateFiles := map[string]string{
				"helm-repository-template": path.Join(testDataPath, "templates/source/helm-repository-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, templateName)
				gomega.Eventually(templateRow.CreateTemplate.Click, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			sourceName := "bitnami"
			sourceNamespace := "dev-system"
			resourceUrl := "https://charts.bitnami.com/bitnami"

			var parameters = []TemplateField{
				{
					Name:   "RESOURCE_NAME",
					Value:  sourceName,
					Option: "",
				},
				{
					Name:   "NAMESPACE",
					Value:  sourceNamespace,
					Option: "",
				},
				{
					Name:   "URL",
					Value:  resourceUrl,
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify PR preview contents for templating functions", func() {
				// Verify resource definition preview
				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				// Verify resource is labelled with template name and namespace
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join(clusterPath, sourceNamespace, sourceName+".yaml")))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRepository[\s]*metadata:[/"\s\w\d./:-]*name: %s[\s]*namespace: %s`, sourceName, sourceNamespace)))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`url:\s+%s`, resourceUrl)))

				// Verify profiles and kustomization tab views are disabled because no profiles and kustomizations are part of pull request
				gomega.Expect(preview.GetPreviewTab("Profiles").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'PROFILES' preview tab should be disabled")
				gomega.Expect(preview.GetPreviewTab("Kustomizations").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'KUSTOMIZATIONS' preview tab should be disabled")

				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")
			})

			_ = createGitopsPR(PullRequest{})

			ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start cluster provisioning", func() {
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			pages.NavigateToPage(webDriver, "Sources")
			sourcePage := pages.GetSourcesPage(webDriver)

			ginkgo.By("And wait for bitnami source to be visibe on the dashboard", func() {
				gomega.Eventually(sourcePage.SourceHeader).Should(matchers.BeVisible())

				totalSourceCount := existingSourceCount + 1
				gomega.Eventually(sourcePage.CountSources, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.Equal(totalSourceCount), fmt.Sprintf("There should be %d sources enteries in source table", totalSourceCount))
			})

			ginkgo.By(fmt.Sprintf("And verify %s Sources page", sourceName), func() {
				sourceInfo := sourcePage.FindSourceInList(sourceName)
				gomega.Eventually(sourceInfo.Name).Should(matchers.MatchText(sourceName), fmt.Sprintf("Failed to list %s source in  source table", sourceName))
				gomega.Eventually(sourceInfo.Namespace).Should(matchers.MatchText(sourceNamespace), fmt.Sprintf("Failed to have expected %s source namespace: %s", sourceName, sourceNamespace))
				gomega.Eventually(sourceInfo.Cluster).Should(matchers.MatchText(kind_upgrade_cluster_name), fmt.Sprintf("Failed to have expected %s source cluster: %s", sourceName, kind_upgrade_cluster_name))
				gomega.Eventually(sourceInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("Ready"), fmt.Sprintf("Failed to have expected %s source status: Ready", sourceName))
				gomega.Eventually(sourceInfo.Url).Should(matchers.MatchText(resourceUrl), fmt.Sprintf("Failed to have expected %s source url: %s", sourceName, resourceUrl))
			})
		})
	})
})
