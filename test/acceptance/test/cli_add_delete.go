package acceptance

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func DescribeCliAddDelete(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops add Tests", func() {
		templateFiles := []string{}
		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(fileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
			})

			By("And the Cluster service is healthy", func() {
				gitopsTestRunner.CheckClusterService(CAPI_ENDPOINT_URL)
			})
		})

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			JustAfterEach(func() {
				deleteRepo(gitProviderEnv)

			})

			It("Verify gitops can set template parameters by specifying multiple parameters --set key=value --set key=value", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
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
					Eventually(string(output)).Should(MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					Eventually(string(output)).Should(MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					Eventually(string(output)).Should(MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
				})
			})

			It("Verify gitops can set template parameters by separate values with commas key1=val1,key2=val2", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
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
					Eventually(string(output)).Should(MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					Eventually(string(output)).Should(MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))

					// Verifying MachineDeployment object of the template for updated  parameter values
					Eventually(string(output)).Should(MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
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
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --url %s --commit-message %s --description %s --endpoint %s",
					clusterName, namespace, k8version, prBranch, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", k8version),
						"--branch", prBranch, "--url", GIT_REPOSITORY_URL, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see an error for required argument to create pull request", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(`Error: unable to create pull request.*template name must be specified`))
				})
			})

			It("Verify gitops can create pull requests to management cluster", func() {
				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = initAndCreateEmptyRepo(gitProviderEnv, true)
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
					capdTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
					templateFiles = append(capdTemplateFile, eksTemplateFile...)
				})

				By(fmt.Sprintf("And I run 'gitops add cluster --form-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --url %s --commit-message %s --description %s  --endpoint %s",
					capdClusterName, capdNamespace, capdK8version, capdPRBranch, capdPRTitle, GIT_REPOSITORY_URL, capdPRCommit, capdPRDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-development-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", capdClusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", capdNamespace), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", capdK8version),
						"--branch", capdPRBranch, "--title", capdPRTitle, "--url", GIT_REPOSITORY_URL, "--commit-message", capdPRCommit, "--description", capdPRDescription, "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				var capdPRUrl string
				By("Then I should see pull request created to management cluster", func() {
					output := session.Wait().Out.Contents()

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					capdPRUrl = string(match[1])
				})

				By("And I should veriyfy the capd pull request in the cluster config repository", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					Expect(createPRUrl).Should(Equal(capdPRUrl))
				})

				By("And the capd manifest is present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, capdPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, capdClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
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

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template eks-fargate-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s",
					eksClusterName, eksNamespace, eksRegion, eksK8version, eksPRBranch, eksPRTitle, GIT_REPOSITORY_URL, eksPRCommit, eksPRDescription, CAPI_ENDPOINT_URL), func() {

					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "eks-fargate-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", eksClusterName),
						"--set", fmt.Sprintf("AWS_REGION=%s", eksRegion), "--set", fmt.Sprintf("KUBERNETES_VERSION=%s", eksK8version),
						"--set", fmt.Sprintf("AWS_SSH_KEY_NAME=%s", eksSshKeyName), "--set", fmt.Sprintf("NAMESPACE=%s", eksNamespace),
						"--branch", eksPRBranch, "--title", eksPRTitle, "--url", GIT_REPOSITORY_URL, "--commit-message", eksPRCommit, "--description", eksPRDescription, "--endpoint", CAPI_ENDPOINT_URL)
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
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					Expect(createPRUrl).Should(Equal(eksPRUrl))
				})

				By("And the eks manifest is present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, eksPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, eksClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see cluster status as 'Creation PR'", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

					re := regexp.MustCompile(fmt.Sprintf(`%s\s+Creation PR`, eksClusterName))
					Eventually((re.Find(output))).ShouldNot(BeNil())
					re = regexp.MustCompile(fmt.Sprintf(`%s\s+Creation PR`, capdClusterName))
					Eventually((re.Find(output))).ShouldNot(BeNil())
				})
			})

			It("Verify giops can not create pull request to management cluster using existing branch", func() {
				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					createGitRepoBranch(repoAbsolutePath, branchName)
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

				By(fmt.Sprintf("And I run 'gitops add cluster --from-template cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s  --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s",
					clusterName, namespace, branchName, prTitle, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--from-template", "cluster-template-0", "--set", fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
						"--set", fmt.Sprintf("NAMESPACE=%s", namespace),
						"--branch", branchName, "--title", prTitle, "--url", GIT_REPOSITORY_URL, "--commit-message", prCommit, "--description", prDescription, "--endpoint", CAPI_ENDPOINT_URL)
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
			kubeconfigPath := path.Join(os.Getenv("HOME"))
			appName := "management"
			appPath := "./management"
			capdClusterNames := []string{"cli-end-to-end-capd-cluster-1", "cli-end-to-end-capd-cluster-2"}

			var output string
			var errOutput string

			JustBeforeEach(func() {
				_ = deleteFile([]string{kubeconfigPath})

				log.Println("Connecting cluster to itself")
				initializeWebdriver(DEFAULT_UI_URL)
				leaf := LeafSpec{
					Status:          "Ready",
					IsWKP:           false,
					AlertManagerURL: "",
					KubeconfigPath:  "",
				}
				connectACluster(webDriver, gitopsTestRunner, leaf)
			})

			JustAfterEach(func() {
				_ = deleteFile([]string{kubeconfigPath})
				// Force delete capicluster incase delete PR fails to delete to free resources
				removeGitopsCapiClusters(appName, capdClusterNames, GITOPS_DEFAULT_NAMESPACE)

				deleteRepo(gitProviderEnv)

				log.Println("Deleting all the wkp agents")
				_ = gitopsTestRunner.KubectlDeleteAllAgents([]string{})
				_ = gitopsTestRunner.ResetDatabase()
				gitopsTestRunner.VerifyWegoPodsRunning()
			})

			It("@capd Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {

				By("Check wge is all running", func() {
					gitopsTestRunner.VerifyWegoPodsRunning()
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				By("And I install gitops to my active cluster", func() {
					Expect(fileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
				})

				By("And I install profiles (enhanced helm chart)", func() {
					installProfiles("weaveworks-charts", GITOPS_DEFAULT_NAMESPACE)
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd-observability.yaml")
				})

				createCluster := func(clusterName string, namespace string, k8version string) {
					//Pull request values
					prBranch := fmt.Sprintf("br-%s", clusterName)
					prTitle := "CAPD pull request"
					prCommit := "CAPD capi template"
					prDescription := "This PR creates a new CAPD Kubernetes cluster"

					By(fmt.Sprintf("And I run 'gitops add cluster --from-template cluster-template-development-observability-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=1 --set WORKER_MACHINE_COUNT=1 --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s",
						clusterName, namespace, k8version, prBranch, prTitle, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
						output, errOutput = runCommandAndReturnStringOutput(fmt.Sprintf(`%s add cluster --from-template cluster-template-development-observability-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=1 --set WORKER_MACHINE_COUNT=1 --branch "%s" --title "%s" --url %s --commit-message "%s" --description "%s" --endpoint %s`,
							GITOPS_BIN_PATH, clusterName, namespace, k8version, prBranch, prTitle, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL))
						Expect(errOutput).Should(BeEmpty())
					})

					By("Then I should see pull request created to management cluster", func() {
						re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
						match := re.FindSubmatch([]byte(output))
						Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request")
					})

					By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", CAPI_ENDPOINT_URL), func() {
						output, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`%s get clusters --endpoint %s`, GITOPS_BIN_PATH, CAPI_ENDPOINT_URL))
					})

					By("And I should see cluster status as 'Creation PR'", func() {
						Eventually(string(output)).Should(MatchRegexp(`NAME\s+STATUS`))

						re := regexp.MustCompile(fmt.Sprintf(`%s\s+Creation PR`, clusterName))
						Eventually((re.Find([]byte(output)))).ShouldNot(BeNil())
					})

					By("Then I should merge the pull request to start cluster provisioning", func() {
						createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
						mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					})

					By("And I add a test kustomization file to the management appliction (because flux doesn't reconcile empty folders on deletion)", func() {
						pullGitRepo(repoAbsolutePath)
						_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cp -f ../../utils/data/test_kustomization.yaml %s", path.Join(repoAbsolutePath, appPath)))
						gitUpdateCommitPush(repoAbsolutePath, "")
					})

					By("And I run gitops add app 'management' command", func() {
						if listGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE) == "" {
							addCommand := fmt.Sprintf("add app . --path=%s  --name=%s  --auto-merge=true", appPath, appName)
							runWegoAddCommand(repoAbsolutePath, addCommand, GITOPS_DEFAULT_NAMESPACE)
						} else {
							log.Printf("Application '%s' alreaded exists", appName)
						}
					})

					By("And I should see cluster status changes to 'clusterFound'", func() {
						verifyWegoAddCommand(appName, GITOPS_DEFAULT_NAMESPACE)
						clusterFound := func() string {
							output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get clusters --endpoint %s`, GITOPS_BIN_PATH, CAPI_ENDPOINT_URL))
							return output
						}
						Eventually(clusterFound, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName)))
					})

					By(fmt.Sprintf("Then I run '%s get cluster %s --kubeconfig --endpoint %s'", GITOPS_BIN_PATH, clusterName, CAPI_ENDPOINT_URL), func() {
						kubeConfigFound := func() string {
							output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get cluster %s --kubeconfig --endpoint %s | tee %s`, GITOPS_BIN_PATH, clusterName, CAPI_ENDPOINT_URL, kubeconfigPath))
							return output

						}
						Eventually(kubeConfigFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)))
					})
				}

				// Parameter values
				clusterName := capdClusterNames[0]
				namespace := "default"
				k8version := "1.23.0"
				// Creating two capd clusters
				createCluster(clusterName, namespace, k8version)

				By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", clusterName), func() {
					verifyCapiClusterKubeconfig(kubeconfigPath, clusterName)
				})

				By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", clusterName), func() {
					verifyCapiClusterHealth(kubeconfigPath, clusterName)
				})

				clusterName2 := capdClusterNames[1]
				createCluster(clusterName2, namespace, k8version)

				// Deleting first cluster
				prBranch := fmt.Sprintf("%s-delete", clusterName)
				prTitle := "CAPD delete pull request"
				prCommit := "CAPD capi template deletion"
				prDescription := "This PR deletes CAPD Kubernetes cluster"

				By(fmt.Sprintf("Then I run '%s delete cluster %s --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s",
					GITOPS_BIN_PATH, clusterName, prBranch, prTitle, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL), func() {
					output, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`%s delete cluster %s --branch %s --title "%s" --url %s --commit-message "%s" --description "%s" --endpoint %s`,
						GITOPS_BIN_PATH, clusterName, prBranch, prTitle, GIT_REPOSITORY_URL, prCommit, prDescription, CAPI_ENDPOINT_URL))
				})

				By("Then I should see delete pull request created to management cluster", func() {
					re := regexp.MustCompile(`Created pull request for clusters deletion:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(output))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request for deleting cluster")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					clusterDelete := func() string {
						output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get cluster %s --endpoint %s`, GITOPS_BIN_PATH, clusterName, CAPI_ENDPOINT_URL))
						return output

					}
					Eventually(clusterDelete, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+Deletion PR`, clusterName)))
				})

				var deletePRUrl string
				By("And I should veriyfy the delete pull request in the cluster config repository", func() {
					deletePRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				})

				By("Then I should merge the delete pull request to delete cluster", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, deletePRUrl)
				})

				By("And the delete pull request manifests are not present in the cluster config repository", func() {
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("And I should see the '%s' cluster status remains unchanged as 'clusterFound'", clusterName2), func() {
					clusterFound := func() string {
						output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get cluster %s --endpoint %s`, GITOPS_BIN_PATH, clusterName2, CAPI_ENDPOINT_URL))
						return output

					}
					Eventually(clusterFound).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName2)))
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
