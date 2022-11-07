package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func DescribeCliUpgrade(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Gitops upgrade Tests", func() {

		UI_NODEPORT := "30081"
		var upgrade_capi_endpoint_url string
		var upgrade_test_ui_url string
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
					upgradeCommand := fmt.Sprintf(" %s upgrade --version %s --branch %s --config-repo %s --path=./clusters/management/clusters --set 'service.nodePorts.https=%s' --set 'service.type=NodePort' --dry-run", gitops_bin_path, version, prBranch, repositoryURL, UI_NODEPORT)
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

		ginkgo.Context("[CLI] When Wego core is installed in the cluster", func() {
			var currentConfigRepo string
			var currentContext string
			kind_upgrade_cluster_name := "test-upgrade"

			templateFiles := []string{}

			ginkgo.JustBeforeEach(func() {
				currentContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				currentConfigRepo = gitProviderEnv.Repo
				gitProviderEnv.Repo = "upgrade-" + currentConfigRepo

				// Create vanilla cluster for WGE upgrade
				createCluster("kind", kind_upgrade_cluster_name, "upgrade-kind-config.yaml")

			})

			ginkgo.JustAfterEach(func() {

				gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
				templateFiles = []string{}

				deleteRepo(gitProviderEnv)              // Delete the upgrade config repository to keep the org clean
				gitProviderEnv.Repo = currentConfigRepo // Revert to original config repository for subsequent tests
				err := runCommandPassThrough("kubectl", "config", "use-context", currentContext)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

				deleteCluster("kind", kind_upgrade_cluster_name, "")

				// Login to management cluster console, in case it has been logged out
				InitializeWebdriver(test_ui_url)
				loginUser()

			})

			ginkgo.It("Verify wego core can be upgraded to wego enterprise", ginkgo.Label("upgrade", "git"), func() {
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
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to create/configure entitlement")
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

				prBranch := "wego-upgrade-enterprise"
				version := "0.9.1"
				ginkgo.By(fmt.Sprintf("And I run gitops upgrade command from directory %s", repoAbsolutePath), func() {
					gitRepositoryURL := fmt.Sprintf(`https://%s/%s/%s`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)
					// Explicitly setting the gitprovider type, hostname and repository path url scheme in configmap, the default is github and ssh url scheme which is not supported for capi cluster PR creation.
					upgradeCommand := fmt.Sprintf(" %s upgrade --version %s --branch %s --config-repo %s --path=./clusters/management/clusters  --set 'config.capi.repositoryPath=./clusters/capi/clusters' --set 'config.capi.repositoryClustersPath=./clusters'  --set 'config.capi.repositoryURL=%s' --set 'config.git.type=%s' --set 'config.git.hostname=%s' --set 'service.nodePorts.https=%s' --set 'service.type=NodePort' --set config.oidc.enabled=true --set config.oidc.clientCredentialsSecret=client-credentials --set config.oidc.issuerURL=https://dex-01.wge.dev.weave.works --set config.oidc.redirectURL=https://weave.gitops.upgrade.enterprise.com:%s/oauth2/callback ",
						gitops_bin_path, version, prBranch, gitRepositoryURL, gitRepositoryURL, gitProviderEnv.Type, gitProviderEnv.Hostname, UI_NODEPORT, UI_NODEPORT)

					if gitProviderEnv.HostTypes != "" {
						upgradeCommand += ` --set "config.extraVolumes[0].name=ssh-config" --set "config.extraVolumes[0].configMap.name=ssh-config" --set "config.extraVolumeMounts[0].name=ssh-config" --set "config.extraVolumeMounts[0].mountPath=/root/.ssh"`
					}

					logger.Infof("Upgrade command: '%s'", upgradeCommand)
					stdOut, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && %s", repoAbsolutePath, upgradeCommand), ASSERTION_1MINUTE_TIME_OUT)
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
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "user-role-bindings.yaml")), "Failed to install rolbindings for enterprise default roles")
				})

				ginkgo.By("And I can also use upgraded enterprise UI/CLI after port forwarding (for loadbalancer ingress controller)", func() {
					serviceType, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get service clusters-service -n %s -o jsonpath="{.spec.type}"`, GITOPS_DEFAULT_NAMESPACE))
					if serviceType == "NodePort" {
						upgrade_capi_endpoint_url = fmt.Sprintf(`https://%s:%s`, GetEnv("UPGRADE_MANAGEMENT_CLUSTER_CNAME", "localhost"), UI_NODEPORT)
						upgrade_test_ui_url = fmt.Sprintf(`https://%s:%s`, GetEnv("UPGRADE_MANAGEMENT_CLUSTER_CNAME", "localhost"), UI_NODEPORT)
					} else {
						commandToRun := fmt.Sprintf("kubectl port-forward --namespace %s svc/clusters-service 8000:80", GITOPS_DEFAULT_NAMESPACE)

						cmd := exec.Command("sh", "-c", commandToRun)
						session, _ := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

						go func() {
							_ = session.Command.Wait()
						}()

						upgrade_test_ui_url = "http://localhost:8000"
						upgrade_capi_endpoint_url = "http://localhost:8000"
					}
					InitializeWebdriver(upgrade_test_ui_url)

					ginkgo.By(fmt.Sprintf("Login as a %s user", userCredentials.UserType), func() {
						loginUser() // Login to the weaveworks enterprise
					})
				})

				ginkgo.By("And the Cluster service is healthy", func() {
					CheckClusterService(upgrade_capi_endpoint_url)
				})

				// FIXME: CLI checks are disabled due to authentication not being supported
				// ginkgo.By("Then I should run enterprise CLI commands", func() {
				// 	testGetCommand := func(subCommand string) {
				// 		logger.Infof("Running 'gitops get %s --endpoint %s'", subCommand, upgrade_capi_endpoint_url)

				// 		cmd := fmt.Sprintf(`%s get %s --endpoint %s`, gitops_bin_path, subCommand, upgrade_capi_endpoint_url)
				// 		stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				// 		gomega.Expect(stdErr).Should(gomega.BeEmpty(), fmt.Sprintf("'%s get %s' command failed", gitops_bin_path, subCommand))
				// 		gomega.Expect(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`No %s[\s\w]+found`, subCommand)), fmt.Sprintf("'%s get %s' command failed", gitops_bin_path, subCommand))
				// 	}

				// 	testGetCommand("templates")
				// 	testGetCommand("credentials")
				// 	testGetCommand("clusters")
				// })

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				// Parameter values
				clusterName := "quick-capd-cluster"
				namespace := "quick-capi"
				k8Version := "1.23.3"
				controlPlaneMachineCount := "3"
				workerMachineCount := "3"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  namespace,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "",
						Option: controlPlaneMachineCount,
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)

				pages.ScrollWindow(webDriver, 0, 500)
				ginkgo.By("And verify selected weave-policy-agent profile values.yaml", func() {
					profile := createPage.GetProfileInList("weave-policy-agent")

					gomega.Eventually(profile.Version.Click).Should(gomega.Succeed())
					gomega.Eventually(pages.GetOption(webDriver, "0.3.1").Click).Should(gomega.Succeed(), "Failed to select weave-policy-agent version: 0.3.1")

					gomega.Eventually(profile.Layer.Text).Should(gomega.MatchRegexp("layer-1"))
					gomega.Expect(profile.Namespace.SendKeys("policy-system")).To(gomega.Succeed())

					gomega.Eventually(profile.Values.Click).Should(gomega.Succeed())
					valuesYaml := pages.GetValuesYaml(webDriver)

					gomega.Eventually(valuesYaml.Title.Text).Should(gomega.MatchRegexp("weave-policy-agent"))
					gomega.Eventually(valuesYaml.TextArea.Text).Should(gomega.MatchRegexp("namespace: policy-system"))

					text, _ := valuesYaml.TextArea.Text()
					text = strings.ReplaceAll(text, `accountId: ""`, `accountId: "weaveworks"`)
					text = strings.ReplaceAll(text, `clusterId: ""`, fmt.Sprintf(`clusterId: "%s"`, clusterName))
					gomega.Expect(valuesYaml.TextArea.Clear()).To(gomega.Succeed())
					gomega.Expect(valuesYaml.TextArea.SendKeys(text)).To(gomega.Succeed(), "Failed to change values.yaml for weave-policy-agent profile")

					gomega.Eventually(valuesYaml.Save.Click).Should(gomega.Succeed(), "Failed to save values.yaml for weave-policy-agent profile")
				})

				ginkgo.By("Then I should preview the PR", func() {
					preview := pages.GetPreview(webDriver)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")

					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*cni: calico`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*weave.works/flux: bootstrap`))
					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
				})

				pullRequest := PullRequest{
					Branch:  "feature-capd",
					Title:   "My first pull request",
					Message: "First capd capi template",
				}

				_ = createGitopsPR(pullRequest)

				var createPRUrl string
				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("And the manifests are present in the cluster config repository", func() {
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/clusters/capi/clusters/%s/%s.yaml", repoAbsolutePath, namespace, clusterName))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})
	})
}
