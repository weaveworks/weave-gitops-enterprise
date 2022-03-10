package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func DescribeCliAddDelete(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops add Tests", func() {
		var stdOut string
		var stdErr string
		var repoAbsolutePath string
		templateFiles := []string{}

		// Using self signed certs, all `gitops get clusters` etc commands should use insecure tls connections
		insecureFlag := "--insecure-skip-tls-verify"

		BeforeEach(func() {
			repoAbsolutePath = configRepoAbsolutePath(gitProviderEnv)

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(fileExists(gitops_bin_path)).To(BeTrue(), fmt.Sprintf("%s can not be found.", gitops_bin_path))
			})

			By("And the Cluster service is healthy", func() {
				gitopsTestRunner.CheckClusterService(capi_endpoint_url)
			})
		})

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("@git Verify gitops can set template parameters by specifying multiple parameters --set key=value --set key=value", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				cmd := fmt.Sprintf(`%s add cluster --from-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --endpoint %s`, gitops_bin_path, clusterName, namespace, k8version, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should see template preview with updated parameter values", func() {
					// Verifying cluster object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
				})
			})

			It("@git Verify gitops can set template parameters by separate values with commas key1=val1,key2=val2", func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				cmd := fmt.Sprintf(`%s add cluster --from-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s --endpoint %s %s`, gitops_bin_path, clusterName, namespace, k8version, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should see template preview with updated parameter values", func() {

					// Verifying cluster object of tbe template for updated  parameter values
					Eventually(string(stdOut)).Should(MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					Eventually(string(stdOut)).Should(MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))

					// Verifying MachineDeployment object of the template for updated  parameter values
					Eventually(string(stdOut)).Should(MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[3]v`,
						clusterName, namespace, k8version))
				})
			})

			It("@git Verify gitops reports an error when trying to create pull request with missing --from-template argument", func() {
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

				cmd := fmt.Sprintf(`%s add cluster --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --url %s --commit-message %s --description %s --endpoint %s %s`, gitops_bin_path, clusterName, namespace, k8version, prBranch, git_repository_url, prCommit, prDescription, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should see an error for required argument to create pull request", func() {
					Eventually(stdErr).Should(MatchRegexp(`Error: unable to create pull request.*template name must be specified`))
				})
			})

			It("@smoke @git Verify gitops can create pull requests to management cluster", func() {
				By("When I create a private repository for cluster configs", func() {
					initAndCreateEmptyRepo(gitProviderEnv, true)
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

				cmd := fmt.Sprintf(`%s add cluster --from-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s %s`,
					gitops_bin_path, capdClusterName, capdNamespace, capdK8version, capdPRBranch, capdPRTitle, git_repository_url, capdPRCommit, capdPRDescription, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				var capdPRUrl string
				By("Then I should see pull request created to management cluster", func() {
					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
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

				cmd = fmt.Sprintf(`%s add cluster --from-template eks-fargate-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s %s`,
					gitops_bin_path, eksClusterName, eksRegion, eksK8version, eksSshKeyName, eksNamespace, eksPRBranch, eksPRTitle, git_repository_url, eksPRCommit, eksPRDescription, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				var eksPRUrl string
				By("Then I should see pull request created for eks to management cluster", func() {

					re := regexp.MustCompile(`Created pull request:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
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

				By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", capi_endpoint_url), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`%s get clusters --endpoint %s %s`, gitops_bin_path, capi_endpoint_url, insecureFlag))
				})

				By("Then I should see cluster status as 'Creation PR'", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+STATUS`))

					re := regexp.MustCompile(fmt.Sprintf(`%s\s+Creation PR`, eksClusterName))
					Eventually(re.Find([]byte(stdOut))).ShouldNot(BeNil())
					re = regexp.MustCompile(fmt.Sprintf(`%s\s+Creation PR`, capdClusterName))
					Eventually(re.Find([]byte(stdOut))).ShouldNot(BeNil())
				})
			})

			It("@git Verify giops can not create pull request to management cluster using existing branch", func() {
				By("When I create a private repository for cluster configs", func() {
					initAndCreateEmptyRepo(gitProviderEnv, true)
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

				cmd := fmt.Sprintf(`%s add cluster --from-template cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --branch %s --title %s --url %s --commit-message %s --description %s --endpoint %s %s`,
					gitops_bin_path, clusterName, namespace, branchName, prTitle, git_repository_url, prCommit, prDescription, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should not see pull request to be created", func() {
					Eventually(stdErr).Should(MatchRegexp(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})

		Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {

			JustAfterEach(func() {
				gitopsTestRunner.DeleteIPCredentials("AWS")
				gitopsTestRunner.DeleteIPCredentials("AZURE")
			})

			It("@git Verify gitops can use the matching selected credential for cluster creation", func() {
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

				cmd := fmt.Sprintf(` %s add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials aws-test-identity --dry-run --endpoint %s %s`,
					gitops_bin_path, awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should see preview containing identity reference added in the template", func() {
					// Verifying cluster object of the template for added credential reference
					re := regexp.MustCompile(fmt.Sprintf(`kind: AWSCluster\s+metadata:[\s\w\d-.:/]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterStaticIdentity\s+name: aws-test-identity`, awsClusterName))
					Eventually((re.Find([]byte(stdOut)))).ShouldNot(BeNil(), "Failed to find identity reference in preview pull request AWSCluster object")
				})
			})

			It("@git Verify gitops restrict user from using wrong credentials for infrastructure provider", func() {
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

				cmd := fmt.Sprintf(`%s add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials azure-cluster-identity --dry-run --endpoint %s %s`,
					gitops_bin_path, awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("And I run '%s'", cmd), func() {
					stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
				})

				// FIXME - User should get some warning or error as well for chossing wrong credential/identity for the infrastructure provider
				By("Then I should see preview without identity reference added to the template", func() {
					Eventually(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: [\s\w\d-.:/]+%s\s+---`, awsSshKeyName)), "Identity reference should not be found in preview pull request AWSCluster object")
				})
			})
		})

		Context("[CLI] When leaf cluster pull request is available in the management cluster", func() {
			kubeconfigPath := path.Join(os.Getenv("HOME"), "capi.kubeconfig")
			kustomizationFile := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "test_kustomization.yaml")
			appName := "management"
			appPath := "./management"
			capdClusterNames := []string{"cli-end-to-end-capd-cluster-1", "cli-end-to-end-capd-cluster-2"}

			JustBeforeEach(func() {
				_ = deleteFile([]string{kubeconfigPath})

				initializeWebdriver(test_ui_url)
				logger.Info("Connecting cluster to itself")
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

				logger.Info("Deleting all the wkp agents")
				_ = gitopsTestRunner.KubectlDeleteAllAgents([]string{})
				gitopsTestRunner.ResetControllers("enterprise")
				gitopsTestRunner.VerifyWegoPodsRunning()
			})

			It("@git @capd Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {

				By("Check wge is all running", func() {
					gitopsTestRunner.VerifyWegoPodsRunning()
				})

				By("When I create a private repository for cluster configs", func() {
					initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				By("And I install gitops to my active cluster", func() {
					Expect(fileExists(gitops_bin_path)).To(BeTrue(), fmt.Sprintf("%s can not be found.", gitops_bin_path))
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
				})

				By("Wait for cluster-service to cache profiles", func() {
					Expect(waitForProfiles(context.Background(), ASSERTION_30SECONDS_TIME_OUT)).To(Succeed())
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

					cmd := fmt.Sprintf(`%s add cluster --from-template cluster-template-development-observability-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=1 --set WORKER_MACHINE_COUNT=1 `, gitops_bin_path, clusterName, namespace, k8version) +
						fmt.Sprintf(`--profile 'name=podinfo,version=6.0.1' --branch "%s" --title "%s" --url %s --commit-message "%s" --description "%s" --endpoint %s %s`, prBranch, prTitle, git_repository_url, prCommit, prDescription, capi_endpoint_url, insecureFlag)
					By(fmt.Sprintf("And I run '%s'", cmd), func() {
						stdOut, stdErr = runCommandAndReturnStringOutput(cmd)
					})

					By("Then I should see pull request created to management cluster", func() {
						Expect(stdOut).Should(MatchRegexp(`name=podinfo[\s\w\d./:-]*version=6.0.1`))
						Expect(stdOut).Should(MatchRegexp(`Created pull request:\s*(?P<url>https:.*\/\d+)`))
					})

					By(fmt.Sprintf("Then I run 'gitops get clusters --endpoint %s'", capi_endpoint_url), func() {
						stdOut, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`%s get clusters --endpoint %s %s`, gitops_bin_path, capi_endpoint_url, insecureFlag))
					})

					By("And I should see cluster status as 'Creation PR'", func() {
						Expect(stdOut).Should(MatchRegexp(`NAME\s+STATUS`))
						Expect(stdOut).Should(MatchRegexp(fmt.Sprintf(`%s\s+Creation PR`, clusterName)))
					})

					By("Then I should merge the pull request to start cluster provisioning", func() {
						createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
						mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					})

					By("And I add a test kustomization file to the management appliction (because flux doesn't reconcile empty folders on deletion)", func() {
						pullGitRepo(repoAbsolutePath)
						_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("cp -f %s %s", kustomizationFile, path.Join(repoAbsolutePath, appPath)))
						gitUpdateCommitPush(repoAbsolutePath, "")
					})

					By("And I run gitops add app 'management' command", func() {
						if listGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE) == "" {
							addCommand := fmt.Sprintf("add app . --path=%s  --name=%s  --auto-merge=true", appPath, appName)
							runWegoAddCommand(repoAbsolutePath, addCommand, GITOPS_DEFAULT_NAMESPACE)
						} else {
							logger.Infof("Application '%s' alreaded exists", appName)
						}
					})

					By("And I should see cluster status changes to 'clusterFound'", func() {
						verifyWegoAddCommand(appName, GITOPS_DEFAULT_NAMESPACE)
						clusterFound := func() string {
							output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get clusters --endpoint %s %s`, gitops_bin_path, capi_endpoint_url, insecureFlag))
							return output
						}
						Eventually(clusterFound, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName)))
					})

					cmd = fmt.Sprintf(`%s get cluster %s --kubeconfig --endpoint %s %s | tee %s`, gitops_bin_path, clusterName, capi_endpoint_url, insecureFlag, kubeconfigPath)
					By(fmt.Sprintf("Then I run '%s'", cmd), func() {
						kubeConfigFound := func() string {
							output, _ := runCommandAndReturnStringOutput(cmd)
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
					verifyCapiClusterHealth(kubeconfigPath, clusterName, []string{}, GITOPS_DEFAULT_NAMESPACE)
				})

				clusterName2 := capdClusterNames[1]
				createCluster(clusterName2, namespace, k8version)

				// Deleting first cluster
				prBranch := fmt.Sprintf("%s-delete", clusterName)
				prTitle := "CAPD delete pull request"
				prCommit := "CAPD capi template deletion"
				prDescription := "This PR deletes CAPD Kubernetes cluster"

				cmd := fmt.Sprintf(`%s delete cluster %s --branch %s --title "%s" --url %s --commit-message "%s" --description "%s" --endpoint %s %s`,
					gitops_bin_path, clusterName, prBranch, prTitle, git_repository_url, prCommit, prDescription, capi_endpoint_url, insecureFlag)
				By(fmt.Sprintf("Then I run '%s'", cmd), func() {
					stdOut, _ = runCommandAndReturnStringOutput(cmd)
				})

				By("Then I should see delete pull request created to management cluster", func() {
					re := regexp.MustCompile(`Created pull request for clusters deletion:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request for deleting cluster")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					clusterDelete := func() string {
						output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get cluster %s --endpoint %s %s`, gitops_bin_path, clusterName, capi_endpoint_url, insecureFlag))
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
						output, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`%s get cluster %s --endpoint %s %s`, gitops_bin_path, clusterName2, capi_endpoint_url, insecureFlag))
						return output

					}
					Eventually(clusterFound).Should(MatchRegexp(fmt.Sprintf(`%s\s+clusterFound`, clusterName2)))
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster deleted", clusterName), func() {
					clusterFound := func() error {
						return runCommandPassThrough("kubectl", "get", "cluster", clusterName)
					}
					Eventually(clusterFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(HaveOccurred())
				})
			})
		})
	})
}
