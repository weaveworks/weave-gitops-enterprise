package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func DescribeMccpCliRender(mccpTestRunner MCCPTestRunner) {
	var _ = Describe("MCCP Template Render Tests", func() {

		MCCP_BIN_PATH := GetMCCBinPath()
		CAPI_ENDPOINT_URL := GetCapiEndpointUrl()

		templateFiles := []string{}
		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a mccp binary installed on my local machine", func() {
				Expect(FileExists(MCCP_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", MCCP_BIN_PATH))
			})
		})

		AfterEach(func() {
			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp can render template parameters of a template from template library", func() {

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By("And I run 'mccp templates render <template-name> --list-parameters --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template parameter table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+DESCRIPTION`))
				})

				By("And I should see parameter rows", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`CLUSTER_NAME+\s+This is used for the cluster naming.\s+KUBERNETES_VERSION\s+Kubernetes version to use for the cluster\s+NAMESPACE\s+Namespace to create the cluster in`)
					Eventually((re.Find(output))).ShouldNot(BeNil())

				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp can set template parameters by specifying multiple parameters --set key=value --set key=value", func() {
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By("And I run 'mccp templates render <template-name> --set <parameter=value> --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version), "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template preview with updated parameter values", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of tbe template for updated  parameter values
					re := regexp.MustCompile(fmt.Sprintf(`kind: Cluster\s+metadata:\s+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: KubeadmControlPlane\s+metadata:\s+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: MachineDeployment\s+metadata:\s+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			XIt("Verify mccp can set template parameters by separate values with commas key1=val1,key2=val2", func() {
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By("And I run 'mccp templates render <template-name> --set <parameter1=value1,parameter2=value2> --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0",
						"--set", fmt.Sprintf("CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s", clusterName, namespace, k8version), "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template preview with updated parameter values", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of tbe template for updated  parameter values
					re := regexp.MustCompile(fmt.Sprintf(`kind: Cluster\s+metadata:\s+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: KubeadmControlPlane\s+metadata:\s+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: MachineDeployment\s+metadata:\s+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp reports an error when trying to create pull request with missing --create-pr arguments", func() {
				// Parameter values
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By("And I run 'mccp templates render <template-name> --set <parameter1=value1,parameter2=value2> --create-pr --pr-branch <branch name> --pr-title <PR title> --pr-commit-message <commit message> --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
						"--create-pr", "--pr-branch", prBranch, "--pr-commit-message", prCommit, "--pr-description", prDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see an error for required argument to create pull request", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(`Error: unable to create pull request.*title must be specified`))
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			FIt("Verify mccp can create pull request to management cluster", func() {

				defer deleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					deleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = initAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					gitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And repo created has private visibility", func() {
					Expect(getRepoVisibility(GITHUB_ORG, CLUSTER_REPOSITORY)).Should(ContainSubstring("true"))
				})

				// Parameter values
				clusterName := "my-dev-cluster1"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prTitle := "My first pull request"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By("And I run 'mccp templates render <template-name> --set <parameter1=value1,parameter2=value2> --create-pr --pr-branch <branch name> --pr-title <PR title> --pr-commit-message <commit message> --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
						"--create-pr", "--pr-branch", prBranch, "--pr-title", prTitle, "--pr-commit-message", prCommit, "--pr-description", prDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				var prUrl string
				By("Then I should see pull request created to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					prUrl = string(match[1])
				})

				By("And I should veriyfy the pull request in the cluster config repository", func() {
					pullRequest := listPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(pullRequest[2]).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					pullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})
	})
}
