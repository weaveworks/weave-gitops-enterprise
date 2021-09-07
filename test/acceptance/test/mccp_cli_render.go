package acceptance

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func DescribeMccpCliRender(mccpTestRunner MCCPTestRunner) {
	var _ = Describe("MCCP Template Render Tests", func() {

		MCCP_BIN_PATH := GetMccpBinPath()
		WEGO_BIN_PATH := GetWegoBinPath()
		CAPI_ENDPOINT_URL := GetCapiEndpointUrl()

		templateFiles := []string{}
		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a mccp binary installed on my local machine", func() {
				Expect(FileExists(MCCP_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", MCCP_BIN_PATH))
			})

			By("And the Cluster service is healthy", func() {
				mccpTestRunner.checkClusterService()
			})
		})

		AfterEach(func() {
			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
			// Reset/empty the templateFiles list
			templateFiles = []string{}
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can list template parameters of a template from template library", func() {

				By("Apply/Install CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run mccp templates render cluster-template-development-0 --list-parameters --endpoint %s", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template parameter table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see parameter rows", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`CLUSTER_NAME+\s+This is used for the cluster naming.\s+KUBERNETES_VERSION\s+Kubernetes version to use for the cluster\s+1.19.7, 1.19.8\s+NAMESPACE\s+Namespace to create the cluster in`)
					Eventually((re.Find(output))).ShouldNot(BeNil())

				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can set template parameters by specifying multiple parameters --set key=value --set key=value", func() {
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
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

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can set template parameters by separate values with commas key1=val1,key2=val2", func() {
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
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

					// Verifying MachineDeployment object of the template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: MachineDeployment\s+metadata:\s+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})
		})

		Context("[CLI] When invalid Capi Template(s) are available in the cluster", func() {
			It("Verify mccp reports an error when rendering template parameters of invalid template from template library", func() {

				noOfTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-invalid-capitemplate.yaml")
				})

				By(fmt.Sprintf(`And I run 'mccp templates render cluster-invalid-template-0 --list-parameters --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-invalid-template-0", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should not see template parameter table header", func() {
					Eventually(string(session.Wait().Out.Contents())).ShouldNot(MatchRegexp(`NAME\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see error message related to invalid template", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`Error: unable to retrieve template parameters`)
					Eventually((re.Find(output))).Should(BeNil())
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp reports an error when trying to create pull request with missing --create-pr arguments", func() {
				// Parameter values
				clusterName := "development-cluster"
				namespace := "mccp-dev"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
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

		Context("[CLI] When no clusters are available in the management cluster", func() {
			It("Verify mccp lists no clusters", func() {

				By("And MCCP state is reset", func() {
					mccpTestRunner.ResetDatabase()
					mccpTestRunner.VerifyMCCPPodsRunning()
					mccpTestRunner.checkClusterService()
				})

				By(fmt.Sprintf("Then I run 'mccp clusters list --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then mccp lists no clusters", func() {
					Eventually(session).Should(gbytes.Say("No clusters found"))
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can create pull request to management cluster", func() {

				defer mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And repo created has private visibility", func() {
					Expect(mccpTestRunner.GetRepoVisibility(GITHUB_ORG, CLUSTER_REPOSITORY)).Should(ContainSubstring("true"))
				})

				// Parameter values
				clusterName := "my-capd-cluster"
				namespace := "default"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prTitle := "My first pull request"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
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
					pullRequest := mccpTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					mccpTestRunner.PullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By(fmt.Sprintf("Then I run 'mccp clusters list --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see cluster status as 'pullRequestCreated'", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

					re := regexp.MustCompile(fmt.Sprintf(`%s\s+pullRequestCreated`, clusterName))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can create multiple pull request to management cluster", func() {

				defer mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
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

				By("Apply/Install CAPITemplate", func() {
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
					pullRequest := mccpTestRunner.ListPullRequest(repoAbsolutePath)
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
					pullRequest := mccpTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(eksPRTitle))
					Expect(pullRequest[1]).Should(Equal(eksPRBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(eksPRUrl))
				})

				By("And the capd manifest is present in the cluster config repository", func() {
					mccpTestRunner.PullBranch(repoAbsolutePath, capdPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, capdClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By("And the eks manifest is present in the cluster config repository", func() {
					mccpTestRunner.PullBranch(repoAbsolutePath, eksPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, eksClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By(fmt.Sprintf("Then I run 'mccp clusters list --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see cluster status as 'pullRequestCreated'", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

					re := regexp.MustCompile(fmt.Sprintf(`%s\s+pullRequestCreated`, eksClusterName))
					Eventually((re.Find(output))).ShouldNot(BeNil())
					re = regexp.MustCompile(fmt.Sprintf(`%s\s+pullRequestCreated`, capdClusterName))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify mccp can not create pull request to management cluster using existing branch", func() {

				defer mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					mccpTestRunner.CreateGitRepoBranch(repoAbsolutePath, branchName)
				})

				// Parameter values
				clusterName := "my-dev-cluster"
				namespace := "mccp-dev"

				//Pull request values
				prTitle := "My dev pull request"
				prCommit := "First dev capi template"
				prDescription := "This PR creates a new dev Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
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

		Context("[CLI] When no infrastructure provider credentials are available in the management cluster", func() {
			It("Verify mccp lists no credentials", func() {
				By("Apply/Install CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --list-credentials --endpoint %s", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--list-credentials", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then mccp lists no credentials", func() {
					Eventually(session).Should(gbytes.Say("No credentials found"))
				})
			})
		})

		Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {
			It("Verify mccp can use the matching selected credential for cluster creation", func() {
				defer mccpTestRunner.DeleteIPCredentials("AWS")
				defer mccpTestRunner.DeleteIPCredentials("AZURE")

				By("Apply/Install CAPITemplates", func() {
					eksTemplateFile := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
					azureTemplateFiles := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
					templateFiles = append(azureTemplateFiles, eksTemplateFile...)
				})

				By("And create AWS credentials)", func() {
					mccpTestRunner.CreateIPCredentials("AWS")
				})

				By(fmt.Sprintf("And I run 'mccp templates render aws-cluster-template-0 --list-credentials --endpoint %s", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "aws-cluster-template-0", "--list-credentials", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then mccp lists AWS credentials", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`aws-test-identity`))
					Eventually(string(output)).Should(MatchRegexp(`test-role-identity`))
				})

				By("And create AZURE credentials)", func() {
					mccpTestRunner.CreateIPCredentials("AZURE")
				})

				By(fmt.Sprintf("And I run 'mccp templates render azure-capi-quickstart-template-0 --list-credentials --endpoint %s", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "azure-capi-quickstart-template-0", "--list-credentials", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then mccp lists AZURE credentials", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`azure-cluster-identity`))
				})

				// AWS Parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				By(fmt.Sprintf("And I run 'mccp templates render aws-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials aws-test-identity --endpoint %s", awsClusterName, awsNamespace, awsRegion, awsK8version, awsControlMAchineType, awsNodeMAchineType, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "aws-cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", awsClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", awsRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", awsK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", awsSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", awsNamespace),
						"--set", "CONTROL_PLANE_MACHINE_COUNT=2", "--set", fmt.Sprintf("AWS_CONTROL_PLANE_MACHINE_TYPE=%s", awsControlMAchineType),
						"--set", "WORKER_MACHINE_COUNT=3", "--set", fmt.Sprintf("AWS_NODE_MACHINE_TYPE=%s", awsNodeMAchineType),
						"--set-credentials", "aws-test-identity", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see preview containing identity reference added in the template", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of the template for added credential reference
					re := regexp.MustCompile(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterStaticIdentity\s+name: aws-test-identity`, awsClusterName))

					Eventually((re.Find(output))).ShouldNot(BeNil(), "Failed to find identity reference in preview pull request AWSCluster object")
				})
			})
		})

		Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {
			It("Verify mccp restrict user from using wrong credentials for infrastructure provider", func() {
				defer mccpTestRunner.DeleteIPCredentials("AZURE")

				By("Apply/Install CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
				})

				By("And create AZURE credentials)", func() {
					mccpTestRunner.CreateIPCredentials("AZURE")
				})

				// AWS Parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				By(fmt.Sprintf("And I run 'mccp templates render aws-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials azure-cluster-identity --endpoint %s", awsClusterName, awsNamespace, awsRegion, awsK8version, awsControlMAchineType, awsNodeMAchineType, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "aws-cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", awsClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", awsRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", awsK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", awsSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", awsNamespace),
						"--set", "CONTROL_PLANE_MACHINE_COUNT=2", "--set", fmt.Sprintf("AWS_CONTROL_PLANE_MACHINE_TYPE=%s", awsControlMAchineType),
						"--set", "WORKER_MACHINE_COUNT=3", "--set", fmt.Sprintf("AWS_NODE_MACHINE_TYPE=%s", awsNodeMAchineType),
						"--set-credentials", "azure-cluster-identity", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				// FIXME - User should get some warning or error as well for chossing wrong credential/identity for the infrastructure provider

				By("Then I should see preview without identity reference added to the template", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`kind: AWSCluster[\s\w\d-.:/]+identityRef:`)
					Eventually((re.Find(output))).Should(BeNil(), "Identity reference should not be found in preview pull request AWSCluster object")
				})
			})
		})

		Context("[CLI] When leaf cluster pull request is available in the management cluster", func() {
			JustBeforeEach(func() {
				log.Println("Connecting cluster to itself")
				initializeWebdriver()
				leaf := LeafSpec{
					Status:          "Ready",
					IsWKP:           false,
					AlertManagerURL: "",
					KubeconfigPath:  "",
				}
				connectACluster(webDriver, mccpTestRunner, leaf)
			})

			JustAfterEach(func() {
				log.Println("Deleting all the wkp agents")
				mccpTestRunner.KubectlDeleteAllAgents([]string{})
				mccpTestRunner.ResetDatabase()
				mccpTestRunner.VerifyMCCPPodsRunning()
			})

			It("@VM Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {
				defer mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				defer resetWegoRuntime(WEGO_DEFAULT_NAMESPACE)

				By("And template repo does not already exist", func() {
					mccpTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And I reset wego runtime", func() {
					resetWegoRuntime(WEGO_DEFAULT_NAMESPACE)
				})

				By("And I install wego to my active cluster", func() {
					Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", WEGO_BIN_PATH))
					installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
				})

				addCommand := "app add . --path=./management  --name=management  --auto-merge=true"
				By(fmt.Sprintf("And I run wego app add command '%s in namespace %s from dir %s'", addCommand, WEGO_DEFAULT_NAMESPACE, repoAbsolutePath), func() {
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, WEGO_BIN_PATH, addCommand))
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
				})

				By("And I install Docker provider infrastructure", func() {
					installInfrastructureProvider("docker")
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				createCluster := func(clusterName string, namespace string, k8version string) {
					//Pull request values
					prBranch := fmt.Sprintf("br-%s", clusterName)
					prTitle := "CAPD pull request"
					prCommit := "CAPD capi template"
					prDescription := "This PR creates a new CAPD Kubernetes cluster"

					By(fmt.Sprintf("And I run 'mccp templates render cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --create-pr --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", clusterName, namespace, k8version, prBranch, prTitle, prCommit, CAPI_ENDPOINT_URL), func() {
						command := exec.Command(MCCP_BIN_PATH, "templates", "render", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
							"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
							"--create-pr", "--pr-branch", prBranch, "--pr-title", prTitle, "--pr-commit-message", prCommit, "--pr-description", prDescription,
							"--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
					})

					By("Then I should see pull request created to management cluster", func() {
						output := session.Wait().Out.Contents()

						re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
						match := re.FindSubmatch([]byte(output))
						Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					})

					By(fmt.Sprintf("Then I run 'mccp clusters list --endpoint %s'", CAPI_ENDPOINT_URL), func() {
						command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
					})

					By("And I should see cluster status as 'pullRequestCreated'", func() {
						output := session.Wait().Out.Contents()
						Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

						re := regexp.MustCompile(fmt.Sprintf(`%s\s+pullRequestCreated`, clusterName))
						Eventually((re.Find(output))).ShouldNot(BeNil())
					})

					By("Then I should merge the pull request to start cluster provisioning", func() {
						mccpTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
					})

					By("And I should see cluster status changes to 'Provisioned'", func() {
						output := func() string {
							command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
							session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
							Expect(err).ShouldNot(HaveOccurred())
							return string(session.Wait().Out.Contents())

						}
						Eventually(output, ASSERTION_2MINUTE_TIME_OUT, CLI_POLL_INTERVAL).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName)))
					})

					By(fmt.Sprintf("Then I run 'mccp clusters get cli-end-to-end-capd-cluster --kubeconfig --endpoint %s'", CAPI_ENDPOINT_URL), func() {
						output := func() string {
							command := exec.Command(MCCP_BIN_PATH, "clusters", "get", clusterName, "--kubeconfig", "--endpoint", CAPI_ENDPOINT_URL)
							session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
							Expect(err).ShouldNot(HaveOccurred())

							return string(session.Wait().Out.Contents())

						}
						Eventually(output, ASSERTION_2MINUTE_TIME_OUT, CLI_POLL_INTERVAL).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)))
					})
				}

				// Parameter values
				clusterName := "cli-end-to-end-capd-cluster-11"
				namespace := "default"
				k8version := "1.19.7"
				// Creating two capd clusters
				createCluster(clusterName, namespace, k8version)
				clusterName2 := "cli-end-to-end-capd-cluster-21"
				createCluster(clusterName2, namespace, k8version)

				// Deleting first cluster
				prBranch := fmt.Sprintf("%s-delete", clusterName)
				prTitle := "CAPD delete pull request"
				prCommit := "CAPD capi template deletion"
				prDescription := "This PR deletes CAPD Kubernetes cluster"

				By(fmt.Sprintf("Then I run 'mccp clusters delete cli-end-to-end-capd-cluster --pr-branch %s --pr-title %s --pr-commit-message %s --endpoint %s", prBranch, prTitle, prCommit, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(MCCP_BIN_PATH, "clusters", "delete", clusterName,
						"--pr-branch", prBranch, "--pr-title", prTitle, "--pr-commit-message", prCommit, "--pr-description", prDescription,
						"--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				var prUrl string
				By("Then I should see delete pull request created to management cluster", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`Created pull request for clusters deletion:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request for deleting cluster")
					prUrl = string(match[1])
				})

				By("And I should veriyfy the delete pull request in the cluster config repository", func() {
					pullRequest := mccpTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the delete pull request manifests are not present in the cluster config repository", func() {
					mccpTestRunner.PullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					output := func() string {
						command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
						return string(session.Wait().Out.Contents())

					}
					Eventually(output, ASSERTION_2MINUTE_TIME_OUT, CLI_POLL_INTERVAL).Should(MatchRegexp(fmt.Sprintf(`%s\s+pullRequestCreated`, clusterName)))
				})

				By(fmt.Sprintf("And I should see the '%s' cluster status remains unchanged as 'clusterFound'", clusterName2), func() {
					output := func() string {
						command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
						return string(session.Wait().Out.Contents())

					}
					Eventually(output).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName2)))
				})

				By("Then I should merge the delete pull request to delete cluster", func() {
					mccpTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
				})
			})
		})

	})
}
