package acceptance

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func DescribeCliUpgrade(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops upgrade Tests", func() {

		UI_NODEPORT := "30081"
		NATS_NODEPORT := "31491"
		var capi_endpoint_url string
		var test_ui_url string
		var repoAbsolutePath string

		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(fileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
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

				deleteRepo(gitProviderEnv)

				err := runCommandPassThrough([]string{}, "kubectl", "config", "use-context", current_context)
				Expect(err).ShouldNot(HaveOccurred())

				deleteClusters("kind", []string{kind_upgrade_cluster_name})

			})

			It("@upgrade @git Verify wego core can be upgraded to wego enterprise", func() {

				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				By("When I install gitops/wego to my active cluster", func() {
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
				})

				By("And I install the entitlement for cluster upgrade", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				By("And I install the git repository secret for cluster service", func() {
					cmd := fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"`, GITOPS_DEFAULT_NAMESPACE, gitProviderEnv.Token)
					_, err := runCommandAndReturnStringOutput(cmd)
					Expect(err).Should(BeEmpty(), "Failed to create git repository secret for cluster service")
				})

				By("And I should update/modify the default upgrade manifest ", func() {
					public_ip = clusterWorkloadNonePublicIP("KIND")
				})

				prBranch := "wego-upgrade-enterprise"
				version := "0.0.17"
				By(fmt.Sprintf("And I run gitops upgrade command from directory %s", repoAbsolutePath), func() {
					natsURL := public_ip + ":" + NATS_NODEPORT
					upgradeCommand := fmt.Sprintf("upgrade --version %s --branch %s --set 'agentTemplate.natsURL=%s' --set 'nats.client.service.nodePort=%s'", version, prBranch, natsURL, NATS_NODEPORT)
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, upgradeCommand))
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
					Expect(string(session.Err.Contents())).Should(BeEmpty())
				})

				By("Then I should see pull request created to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`Pull Request created.*:[\s\w\d]+(?P<URL>https.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
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
					serviceType, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get service weave-gitops-enterprise-nginx-ingress-controller -n %s -o jsonpath="{.spec.type}"`, GITOPS_DEFAULT_NAMESPACE))
					if serviceType == "NodePort" {
						capi_endpoint_url = "http://" + public_ip + ":" + UI_NODEPORT
						test_ui_url = "http://" + public_ip + ":" + UI_NODEPORT
					} else {
						commandToRun := fmt.Sprintf("kubectl port-forward --namespace %s deployments.apps/weave-gitops-enterprise-nginx-ingress-controller 8000:80", GITOPS_DEFAULT_NAMESPACE)

						cmd := exec.Command("sh", "-c", commandToRun)
						session, _ := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

						go func() {
							_ = session.Command.Wait()
						}()

						test_ui_url = "http://localhost:8000"
						capi_endpoint_url = "http://localhost:8000"
					}
					initializeWebdriver(test_ui_url)
				})

				By("And the Cluster service is healthy", func() {
					gitopsTestRunner.CheckClusterService(capi_endpoint_url)
				})

				By("Then I should run enterprise CLI commands", func() {
					testGetCommand := func(subCommand string) {
						log.Printf("Running 'gitops get %s --endpoint %s'", subCommand, capi_endpoint_url)

						command := exec.Command(GITOPS_BIN_PATH, "get", subCommand, "--endpoint", capi_endpoint_url)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
						Eventually(session).Should(gexec.Exit())
						Expect(string(session.Err.Contents())).Should(BeEmpty(), fmt.Sprintf("'%s get %s' command failed", GITOPS_BIN_PATH, subCommand))
						Expect(string(session.Out.Contents())).Should(MatchRegexp(fmt.Sprintf(`No %s[\s\w]+found`, subCommand)), fmt.Sprintf("'%s get %s' command failed", GITOPS_BIN_PATH, subCommand))
					}

					testGetCommand("templates")
					testGetCommand("credentials")
					testGetCommand("clusters")
				})

				By("And I can connect cluster to itself", func() {
					leaf := LeafSpec{
						Status:          "Ready",
						IsWKP:           false,
						AlertManagerURL: "",
						KubeconfigPath:  "",
					}
					connectACluster(webDriver, gitopsTestRunner, leaf)
				})
			})
		})
	})
}
