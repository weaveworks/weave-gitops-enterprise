package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

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

				By(fmt.Sprintf("And I run mccp templates render cluster-template-development-0 --list-parameters --endpoint %s", CAPI_ENDPOINT_URL), func() {
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

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --endpoint %s", clusterName, namespace, k8version, CAPI_ENDPOINT_URL), func() {
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

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s  --endpoint %s", clusterName, namespace, k8version, CAPI_ENDPOINT_URL), func() {
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

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --create-pr --pr-branch %s --pr-commit-message %s --endpoint %s", clusterName, namespace, k8version, prBranch, prCommit, CAPI_ENDPOINT_URL), func() {
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
			It("Verify mccp can create pull request to management cluster", func() {

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
				clusterName := "my-capd-cluster"
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

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --create-pr --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", clusterName, namespace, k8version, prBranch, prTitle, prCommit, CAPI_ENDPOINT_URL), func() {
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
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					pullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp can create multiple pull request to management cluster", func() {

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

				// CAPD Parameter values
				capdClusterName := "my-capd-cluster2"
				capdNamespace := "mccp-dev"
				capdK8version := "1.19.7"

				//CAPD Pull request values
				capdPRBranch := "feature-capd"
				capdPRTitle := "My first pull request"
				capdPRCommit := "First capd capi template"
				capdPRDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Insall CAPITemplate", func() {
					capdTemplateFile := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
					eksTemplateFile := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
					templateFiles = append(capdTemplateFile, eksTemplateFile...)
				})

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --create-pr --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", capdClusterName, capdNamespace, capdK8version, capdPRBranch, capdPRTitle, capdPRCommit, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", capdClusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", capdNamespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", capdK8version),
						"--create-pr", "--pr-branch", capdPRBranch, "--pr-title", capdPRTitle, "--pr-commit-message", capdPRCommit, "--pr-description", capdPRDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				var capdPRUrl string
				By("Then I should see pull request created for capd to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					capdPRUrl = string(match[1])
				})

				By("And I should veriyfy the capd pull request in the cluster config repository", func() {
					pullRequest := listPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(capdPRTitle))
					Expect(pullRequest[1]).Should(Equal(capdPRBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(capdPRUrl))
				})

				// EKS Parameter values
				eksClusterName := "my-eks-cluster"
				eksRegion := "eu-west-3"
				eksK8version := "1.19.8"
				eksSshKeyName := "my-aws-ssh-key"
				eksNamespace := "default" // FIXME: NAMESPACE parameter value is not required, need to get rid of it. it is just there to mask an existing bug WKP-2203

				//EKS Pull request values
				eksPRBranch := "feature-eks"
				eksPRTitle := "My second pull request"
				eksPRCommit := "First eks capi template"
				eksPRDescription := "This PR creates a new eks Kubernetes cluster"

				By(fmt.Sprintf("And I run 'mccp templates render eks-fargate-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --create-pr --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", eksClusterName, eksNamespace, eksRegion, eksK8version, eksPRBranch, eksPRTitle, eksPRCommit, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "eks-fargate-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", eksClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", eksRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", eksK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", eksSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", eksNamespace),
						"--create-pr", "--pr-branch", eksPRBranch, "--pr-title", eksPRTitle, "--pr-commit-message", eksPRCommit, "--pr-description", eksPRDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				var eksPRUrl string
				By("Then I should see pull request created for eks to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					eksPRUrl = string(match[1])
				})

				By("And I should veriyfy the eks pull request in the cluster config repository", func() {
					pullRequest := listPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(eksPRTitle))
					Expect(pullRequest[1]).Should(Equal(eksPRBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(eksPRUrl))
				})

				By("And the capd manifest is present in the cluster config repository", func() {
					pullBranch(repoAbsolutePath, capdPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, capdClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By("And the eks manifest is present in the cluster config repository", func() {
					pullBranch(repoAbsolutePath, eksPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, eksClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp can not create pull request to management cluster using existing branch", func() {

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

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					createGitRepoBranch(repoAbsolutePath, branchName)
				})

				// Parameter values
				clusterName := "my-dev-cluster"
				namespace := "mccp-dev"

				//Pull request values
				prTitle := "My dev pull request"
				prCommit := "First dev capi template"
				prDescription := "This PR creates a new dev Kubernetes cluster"

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-capitemplate.yaml")
				})

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s  --create-pr --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", clusterName, namespace, branchName, prTitle, prCommit, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace),
						"--create-pr", "--pr-branch", branchName, "--pr-title", prTitle, "--pr-commit-message", prCommit, "--pr-description", prDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should not see pull request to be created", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})
	})
}
