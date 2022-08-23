package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

type CapiClusterConfig struct {
	Type      string
	Name      string
	Namespace string
}

type Profile struct {
	Name      string
	Namespace string
	Version   string
	Values    string
	Layer     string
}

func createProfileValuesYaml(clusterName string) string {
	profileValues := `/tmp/profile-values.yaml`

	values := map[string]string{
		"installCRDs": "true",                                 // cert-manager values
		"accountId":   "weaveworks", "clusterId": clusterName} // policy-agent values

	data, err := yaml.Marshal(&values)
	Expect(err).ShouldNot(HaveOccurred(), "Failed to serializes yaml values")

	err = ioutil.WriteFile(profileValues, data, 0644)
	Expect(err).ShouldNot(HaveOccurred(), "Failed to write data to file "+profileValues)

	return profileValues
}

func DescribeCliAddDelete(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops add Tests", func() {
		var stdOut string
		var stdErr string
		var repoAbsolutePath string
		templateFiles := []string{}
		clusterPath := "./clusters/my-cluster/clusters"

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {

			It("Verify gitops can set template parameters by specifying multiple parameters --set key=value --set key=value", Label("git"), func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.19.7"
				controlPlaneMachineCount := "2"
				workerMachineCount := "3"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --from-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=%s --set WORKER_MACHINE_COUNT=%s`,
					clusterName, namespace, k8version, controlPlaneMachineCount, workerMachineCount)
				stdOut, stdErr = runGitopsCommand(cmd)

				By("Then I should see template preview with updated parameter values", func() {
					// Verifying cluster object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: Cluster\s+metadata:\s+labels:\s+cni: calico[\s\w\d-.:/]+name: %[1]v\s+namespace: %[2]v[\s\w\d-.:/]+kind: KubeadmControlPlane\s+name: %[1]v-control-plane\s+namespace: %[2]v`,
						clusterName, namespace))

					// Verifying KubeadmControlPlane object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: KubeadmControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: %[2]v[\s\w\d"<%%,/:.-]+replicas: %[3]v[\s\w\d/:.-]+version: %[4]v`,
						clusterName, namespace, controlPlaneMachineCount, k8version))

					// Verifying MachineDeployment object of tbe template for updated  parameter values
					Eventually(stdOut).Should(MatchRegexp(`kind: MachineDeployment\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-md-0\s+namespace: %[2]v\s+spec:\s+clusterName: %[1]v[\s\w\d/:.-]+replicas: %[3]v[\s\w\d/:.-]+infrastructureRef:[\s\w\d/:.-]+version: %[4]v`,
						clusterName, namespace, workerMachineCount, k8version))
				})
			})

			It("Verify gitops can set template parameters by separate values with commas key1=val1,key2=val2", Label("git"), func() {
				clusterName := "development-cluster"
				namespace := "gitops-dev"
				k8version := "1.23.6"
				controlPlaneMachineCount := "1"
				workerMachineCount := "2"

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				cmd := fmt.Sprintf(`add cluster --from-template cluster-template-development-0 --dry-run --set CLUSTER_NAME=%s,NAMESPACE=%s,KUBERNETES_VERSION=%s,CONTROL_PLANE_MACHINE_COUNT=%s,WORKER_MACHINE_COUNT=%s`,
					clusterName, namespace, k8version, controlPlaneMachineCount, workerMachineCount)
				stdOut, stdErr = runGitopsCommand(cmd)

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

			It("Verify gitops reports an error when trying to create pull request with missing --from-template argument", Label("git"), func() {
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

				cmd := fmt.Sprintf(`add cluster --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --branch %s --url %s --commit-message %s --description %s`, clusterName, namespace, k8version, prBranch, git_repository_url, prCommit, prDescription)
				stdOut, stdErr = runGitopsCommand(cmd)

				By("Then I should see an error for required argument to create pull request", func() {
					Eventually(stdErr).Should(MatchRegexp(`Error: unable to create pull request.*template name must be specified`))
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster to create pull requests", func() {

			BeforeEach(func() {
				repoAbsolutePath = configRepoAbsolutePath(gitProviderEnv)
			})

			JustAfterEach(func() {
				cleanGitRepository(clusterPath)
			})

			It("Verify gitops can create pull requests to management cluster", Label("smoke", "git"), func() {
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

				By("Apply/Install CAPITemplate", func() {
					capdTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
					templateFiles = append(capdTemplateFile, eksTemplateFile...)
				})

				cmd := fmt.Sprintf(`add cluster --from-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=%s --set WORKER_MACHINE_COUNT=%s --branch %s --title %s --url %s --commit-message %s --description %s`,
					capdClusterName, capdNamespace, capdK8version, controlPlaneMachineCount, workerMachineCount, capdPRBranch, capdPRTitle, git_repository_url, capdPRCommit, capdPRDescription)
				stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

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
					_, err := os.Stat(fmt.Sprintf("%s/clusters/my-cluster/clusters/%s/%s.yaml", repoAbsolutePath, capdNamespace, capdClusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
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

				cmd = fmt.Sprintf(`add cluster --from-template eks-fargate-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --branch %s --title %s --url %s --commit-message %s --description %s`,
					eksClusterName, eksRegion, eksK8version, eksSshKeyName, eksNamespace, eksPRBranch, eksPRTitle, git_repository_url, eksPRCommit, eksPRDescription)
				stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

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
					_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, eksNamespace, eksClusterName+".yaml"))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})

				stdOut, stdErr = runGitopsCommand(`get clusters`)

			})

			It("Verify giops can not create pull request to management cluster using existing branch", Label("git"), func() {
				branchName := "cli-test-branch"
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

				cmd := fmt.Sprintf(`add cluster --from-template cluster-template-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --branch %s --title %s --url %s --commit-message %s --description %s`,
					clusterName, namespace, branchName, prTitle, git_repository_url, prCommit, prDescription)
				stdOut, stdErr = runGitopsCommand(cmd)

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

			It("Verify gitops can use the matching selected credential for cluster creation", Label("git"), func() {
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

				cmd := fmt.Sprintf(`add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials aws-test-identity --dry-run`,
					awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType)
				stdOut, stdErr = runGitopsCommand(cmd)

				By("Then I should see preview containing identity reference added in the template", func() {
					// Verifying cluster object of the template for added credential reference
					re := regexp.MustCompile(fmt.Sprintf(`kind: AWSCluster\s+metadata:[\s\w\d-.:/]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterStaticIdentity\s+name: aws-test-identity`, awsClusterName))
					Eventually((re.Find([]byte(stdOut)))).ShouldNot(BeNil(), "Failed to find identity reference in preview pull request AWSCluster object")
				})
			})

			It("Verify gitops restrict user from using wrong credentials for infrastructure provider", Label("git"), func() {
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

				cmd := fmt.Sprintf(`add cluster --from-template aws-cluster-template-0 --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials azure-cluster-identity --dry-run`,
					awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType)
				stdOut, stdErr = runGitopsCommand(cmd)

				// FIXME - User should get some warning or error as well for chossing wrong credential/identity for the infrastructure provider
				By("Then I should see preview without identity reference added to the template", func() {
					Eventually(stdOut).Should(MatchRegexp(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: [\s\w\d-.:/]+%s\s+---`, awsSshKeyName)), "Identity reference should not be found in preview pull request AWSCluster object")
				})
			})
		})

		Context("[CLI] When leaf cluster pull request is available in the management cluster", func() {
			var clusterBootstrapCopnfig string
			var clusterResourceSet string
			var crsConfigmap string
			var capdClusters []CapiClusterConfig
			var kubeconfigPath string

			clusterNamespace := map[string]string{
				// GitProviderGitLab: "capi-test-system",
				GitProviderGitLab: "default",
				GitProviderGitHub: "default",
			}

			bootstrapLabel := "bootstrap"
			patSecret := "capi-pat"

			JustBeforeEach(func() {
				kubeconfigPath = path.Join(os.Getenv("HOME"), "capi.kubeconfig")
				capdClusters = []CapiClusterConfig{
					{"capd", "cli-end-to-end-capd-cluster-1", clusterNamespace[gitProviderEnv.Type]},
					{"capd", "cli-end-to-end-capd-cluster-2", "default"},
				}
				_ = deleteFile([]string{kubeconfigPath})

				repoAbsolutePath = configRepoAbsolutePath(gitProviderEnv)
				createNamespace([]string{capdClusters[0].Namespace})
				createPATSecret(capdClusters[0].Namespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(capdClusters[0].Name, capdClusters[0].Namespace, bootstrapLabel, patSecret)
				clusterResourceSet = createClusterResourceSet(capdClusters[0].Name, capdClusters[0].Namespace)
				crsConfigmap = createCRSConfigmap(capdClusters[0].Name, capdClusters[0].Namespace)
			})

			JustAfterEach(func() {
				_ = deleteFile([]string{kubeconfigPath})
				deleteSecret([]string{patSecret}, capdClusters[0].Namespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, crsConfigmap)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterResourceSet)

				// Force clean the repository directory for subsequent tests
				cleanGitRepository(clusterPath)
				// Force delete capicluster incase delete PR fails to delete to free resources
				removeGitopsCapiClusters(capdClusters)
			})

			It("Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", Label("capd", "git"), func() {
				By("And wait for cluster-service to cache profiles", func() {
					Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS)).To(Succeed(), "Failed to get a successful response from /v1/profiles ")
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				createCluster := func(clusterName string, namespace string, k8version string, profiles []Profile) {
					//Pull request values
					prBranch := fmt.Sprintf("br-%s", clusterName)
					prTitle := "CAPD pull request"
					prCommit := "CAPD capi template"
					prDescription := "This PR creates a new CAPD Kubernetes cluster"

					profileFlag := ""
					if len(profiles) > 0 {
						for _, p := range profiles {
							// profileFlag += fmt.Sprintf(`--profile 'name=%s,version=%s,values=%s'`, p.Name, p.Version, p.Values)
							profileFlag += fmt.Sprintf(`--profile 'name=%s,version=%s'`, p.Name, p.Version)
						}
					}

					cmd := fmt.Sprintf(`add cluster --from-template cluster-template-development-0 --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=1 --set WORKER_MACHINE_COUNT=1 `, clusterName, namespace, k8version) +
						fmt.Sprintf(`%s --branch "%s" --title "%s" --url %s --commit-message "%s" --description "%s"`,
							profileFlag, prBranch, prTitle, git_repository_url, prCommit, prDescription)
					stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

					By("Then I should see pull request created to management cluster", func() {
						Expect(stdOut).Should(MatchRegexp(`Created pull request:\s*(?P<url>https:.*\/\d+)`))
					})

					By("Then I should merge the pull request to start cluster provisioning", func() {
						createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
						mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					})

					By("And I should see cluster status changes to 'Ready'", func() {
						waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
						clusterFound := func() string {
							output, _ := runGitopsCommand(`get clusters`)
							return output
						}
						Eventually(clusterFound, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`%s\s+Ready`, clusterName)), clusterName+" cluster status should be Ready")
					})

					cmd = fmt.Sprintf(`get cluster %s --print-kubeconfig | tee %s`, clusterName, kubeconfigPath)
					By("And I should print/download the kubeconfig for the CAPD capi cluster "+clusterName, func() {
						kubeConfigFound := func() string {
							output, _ := runGitopsCommand(cmd)
							return output

						}
						Eventually(kubeConfigFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)), "Failed to download kubeconfig for cluster "+clusterName)
					})
				}

				// Parameter values
				clusterName := capdClusters[0].Name
				namespace := capdClusters[0].Namespace
				k8version := "1.23.3"
				profileValues := createProfileValuesYaml(clusterName)
				profiles := []Profile{
					{
						Name:      "cert-manager",
						Namespace: "cert-manager",
						Version:   "0.0.7",
						Values:    profileValues,
					},
					{
						Name:      "weave-policy-agent",
						Namespace: "policy-system",
						Version:   "0.3.1",
						Values:    profileValues,
					},
				}

				// Creating two capd clusters
				createCluster(clusterName, namespace, k8version, profiles)

				By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", clusterName), func() {
					verifyCapiClusterKubeconfig(kubeconfigPath, clusterName)
				})

				By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", clusterName), func() {
					// Fixme: profiles are not installed via cli; issue #1228
					// podInfo := Profile{
					// 	Name:      "podinfo",
					// 	Namespace: "flux-system",
					// }
					verifyCapiClusterHealth(kubeconfigPath, []Profile{})
				})

				clusterName2 := capdClusters[1].Name
				namespace2 := capdClusters[1].Namespace
				createCluster(clusterName2, namespace2, k8version, nil)

				// Deleting first cluster
				prBranch := fmt.Sprintf("%s-delete", clusterName)
				prTitle := "CAPD delete pull request"
				prCommit := "CAPD capi template deletion"
				prDescription := "This PR deletes CAPD Kubernetes cluster"

				cmd := fmt.Sprintf(`delete cluster %s --branch %s --title "%s" --url %s --commit-message "%s" --description "%s"`,
					clusterName, prBranch, prTitle, git_repository_url, prCommit, prDescription)
				stdOut, _ = runGitopsCommand(cmd)

				By("Then I should see delete pull request created to management cluster", func() {
					re := regexp.MustCompile(`Created pull request for clusters deletion:\s*(?P<url>https:.*\/\d+)`)
					match := re.FindSubmatch([]byte(stdOut))
					Eventually(match).ShouldNot(BeNil(), "Failed to Create pull request for deleting cluster")
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
					_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, namespace, clusterName+".yaml"))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("And I should see the '%s' cluster (not deleted) status remains unchanged as 'Ready'", clusterName2), func() {
					clusterStatus := func() string {
						output, _ := runGitopsCommand(`get cluster ` + clusterName2)
						return output

					}
					Eventually(clusterStatus).Should(MatchRegexp(fmt.Sprintf(`%s\s+Ready`, clusterName2)), clusterName2+" cluster status should be Ready")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster deleted/disappeared", clusterName), func() {
					clusterFound := func() error {
						return runCommandPassThrough("kubectl", "get", "cluster", clusterName)
					}
					Eventually(clusterFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(HaveOccurred(), clusterName+" cluster should be deleted")
				})
			})
		})
	})
}
