package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/fluxcd/go-git-providers/gitlab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func DescribeCliUpgrade(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops upgrade Tests", func() {

		UI_NODEPORT := "30081"
		NATS_NODEPORT := "31491"
		var capi_endpoint_url string
		var test_ui_url string
		var stdOut string
		var stdErr string

		BeforeEach(func() {

		})

		Context("[CLI] When gitops upgrade command is available", func() {
			It("Verify gitops upgrade command in --dry-run mode", func() {
				repositoryURL := fmt.Sprintf(`(https://|ssh://git@)%s/%s/%s.git`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)
				natsURL := "192.168.10.20:" + NATS_NODEPORT
				prBranch := "wego-enterprise-dry-run"
				version := "1.4.09"

				By("And I run gitops upgrade command", func() {
					upgradeCommand := fmt.Sprintf(" %s upgrade --version %s --branch %s --set 'agentTemplate.natsURL=%s' --set 'nats.client.service.nodePort=%s' --set 'nginx-ingress-controller.service.type=NodePort' --set 'nginx-ingress-controller.service.nodePorts.http=%s' --dry-run", gitops_bin_path, version, prBranch, natsURL, NATS_NODEPORT, UI_NODEPORT)
					logger.Infof("Upgrade command: '%s'", upgradeCommand)
					stdOut, stdErr = runCommandAndReturnStringOutput(upgradeCommand)
					Expect(stdErr).Should(BeEmpty())
				})

				By("And verify kind 'HelmRepository' in upgrade manifest", func() {
					Expect(stdOut).Should(MatchRegexp(`kind: HelmRepository[\s\w\d./:-]*name: weave-gitops-enterprise-charts[\s\w\d./:-]*secretRef:[\s]*name: weave-gitops-enterprise-credentials[\s]*url: https://charts.dev.wkp.weave.works/releases/charts-v3`))
				})

				By("And verify kind 'HelmRelease' in upgrade manifest", func() {
					Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*sourceRef:[\s]*kind: HelmRepository[\s]*name: weave-gitops-enterprise-charts[\s\w\d./:-]*version: %s`, version)))
					Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d.@/:-]*nginx-ingress-controller[\s\w\d.@/:-]*nodePorts:[\s]*http: %s`, UI_NODEPORT)))
					Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*natsURL: %s`, natsURL)))
					Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*repositoryURL: %s`, repositoryURL)))
				})

				By("And verify kind 'Kustomization' in upgrade manifest", func() {
					Expect(stdOut).Should(MatchRegexp(`kind: Kustomization[\s\w\d./:-]*sourceRef:[\s]*kind: GitRepository[\s]*name: wego-`))
				})

			})
		})

		Context("[CLI] When Wego core is installed in the cluster", func() {
			var current_context string
			var public_ip string
			kind_upgrade_cluster_name := "test-upgrade"

			templateFiles := []string{}

			JustBeforeEach(func() {
				current_context, _ = runCommandAndReturnStringOutput("kubectl config current-context")

				// Create vanilla cluster for WGE upgrade
				createCluster("kind", kind_upgrade_cluster_name, "upgrade-kind-config.yaml")

			})

			JustAfterEach(func() {

				gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
				templateFiles = []string{}

				err := runCommandPassThrough("kubectl", "config", "use-context", current_context)
				Expect(err).ShouldNot(HaveOccurred())

				deleteClusters("kind", []string{kind_upgrade_cluster_name})

			})

			It("@upgrade @git Verify wego core can be upgraded to wego enterprise", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				By("When I create a private repository for cluster configs", func() {
					initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				By("When I install gitops/wego to my active cluster", func() {
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
				})

				By("And I install Profile (HelmRepository chart)", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "profile-repo.yaml")), "Failed to install profiles HelmRepository chart")
				})

				By("And I install the entitlement for cluster upgrade", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to create/configure entitlement")
				})

				By("And I install the git repository secret for cluster service", func() {
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
							Expect(stdOut).Should(MatchRegexp(`configmap/ssh-config created`), "Failed to create on-prem known hosts 'ssh-config' configmap")

							cmd = fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"  --from-literal="GITLAB_CLIENT_ID=%s" --from-literal="GITLAB_CLIENT_SECRET=%s"  --from-literal="GITLAB_HOSTNAME=%s" --from-literal="GIT_HOST_TYPES=%s"`,
								GITOPS_DEFAULT_NAMESPACE, gitProviderEnv.Token, gitProviderEnv.ClientId, gitProviderEnv.ClientSecret, gitProviderEnv.GitlabHostname, gitProviderEnv.HostTypes)
						}
					}
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
					Expect(stdErr).Should(BeEmpty(), "Failed to create git repository secret for cluster service")
				})

				By("And I should update/modify the default upgrade manifest ", func() {
					public_ip = clusterWorkloadNonePublicIP("KIND")
				})

				prBranch := "wego-upgrade-enterprise"
				version := "0.0.19"
				By(fmt.Sprintf("And I run gitops upgrade command from directory %s", repoAbsolutePath), func() {
					natsURL := public_ip + ":" + NATS_NODEPORT
					// Explicitly setting the gitprovider type, hostname and repository path url scheme in configmap, the default is github and ssh url scheme which is nopt supported for capi cluster PR creation.
					upgradeCommand := fmt.Sprintf(" %s upgrade --version %s --branch %s --set 'config.capi.repositoryURL=%s' --set 'config.git.type=%s' --set 'config.git.hostname=%s' --set 'agentTemplate.natsURL=%s' --set 'nats.client.service.nodePort=%s' --set 'nginx-ingress-controller.service.type=NodePort' --set 'nginx-ingress-controller.service.nodePorts.http=%s' ",
						gitops_bin_path, version, prBranch, getGitRepositoryURL(repoAbsolutePath), gitProviderEnv.Type, gitProviderEnv.Hostname, natsURL, NATS_NODEPORT, UI_NODEPORT)

					if gitProviderEnv.HostTypes != "" {
						upgradeCommand += fmt.Sprintf(` --git-host-types="%s" --set "config.extraVolumes[0].name=ssh-config" --set "config.extraVolumes[0].configMap.name=ssh-config" --set "config.extraVolumeMounts[0].name=ssh-config" --set "config.extraVolumeMounts[0].mountPath=/root/.ssh"`,
							gitProviderEnv.HostTypes)
					}

					logger.Infof("Upgrade command: '%s'", upgradeCommand)
					stdOut, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && %s", repoAbsolutePath, upgradeCommand), ASSERTION_1MINUTE_TIME_OUT)
					Expect(stdErr).Should(BeEmpty())
				})

				By("Then I should see pull request created to management cluster", func() {
					re := regexp.MustCompile(`Pull Request created.*:[\s\w\d]+(?P<URL>https.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
					Eventually(match[1]).ShouldNot(BeNil(), "Failed to Create pull request")
				})

				By("Then I should merge the pull request to start weave gitops enterprise upgrade", func() {
					upgradePRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, upgradePRUrl)
				})

				By("And I should see cluster upgraded from 'wego core' to 'wego enterprise'", func() {
					verifyEnterpriseControllers("weave-gitops-enterprise", "mccp-", GITOPS_DEFAULT_NAMESPACE)
				})

				By("And I can also use upgraded enterprise UI/CLI after port forwarding (for loadbalancer ingress controller)", func() {
					// Temporary FIX: Since v0.0.19 uses nginx-ingress-controller, we need to check for ingress service type
					serviceType, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get service weave-gitops-enterprise-nginx-ingress-controller -n %s -o jsonpath="{.spec.type}"`, GITOPS_DEFAULT_NAMESPACE))
					// serviceType, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get service clusters-service -n %s -o jsonpath="{.spec.type}"`, GITOPS_DEFAULT_NAMESPACE))
					if serviceType == "NodePort" {
						capi_endpoint_url = fmt.Sprintf(`http://%s:%s`, GetEnv("UPGRADE_MANAGEMENT_CLUSTER_CNAME", "localhost"), UI_NODEPORT)
						test_ui_url = fmt.Sprintf(`http://%s:%s`, GetEnv("UPGRADE_MANAGEMENT_CLUSTER_CNAME", "localhost"), UI_NODEPORT)
					} else {
						commandToRun := fmt.Sprintf("kubectl port-forward --namespace %s svc/clusters-service 8000:80", GITOPS_DEFAULT_NAMESPACE)

						cmd := exec.Command("sh", "-c", commandToRun)
						session, _ := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

						go func() {
							_ = session.Command.Wait()
						}()

						test_ui_url = "http://localhost:8000"
						capi_endpoint_url = "http://localhost:8000"
					}
					InitializeWebdriver(test_ui_url)
				})

				By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				By("Then I should run enterprise CLI commands", func() {
					testGetCommand := func(subCommand string) {
						logger.Infof("Running 'gitops get %s --endpoint %s'", subCommand, capi_endpoint_url)

						cmd := fmt.Sprintf(`%s get %s --endpoint %s`, gitops_bin_path, subCommand, capi_endpoint_url)
						stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
						Expect(stdErr).Should(BeEmpty(), fmt.Sprintf("'%s get %s' command failed", gitops_bin_path, subCommand))
						Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`No %s[\s\w]+found`, subCommand)), fmt.Sprintf("'%s get %s' command failed", gitops_bin_path, subCommand))
					}

					testGetCommand("templates")
					testGetCommand("credentials")
					testGetCommand("clusters")
				})

				By("And I can connect cluster to itself", func() {
					// leaf := LeafSpec{
					// 	Status:          "Ready",
					// 	IsWKP:           false,
					// 	AlertManagerURL: "",
					// 	KubeconfigPath:  "",
					// }
					// connectACluster(webDriver, gitopsTestRunner, leaf)
				})

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				// Parameter values
				clusterName := "quick-capd-cluster"
				namespace := "quick-capi"
				k8Version := "1.22.0"

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
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
				}
				paramSection["4.KubeadmControlPlane"] = []TemplateField{
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
				}

				setParameterValues(createPage, paramSection)
				pages.ScrollWindow(webDriver, 0, 4000)
				// Temporaroary FIX - Authenticating before profile selection. Gitlab authentication redirect resets the profiles section
				AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
				pages.ScrollWindow(webDriver, 0, 4000)

				By("And select the podinfo profile to install", func() {
					Eventually(createPage.ProfileSelect.Click).Should(Succeed())
					Eventually(createPage.SelectProfile("podinfo").Click).Should(Succeed())
					pages.DissmissProfilePopup(webDriver)
				})

				By("And verify selected podinfo profile values.yaml", func() {
					profile := pages.GetProfile(webDriver, "podinfo")

					Eventually(profile.Version.Click).Should(Succeed())
					Eventually(pages.GetOption(webDriver, "6.0.0").Click).Should(Succeed())

					Eventually(profile.Layer.Text).Should(MatchRegexp("layer-1"))

					Eventually(profile.Values.Click).Should(Succeed())
					valuesYaml := pages.GetValuesYaml(webDriver)

					Eventually(valuesYaml.Title.Text).Should(MatchRegexp("podinfo"))
					Eventually(valuesYaml.TextArea.Text).Should(MatchRegexp("tag: 6.0.0"))
					Eventually(valuesYaml.Cancel.Click).Should(Succeed())
				})

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)

					Eventually(preview.Title).Should(MatchText("PR Preview"))
					Eventually(preview.Close.Click).Should(Succeed())
				})

				//Pull request values
				prBranch = "feature-capd"
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(prBranch)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					Eventually(gitops.GitCredentials).Should(BeVisible())
					if pages.ElementExist(gitops.SuccessBar) {
						Expect(gitops.SuccessBar.Click()).To(Succeed())
					}
					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				var prUrl string
				clustersPage := pages.GetClustersPage(webDriver)
				By("Then I should see cluster appears in the cluster dashboard with the expected status", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(HaveText("Creation PR"))
					anchor := clusterInfo.Status.Find("a")
					Eventually(anchor).Should(BeFound())
					prUrl, _ = anchor.Attribute("href")
				})

				var createPRUrl string
				By("And I should veriyfy the pull request in the cluster config repository", func() {
					createPRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					Expect(createPRUrl).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/.weave-gitops/apps/capi/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})
	})
}
