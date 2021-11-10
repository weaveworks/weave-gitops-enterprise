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
	"github.com/onsi/gomega/gexec"
)

func DescribeCliAddDelete(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops add Tests", func() {

		GITOPS_BIN_PATH := GetGitopsBinPath()
		CAPI_ENDPOINT_URL := GetCapiEndpointUrl()

		templateFiles := []string{}
		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(FileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
			})

			By("And the Cluster service is healthy", func() {
				gitopsTestRunner.CheckClusterService(GetCapiEndpointUrl())
			})
		})

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify gitops can set template parameters by specifying multiple parameters --set key=value --set key=value", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --endpoint %s", clusterName, namespace, k8version, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-development-0", "--dry-run", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version), "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template preview with updated parameter values", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of tbe template for updated  parameter values
					re := regexp.MustCompile(fmt.Sprintf(`kind: Cluster\s+metadata:\s+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, "default"))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, "default", k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, "default", k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})

			It("Verify gitops can set template parameters by separate values with commas key1=val1,key2=val2", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --form-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s  --endpoint %s", clusterName, namespace, k8version, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-development-0", "--dry-run",
						"--set", fmt.Sprintf("CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s", clusterName, namespace, k8version), "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template preview with updated parameter values", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of tbe template for updated  parameter values
					re := regexp.MustCompile(fmt.Sprintf(`kind: Cluster\s+metadata:\s+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, "default"))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, "default", k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())

					// Verifying MachineDeployment object of the template for updated  parameter values
					re = regexp.MustCompile(fmt.Sprintf(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, "default", k8version))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})

			It("Verify gitops reports an error when trying to create pull request with missing --from-template argument", func() {
				// Parameter values
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --commit-message %s --description %s --endpoint %s",
					clusterName, namespace, k8version, prBranch, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
						"--branch", prBranch, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see an error for required argument to create pull request", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(`Error: unable to create pull request.*template name must be specified`))
				})
			})

			It("Verify gitops can create pull requests to management cluster", func() {

				defer gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer func() {
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				}()

				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				// CAPD Parameter values
				capdClusterName := "my-capd-cluster2"
				capdNamespace := "gitops-dev"
				capdK8version := "1.19.7"

				//CAPD Pull request values
				capdPRBranch := "feature-capd"
				capdPRTitle := "My first pull request"
				capdPRCommit := "First capd capi template"
				capdPRDescription := "This PR creates a new capd Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
					capdTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
					templateFiles = append(capdTemplateFile, eksTemplateFile...)
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --form-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --commit-message %s --description %s  --endpoint %s",
					capdClusterName, capdNamespace, capdK8version, capdPRBranch, capdPRTitle, capdPRCommit, capdPRDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", capdClusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", capdNamespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", capdK8version),
						"--branch", capdPRBranch, "--title", capdPRTitle, "--commit-message", capdPRCommit, "--description", capdPRDescription, "--endpoint", CAPI_ENDPOINT_URL)
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
					pullRequest := gitopsTestRunner.ListPullRequest(repoAbsolutePath)
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

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template eks-fargate-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --commit-message %s --description %s --endpoint %s",
					eksClusterName, eksNamespace, eksRegion, eksK8version, eksPRBranch, eksPRTitle, eksPRCommit, eksPRDescription, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "eks-fargate-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", eksClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", eksRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", eksK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", eksSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", eksNamespace),
						"--branch", eksPRBranch, "--title", eksPRTitle, "--commit-message", eksPRCommit, "--description", eksPRDescription, "--endpoint", CAPI_ENDPOINT_URL)
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
					pullRequest := gitopsTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(eksPRTitle))
					Expect(pullRequest[1]).Should(Equal(eksPRBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(eksPRUrl))
				})

				By("And the capd manifest is present in the cluster config repository", func() {
					gitopsTestRunner.PullBranch(repoAbsolutePath, capdPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, capdClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By("And the eks manifest is present in the cluster config repository", func() {
					gitopsTestRunner.PullBranch(repoAbsolutePath, eksPRBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, eksClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
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

			It("Verify giops can not create pull request to management cluster using existing branch", func() {

				defer gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer func() {
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				}()

				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					gitopsTestRunner.CreateGitRepoBranch(repoAbsolutePath, branchName)
				})

				// Parameter values
				clusterName := "my-dev-cluster"
				namespace := "gitops-dev"

				//Pull request values
				prTitle := "My dev pull request"
				prCommit := "First dev capi template"
				prDescription := "This PR creates a new dev Kubernetes cluster"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-capitemplate.yaml")
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s  --branch %s --title %s --commit-message %s --description %s --endpoint %s",
					clusterName, namespace, branchName, prTitle, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace),
						"--branch", branchName, "--title", prTitle, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should not see pull request to be created", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})

		Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {
			It("Verify gitops can use the matching selected credential for cluster creation", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AWS")
				defer gitopsTestRunner.DeleteIPCredentials("AZURE")

				By("Apply/Install CAPITemplates", func() {
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
					azureTemplateFiles := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
					templateFiles = append(azureTemplateFiles, eksTemplateFile...)
				})

				By("And create AWS credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
				})

				By("And create AZURE credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				// AWS Parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials aws-test-identity --dry-run --endpoint %s",
					awsClusterName, awsNamespace, awsRegion, awsK8version, awsSshKeyName, awsControlMAchineType, awsNodeMAchineType, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "aws-cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", awsClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", awsRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", awsK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", awsSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", awsNamespace),
						"--set", "CONTROL_PLANE_MACHINE_COUNT=2", "--set", fmt.Sprintf("AWS_CONTROL_PLANE_MACHINE_TYPE=%s", awsControlMAchineType),
						"--set", "WORKER_MACHINE_COUNT=3", "--set", fmt.Sprintf("AWS_NODE_MACHINE_TYPE=%s", awsNodeMAchineType),
						"--set-credentials", "aws-test-identity", "--dry-run", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see preview containing identity reference added in the template", func() {
					output := session.Wait().Out.Contents()

					// Verifying cluster object of the template for added credential reference
					re := regexp.MustCompile(fmt.Sprintf(`kind: AWSCluster\s+metadata:[\s\w\d-.:/]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterStaticIdentity\s+name: aws-test-identity`, awsClusterName))

					Eventually((re.Find(output))).ShouldNot(BeNil(), "Failed to find identity reference in preview pull request AWSCluster object")
				})
			})

			It("Verify gitops restrict user from using wrong credentials for infrastructure provider", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AZURE")

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
				})

				By("And create AZURE credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				// AWS Parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials azure-cluster-identity --dry-run --endpoint %s",
					awsClusterName, awsNamespace, awsRegion, awsK8version, awsControlMAchineType, awsNodeMAchineType, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "aws-cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", awsClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", awsRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", awsK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", awsSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", awsNamespace),
						"--set", "CONTROL_PLANE_MACHINE_COUNT=2", "--set", fmt.Sprintf("AWS_CONTROL_PLANE_MACHINE_TYPE=%s", awsControlMAchineType),
						"--set", "WORKER_MACHINE_COUNT=3", "--set", fmt.Sprintf("AWS_NODE_MACHINE_TYPE=%s", awsNodeMAchineType),
						"--set-credentials", "azure-cluster-identity", "--dry-run", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				// FIXME - User should get some warning or error as well for chossing wrong credential/identity for the infrastructure provider

				By("Then I should see preview without identity reference added to the template", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: [\s\w\d-.:/]+%s\s+---`, awsSshKeyName)), "Identity reference should not be found in preview pull request AWSCluster object")
				})
			})
		})

		Context("[CLI] When leaf cluster pull request is available in the management cluster", func() {
			appName := "management"
			appPath := "./management"
			capdClusterNames := []string{"cli-end-to-end-capd-cluster-1", "cli-end-to-end-capd-cluster-2"}

			JustBeforeEach(func() {
				log.Println("Connecting cluster to itself")
				InitializeWebdriver(GetWGEUrl())
				leaf := LeafSpec{
					Status:          "Ready",
					IsWKP:           false,
					AlertManagerURL: "",
					KubeconfigPath:  "",
				}
				connectACluster(webDriver, gitopsTestRunner, leaf)
			})

			JustAfterEach(func() {
				// Force delete capicluster incase delete PR fails to delete to free resources
				RemoveGitopsCapiClusters(appName, capdClusterNames, GITOPS_DEFAULT_NAMESPACE)

				gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				log.Println("Deleting all the wkp agents")
				_ = gitopsTestRunner.KubectlDeleteAllAgents([]string{})
				_ = gitopsTestRunner.ResetDatabase()
				gitopsTestRunner.VerifyWegoPodsRunning()
			})

			It("@capd Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {
				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And I install gitops to my active cluster", func() {
					Expect(FileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
					InstallAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE)
				})

				addCommand := fmt.Sprintf("add app . --path=./management  --name=%s  --auto-merge=true", appName)
				By(fmt.Sprintf("And I run gitops add app command '%s in namespace %s from dir %s'", addCommand, GITOPS_DEFAULT_NAMESPACE, repoAbsolutePath), func() {
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, addCommand))
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
					Expect(string(session.Err.Contents())).Should(BeEmpty())
				})

				By("And I install Docker provider infrastructure", func() {
					installInfrastructureProvider("docker")
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				createCluster := func(clusterName string, namespace string, k8version string) {
					//Pull request values
					prBranch := fmt.Sprintf("br-%s", clusterName)
					prTitle := "CAPD pull request"
					prCommit := "CAPD capi template"
					prDescription := "This PR creates a new CAPD Kubernetes cluster"

					By(fmt.Sprintf("And I run 'gitops add cluster --from-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --commit-message %s --description %s --endpoint %s",
						clusterName, namespace, k8version, prBranch, prTitle, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
						command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
							"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
							"--branch", prBranch, "--title", prTitle, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
					})

					By("Then I should see pull request created to management cluster", func() {
						output := session.Wait().Out.Contents()

						re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
						match := re.FindSubmatch([]byte(output))
						Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					})

					By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", CAPI_ENDPOINT_URL), func() {
						command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
					})

					By("And I should see cluster status as 'pullRequestCreated'", func() {
						output := session.Wait().Out.Contents()
						Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

						re := regexp.MustCompile(fmt.Sprintf(`%s\s+pullRequestCreated`, clusterName))
						Eventually((re.Find(output))).ShouldNot(BeNil())
					})

					By("And I add a test kustomization file to the pull request (needs it because flux doesn't work with empty folders on deletion)", func() {
						gitopsTestRunner.PullBranch(repoAbsolutePath, prBranch)
						_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cp -f ../../utils/data/test_kustomization.yaml %s", path.Join(repoAbsolutePath, appPath)))
						GitSetUpstream(repoAbsolutePath, prBranch)
						GitUpdateCommitPush(repoAbsolutePath)
					})

					By("Then I should merge the pull request to start cluster provisioning", func() {
						gitopsTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
					})

					By("And I should see cluster status changes to 'Provisioned'", func() {
						output := func() string {
							command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
							session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
							Expect(err).ShouldNot(HaveOccurred())
							return string(session.Wait().Out.Contents())

						}
						Eventually(output, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName)))
					})

					By(fmt.Sprintf("Then I run 'gitops get cluster cli-end-to-end-capd-cluster --kubeconfig --endpoint %s'", CAPI_ENDPOINT_URL), func() {
						output := func() string {
							command := exec.Command(GITOPS_BIN_PATH, "get", "cluster", clusterName, "--kubeconfig", "--endpoint", CAPI_ENDPOINT_URL)
							session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
							Expect(err).ShouldNot(HaveOccurred())

							return string(session.Wait().Out.Contents())

						}
						Eventually(output, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)))
					})
				}

				// Parameter values
				clusterName := capdClusterNames[0]
				namespace := "default"
				k8version := "1.19.7"
				// Creating two capd clusters
				createCluster(clusterName, namespace, k8version)
				clusterName2 := capdClusterNames[1]
				createCluster(clusterName2, namespace, k8version)

				// Deleting first cluster
				prBranch := fmt.Sprintf("%s-delete", clusterName)
				prTitle := "CAPD delete pull request"
				prCommit := "CAPD capi template deletion"
				prDescription := "This PR deletes CAPD Kubernetes cluster"

				By(fmt.Sprintf("Then I run 'gitops delete cluster cli-end-to-end-capd-cluster --branch %s --title %s --commit-message %s --description %s --endpoint %s",
					prBranch, prTitle, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "delete", "cluster", clusterName,
						"--branch", prBranch, "--title", prTitle, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
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
					pullRequest := gitopsTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the delete pull request manifests are not present in the cluster config repository", func() {
					gitopsTestRunner.PullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					output := func() string {
						command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
						return string(session.Wait().Out.Contents())

					}
					Eventually(output, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+pullRequestCreated`, clusterName)))
				})

				By(fmt.Sprintf("And I should see the '%s' cluster status remains unchanged as 'clusterFound'", clusterName2), func() {
					output := func() string {
						command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ShouldNot(HaveOccurred())
						return string(session.Wait().Out.Contents())

					}
					Eventually(output).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName2)))
				})

				By("Then I should merge the delete pull request to delete cluster", func() {
					gitopsTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster deleted", clusterName), func() {
					clusterFound := func() error {
						return runCommandPassThrough([]string{}, "kubectl", "get", "cluster", clusterName)
					}
					Eventually(clusterFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(HaveOccurred())
				})
			})
		})
	})
}
