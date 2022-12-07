package acceptance

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func DescribeCliAddDelete(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Gitops add Tests", ginkgo.Label("cli"), func() {
		var stdOut string
		var stdErr string
		var repoAbsolutePath string
		templateFiles := []string{}
		clusterPath := "./clusters/management/clusters"

		ginkgo.AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		ginkgo.Context("[CLI] When Capi Templates are available in the cluster", func() {

			ginkgo.It("Verify gitops can set template parameters by specifying multiple parameters --set key=value --set key=value", ginkgo.Label("git"), func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"
				controlPlaneMachineCount := "2"
				workerMachineCount := "3"

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --from-template capd-cluster-template-0 --dry-run --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=%s --set WORKER_MACHINE_COUNT=%s --set INSTALL_CRDS=true`,
					clusterName, namespace, k8version, controlPlaneMachineCount, workerMachineCount)
				stdOut, stdErr = runGitopsCommand(cmd)

				ginkgo.By("Then I should see template preview with updated parameter values", func() {
					// Verifying cluster object of tbe template for updated  parameter values
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+replicas: %[3]v[\s\w\d/:.-]+version: %[4]v`,
						clusterName, namespace, controlPlaneMachineCount, k8version))

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+replicas: %[3]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[4]v`,
						clusterName, namespace, workerMachineCount, k8version))
				})
			})

			ginkgo.It("Verify gitops can set template parameters by separate values with commas key1=val1,key2=val2", ginkgo.Label("git"), func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.23.6"
				controlPlaneMachineCount := "1"
				workerMachineCount := "2"

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --from-template capd-cluster-template-0 --dry-run --set CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s,CONTROL_PLANE_MACHINE_COUNT=%s,WORKER_MACHINE_COUNT=%s --set INSTALL_CRDS=true`,
					clusterName, namespace, k8version, controlPlaneMachineCount, workerMachineCount)
				stdOut, stdErr = runGitopsCommand(cmd)

				ginkgo.By("Then I should see template preview with updated parameter values", func() {

					// Verifying cluster object of tbe template for updated  parameter values
					gomega.Eventually(string(stdOut)).Should(gomega.MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					gomega.Eventually(string(stdOut)).Should(gomega.MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))

					// Verifying MachineDeployment object of the template for updated  parameter values
					gomega.Eventually(string(stdOut)).Should(gomega.MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
				})
			})

			ginkgo.It("Verify gitops reports an error when trying to create pull request with missing --from-template argument", ginkgo.Label("git"), func() {
				// Parameter values
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				//Pull request values
				prBranch := "feature-capd"
				prCommit := "First capd capi template"
				prDescription := "This PR creates a new capd Kubernetes cluster"

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --url %s --commit-message %s --description %s`, clusterName, namespace, k8version, prBranch, git_repository_url, prCommit, prDescription)
				stdOut, stdErr = runGitopsCommand(cmd)

				ginkgo.By("Then I should see an error for required argument to create pull request", func() {
					gomega.Eventually(stdErr).Should(gomega.MatchRegexp(`Error: unable to create pull request.*template name must be specified`))
				})
			})
		})

		ginkgo.Context("[CLI] When Capi Templates are available in the cluster to create pull requests", func() {

			ginkgo.BeforeEach(func() {
				repoAbsolutePath = configRepoAbsolutePath(gitProviderEnv)
			})

			ginkgo.JustAfterEach(func() {
				cleanGitRepository(clusterPath)
			})

			ginkgo.It("Verify gitops can create pull requests to management cluster", ginkgo.Label("smoke", "git"), func() {
				// CAPD Parameter values
				capdClusterName := "my-capd-cluster2"
				capdNamespace := "gitops-dev"
				capdK8version := "1.19.7"
				controlPlaneMachineCount := "2"
				workerMachineCount := "3"

				//CAPD Pull request values
				capdPRBranch := "cli-feature-capd"
				capdPRTitle := "My first pull request"
				capdPRCommit := "First capd capi template"
				capdPRDescription := "This PR creates a new capd Kubernetes cluster"

				ginkgo.By("Apply/Install CAPITemplate", func() {
					capdTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/aws/cluster-template-eks-fargate.yaml")
					templateFiles = append(capdTemplateFile, eksTemplateFile...)
				})

				cmd := fmt.Sprintf(`add cluster --from-template capd-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=%s --set WORKER_MACHINE_COUNT=%s --set INSTALL_CRDS=true --branch %s --title %s --url %s --commit-message %s --description %s`,
					capdClusterName, capdNamespace, capdK8version, controlPlaneMachineCount, workerMachineCount, capdPRBranch, capdPRTitle, git_repository_url, capdPRCommit, capdPRDescription)
				stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

				var capdPRUrl string
				ginkgo.By("Then I should see pull request created to management cluster", func() {
					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
					gomega.Eventually(match).ShouldNot(gomega.BeNil(), "Failed to Create pull request")
					capdPRUrl = string(match[1])
				})

				ginkgo.By("And I should veriyfy the capd pull request in the cluster config repository", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					gomega.Expect(createPRUrl).Should(gomega.Equal(capdPRUrl))
				})

				ginkgo.By("And the capd manifest is present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, capdPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/clusters/management/clusters/%s/%s.yaml", repoAbsolutePath, capdNamespace, capdClusterName))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Cluster config can not be found.")
				})

				// EKS Parameter values
				eksClusterName := "my-eks-cluster"
				eksRegion := "eu-west-3"
				eksK8version := "1.19.8"
				eksSshKeyName := "my-aws-ssh-key"
				eksNamespace := "default"

				//EKS Pull request values
				eksPRBranch := "cli-feature-eks"
				eksPRTitle := "My second pull request"
				eksPRCommit := "First eks capi template"
				eksPRDescription := "This PR creates a new eks Kubernetes cluster"

				cmd = fmt.Sprintf(`add cluster --from-template capa-cluster-template-eks-fargate-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --branch %s --title %s --url %s --commit-message %s --description %s`,
					eksClusterName, eksRegion, eksK8version, eksSshKeyName, eksNamespace, eksPRBranch, eksPRTitle, git_repository_url, eksPRCommit, eksPRDescription)
				stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

				var eksPRUrl string
				ginkgo.By("Then I should see pull request created for eks to management cluster", func() {

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
					gomega.Eventually(match).ShouldNot(gomega.BeNil(), "Failed to Create pull request")
					eksPRUrl = string(match[1])
				})

				ginkgo.By("And I should veriyfy the eks pull request in the cluster config repository", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					gomega.Expect(createPRUrl).Should(gomega.Equal(eksPRUrl))
				})

				ginkgo.By("And the eks manifest is present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, eksPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, eksNamespace, eksClusterName+".yaml"))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Cluster config can not be found.")
				})

				stdOut, stdErr = runGitopsCommand(`get clusters`)

			})

			ginkgo.It("Verify giops can not create pull request to management cluster using existing branch", ginkgo.Label("git"), func() {
				branchName := "cli-test-branch"
				ginkgo.By("And create new git repository branch", func() {
					createGitRepoBranch(repoAbsolutePath, branchName)
				})

				// Parameter values
				clusterName := "my-dev-cluster"
				namespace := "gitops-dev"
				k8Version := "v1.23.3"
				awsRegion := "us-east-1"
				controlPlaneMachineCount := "3"
				workerMachineCount := "3"
				costEstimationFilter := `tenancy=Dedicated`

				//Pull request values
				prTitle := "My dev pull request"
				prCommit := "First dev capi template"
				prDescription := "This PR creates a new dev Kubernetes cluster"

				// Checkout repo main branch in case of test failure
				defer checkoutRepoBranch(repoAbsolutePath, "main")

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/aws/cluster-template-ec2.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --from-template capa-cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=%s --set WORKER_MACHINE_COUNT=%s --set COST_ESTIMATION_FILTERS=%s --branch %s --title %s --url %s --commit-message %s --description %s`,
					clusterName, namespace, awsRegion, k8Version, controlPlaneMachineCount, workerMachineCount, costEstimationFilter, branchName, prTitle, git_repository_url, prCommit, prDescription)
				stdOut, stdErr = runGitopsCommand(cmd)

				ginkgo.By("Then I should not see pull request to be created", func() {
					gomega.Eventually(stdErr).Should(gomega.MatchRegexp(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})
	})
}
