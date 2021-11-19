package acceptance

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func configureConfigMapvalues(repoAbsolutePath string, uiNodePort string, natsNodeport string, natsURL string, repositoryURL string) {
	pullGitRepo(repoAbsolutePath)

	configMapYaml := path.Join(repoAbsolutePath, "upgrade", "weave-gitops-enterprise", "artifacts", "mccp-chart", "helm-chart", "ConfigMap.yaml")

	input, err := ioutil.ReadFile(configMapYaml)
	Expect(err).ShouldNot(HaveOccurred())

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if uiNodePort != "" && strings.Contains(line, "nginx-ingress-controller") {
			lines = append(lines[:i+2], lines[i+1:]...)
			lines[i+2] = "        type: NodePort"
			lines[i+5] = "            " + uiNodePort
		}
		if natsNodeport != "" && strings.Contains(line, "nodePort: 31490") {
			lines[i] = "          nodePort: " + natsNodeport
		}
		if natsURL != "" && strings.Contains(line, "natsURL: ") {
			lines[i] = "      natsURL: " + natsURL
		}
		if repositoryURL != "" && strings.Contains(line, "repositoryURL: ") {
			lines[i] = "        repositoryURL: " + repositoryURL
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(configMapYaml, []byte(output), 0644)
	Expect(err).ShouldNot(HaveOccurred())

}

func DescribeCliUpgrade(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops upgrade Tests", func() {

		GITOPS_BIN_PATH := GetGitopsBinPath()
		UI_NODEPORT := "30081"
		NATS_NODEPORT := "31491"
		var capi_endpoint_url string
		var test_ui_url string
		var repoAbsolutePath string

		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(FileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
			})
		})

		AfterEach(func() {

		})

		Context("[CLI] When Wego core is installed in the cluster", func() {
			var current_context string
			var public_ip string
			kind_upgrade_cluster_name := "test-upgrade"

			appName := "wego-upgrade"
			appPath := "upgrade"
			templateFiles := []string{}

			JustBeforeEach(func() {
				current_context, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				current_context = strings.Trim(current_context, "\n")

				// Create vanilla cluster for WGE upgrade
				CreateCluster("kind", kind_upgrade_cluster_name, "upgrade-kind-config.yaml")

				By("And cluster repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

			})

			JustAfterEach(func() {

				gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
				templateFiles = []string{}

				gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				err := runCommandPassThrough([]string{}, "kubectl", "config", "use-context", current_context)
				Expect(err).ShouldNot(HaveOccurred())

				deleteClusters("kind", []string{kind_upgrade_cluster_name})

			})

			It("@upgrade Verify wego core can be upgraded to wego enterprise", func() {

				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("When I install gitops/wego to my active cluster", func() {
					InstallAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, GetGitRepositoryURL(repoAbsolutePath))
				})

				By("And I install profile controllers to my active cluster", func() {
					InstallAndVerifyPctl(GITOPS_DEFAULT_NAMESPACE)
				})

				addCommand := fmt.Sprintf("add app . --path=./%s  --name=%s  --auto-merge=true", appPath, appName)
				By(fmt.Sprintf("And I run gitops add app command ' %s 'in namespace %s from dir %s", addCommand, GITOPS_DEFAULT_NAMESPACE, repoAbsolutePath), func() {
					cmd := fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, addCommand)
					_, err := runCommandAndReturnStringOutput(cmd)
					Expect(err).Should(BeEmpty())
				})

				By("And I install the entitlement for cluster upgrade", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				By("And I install the git repository secret for cluster service", func() {
					cmd := fmt.Sprintf(`kubectl create secret generic git-provider-credentials --namespace=%s --from-literal="GIT_PROVIDER_TOKEN=%s"`, GITOPS_DEFAULT_NAMESPACE, GITHUB_TOKEN)
					_, err := runCommandAndReturnStringOutput(cmd)
					Expect(err).Should(BeEmpty(), "Failed to create git repository secret for cluster service")
				})

				By("And I install the docker registry secret for wego enteprise components", func() {
					cmd := fmt.Sprintf(`kubectl create secret docker-registry docker-io-pull-secret --namespace=%s --docker-username=%s --docker-password=%s`, GITOPS_DEFAULT_NAMESPACE, DOCKER_IO_USER, DOCKER_IO_PASSWORD)
					_, err := runCommandAndReturnStringOutput(cmd)
					Expect(err).Should(BeEmpty(), "Failed to create git repository secret for cluster service")
				})

				By("Then I should see gitops add command linked the repo to the cluster", func() {
					verifyWegoAddCommand(appName, GITOPS_DEFAULT_NAMESPACE)
				})

				prBranch := "wego-upgrade-enterprise"
				upgradeCommand := fmt.Sprintf("upgrade --git-repository %s/%s --branch %s --out %s", GITOPS_DEFAULT_NAMESPACE, appName, prBranch, appPath)
				By(fmt.Sprintf("And I run gitops upgrade command ' %s ' form firectory %s", upgradeCommand, repoAbsolutePath), func() {
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, upgradeCommand))
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
					Expect(string(session.Err.Contents())).Should(BeEmpty())
				})

				By("Then I should see pull request created to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`PR created.*:[\s\w\d]+(?P<URL>https.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match[1]).ShouldNot(BeNil(), "Failed to Create pull request")
				})

				By("And I should update/modify the default upgrade manifest ", func() {
					public_ip = ClusterWorkloadNonePublicIP("KIND")
					natsURL := public_ip + ":" + NATS_NODEPORT
					repositoryURL := fmt.Sprintf(`https://github.com/%s/%s`, GITHUB_ORG, CLUSTER_REPOSITORY)
					gitopsTestRunner.PullBranch(repoAbsolutePath, prBranch)
					configureConfigMapvalues(repoAbsolutePath, UI_NODEPORT, NATS_NODEPORT, natsURL, repositoryURL)
					GitSetUpstream(repoAbsolutePath, prBranch)
					GitUpdateCommitPush(repoAbsolutePath, "")
				})

				By("Then I should merge the pull request to start weave gitops enterprise upgrade", func() {
					gitopsTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
				})

				By("And I should see cluster upgraded from 'wego core' to 'wego enterprise'", func() {
					VerifyEnterpriseControllers("mccp-chart", GITOPS_DEFAULT_NAMESPACE)
				})

				By("And I can change the config map values for the upgrade profile", func() {
					repositoryURL := fmt.Sprintf(`https://github.com/%s/%s`, GITHUB_ORG, CLUSTER_REPOSITORY)
					configureConfigMapvalues(repoAbsolutePath, "", "", "", repositoryURL)
					GitUpdateCommitPush(repoAbsolutePath, "")
				})

				By("Then I restart the cluster service pod for capi config to take effect", func() {
					configmapValuesUpdated := func() bool {
						data, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get configmap -n %s weave-gitops-enterprise-mccp-chart-defaultvalues -o jsonpath="{.data}"`, GITOPS_DEFAULT_NAMESPACE))
						re := regexp.MustCompile(fmt.Sprintf(`repositoryURL:\s*(?P<url>https:.*%s)`, CLUSTER_REPOSITORY))
						return len(re.FindSubmatch([]byte(data))) > 0
					}

					Eventually(configmapValuesUpdated, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeTrue(), "ConfigMap values failed to reconcile")
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, "mccp-chart-cluster-service", GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I can also use upgraded enterprise UI/CLI after port forwarding (for loadbalancer ingress controller)", func() {
					serviceType, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get service mccp-chart-nginx-ingress-controller -n %s -o jsonpath="{.spec.type}"`, GITOPS_DEFAULT_NAMESPACE))
					if strings.Trim(serviceType, "\n") == "NodePort" {
						capi_endpoint_url = "http://" + public_ip + ":" + UI_NODEPORT
						test_ui_url = "http://" + public_ip + ":" + UI_NODEPORT
					} else {
						commandToRun := fmt.Sprintf("kubectl port-forward --namespace %s deployments.apps/mccp-chart-nginx-ingress-controller 8000:80", GITOPS_DEFAULT_NAMESPACE)

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
