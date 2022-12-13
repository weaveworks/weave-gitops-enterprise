package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func createProfileValuesYaml(profileName string, clusterName string) string {
	profileValues := fmt.Sprintf(`/tmp/%s-values.yaml`, profileName)

	var values map[string]interface{}

	switch profileName {
	case "cert-manager":
		values = map[string]interface{}{
			"installCRDs": "true",
		}
	case "weave-policy-agent":
		values = map[string]interface{}{
			"useCertManager": "true",
			"certificate":    "",
			"key":            "",
			"caCertificate":  "",
			"persistence":    map[string]string{"enabled": "false"},
			"audit":          map[string]string{"enabled": "false"},
			"policySource":   map[string]string{"enabled": "false"},
			"admission":      map[string]interface{}{"enabled": "true", "sinks": map[string]interface{}{"k8sEventsSink": map[string]string{"enabled": "true"}}},
			"config":         map[string]string{"accountId": "weaveworks", "clusterId": clusterName},
		}
	}

	data, err := yaml.Marshal(&values)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to serializes yaml values")

	err = ioutil.WriteFile(profileValues, data, 0644)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to write data to file "+profileValues)

	return profileValues
}

var _ = ginkgo.Describe("Gitops GitOpsTemplate tests for CAPI cluster", ginkgo.Label("cli", "template"), func() {
	var stdOut string
	var repoAbsolutePath string
	clusterPath := "./clusters/management/clusters"

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
	})

	ginkgo.Context("[CLI] When no infrastructure provider credentials are available in the management cluster", func() {
		ginkgo.It("Verify gitops lists no credentials", func() {
			stdOut, _ = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

			ginkgo.By("Then gitops lists no credentials", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("No credentials were found"))
			})
		})
	})

	ginkgo.Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {

		ginkgo.JustAfterEach(func() {
			deleteIPCredentials("AWS")
			deleteIPCredentials("AZURE")
		})

		ginkgo.It("Verify gitops can list credentials present in the management cluster", func() {
			ginkgo.By("And create AWS credentials)", func() {
				createIPCredentials("AWS")
			})

			stdOut, _ = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

			ginkgo.By("Then gitops lists AWS credentials", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`aws-test-identity`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`test-role-identity`))
			})

			ginkgo.By("And create AZURE credentials)", func() {
				createIPCredentials("AZURE")
			})

			stdOut, _ = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

			ginkgo.By("Then gitops lists AZURE credentials", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`azure-cluster-identity`))
			})
		})

		ginkgo.It("Verify gitops can use the matching selected credential for cluster creation", func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-ec2.yaml"),
				"capz-cluster-template": path.Join(testDataPath, "templates/cluster/azure/cluster-template-e2e.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			ginkgo.By("And create AWS credentials)", func() {
				createIPCredentials("AWS")
			})

			ginkgo.By("And create AZURE credentials)", func() {
				createIPCredentials("AZURE")
			})

			// AWS Parameter values
			awsClusterName := "my-aws-cluster"
			awsRegion := "eu-west-3"
			awsK8version := "1.19.8"
			awsSshKeyName := "my-aws-ssh-key"
			awsNamespace := "default"
			awsControlMAchineType := "t4g.large"
			awsNodeMAchineType := "t3.micro"

			cmd := fmt.Sprintf(`add cluster --from-template capa-cluster-template --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials aws-test-identity --dry-run`,
				awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType)
			stdOut, _ = runGitopsCommand(cmd)

			ginkgo.By("Then I should see preview containing identity reference added in the template", func() {
				// Verifying cluster object of the template for added credential reference
				re := regexp.MustCompile(fmt.Sprintf(`kind: AWSCluster\s+metadata:[\s\w\d-.:/]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterStaticIdentity\s+name: aws-test-identity`, awsClusterName))
				gomega.Eventually((re.Find([]byte(stdOut)))).ShouldNot(gomega.BeNil(), "Failed to find identity reference in preview pull request AWSCluster object")
			})
		})

		ginkgo.It("Verify gitops restrict user from using wrong credentials for infrastructure provider", func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-ec2.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			ginkgo.By("And create AZURE credentials)", func() {
				createIPCredentials("AZURE")
			})

			// AWS Parameter values
			awsClusterName := "my-aws-cluster"
			awsRegion := "eu-west-3"
			awsK8version := "1.19.8"
			awsSshKeyName := "my-aws-ssh-key"
			awsNamespace := "default"
			awsControlMAchineType := "t4g.large"
			awsNodeMAchineType := "t3.micro"

			cmd := fmt.Sprintf(`add cluster --from-template capa-cluster-template --set CLUSTER_NAME=%s --set AWS_REGION=%s --set KUBERNETES_VERSION=%s --set AWS_SSH_KEY_NAME=%s --set NAMESPACE=%s --set CONTROL_PLANE_MACHINE_COUNT=2 --set AWS_CONTROL_PLANE_MACHINE_TYPE=%s --set WORKER_MACHINE_COUNT=3 --set AWS_NODE_MACHINE_TYPE=%s --set-credentials azure-cluster-identity --dry-run`,
				awsClusterName, awsRegion, awsK8version, awsSshKeyName, awsNamespace, awsControlMAchineType, awsNodeMAchineType)
			stdOut, _ = runGitopsCommand(cmd)

			// FIXME - User should get some warning or error as well for chossing wrong credential/identity for the infrastructure provider
			ginkgo.By("Then I should see preview without identity reference added to the template", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: [\s\w\d-.:/]+%s\s+---`, awsSshKeyName)), "Identity reference should not be found in preview pull request AWSCluster object")
			})
		})
	})

	ginkgo.Context("[CLI] When leaf cluster pull request is available in the management cluster", func() {
		var clusterBootstrapCopnfig string
		var clusterResourceSet string
		var crsConfigmap string
		var capdClusters []ClusterConfig
		var kubeconfigPath string

		clusterNamespace := map[string]string{
			// GitProviderGitLab: "capi-test-system",
			GitProviderGitLab: "default",
			GitProviderGitHub: "default",
		}

		bootstrapLabel := "bootstrap"
		patSecret := "capi-pat"

		ginkgo.JustBeforeEach(func() {
			kubeconfigPath = path.Join(os.Getenv("HOME"), "capi.kubeconfig")
			capdClusters = []ClusterConfig{
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

		ginkgo.JustAfterEach(func() {
			_ = deleteFile([]string{kubeconfigPath})
			deleteSecret([]string{patSecret}, capdClusters[0].Namespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", crsConfigmap)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterResourceSet)

			reconcile("suspend", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			// Force clean the repository directory for subsequent tests
			cleanGitRepository(clusterPath)
			// Force delete capicluster incase delete PR fails to delete to free resources
			removeGitopsCapiClusters(capdClusters)
			reconcile("resume", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
		})

		ginkgo.It("Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", ginkgo.Label("capd"), func() {
			ginkgo.By("And wait for cluster-service to cache profiles", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
			})

			templateFiles := map[string]string{
				"capd-cluster-template": path.Join(testDataPath, "templates/cluster/docker/cluster-template.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			createCluster := func(clusterName string, namespace string, k8version string, profiles []Application) {
				//Pull request values
				prBranch := fmt.Sprintf("br-%s", clusterName)
				prTitle := "CAPD pull request"
				prCommit := "CAPD capi template"
				prDescription := "This PR creates a new CAPD Kubernetes cluster"

				profileFlag := ""
				if len(profiles) > 0 {
					for _, p := range profiles {
						profileFlag += fmt.Sprintf(`--profile 'name=%s,version=%s,namespace=%s,values=%s' `, p.Name, p.Version, p.TargetNamespace, p.Values)
					}
				}

				cmd := fmt.Sprintf(`add cluster --from-template capd-cluster-template --set CLUSTER_NAME=%s --set NAMESPACE=%s --set KUBERNETES_VERSION=%s --set CONTROL_PLANE_MACHINE_COUNT=1 --set WORKER_MACHINE_COUNT=1 --set INSTALL_CRDS=true`, clusterName, namespace, k8version) +
					fmt.Sprintf(`%s --branch "%s" --title "%s" --url %s --commit-message "%s" --description "%s"`,
						profileFlag, prBranch, prTitle, gitRepositoryUrl, prCommit, prDescription)
				stdOut, _ = runGitopsCommand(cmd, ASSERTION_30SECONDS_TIME_OUT)

				ginkgo.By("Then I should see pull request created to management cluster", func() {
					gomega.Expect(stdOut).Should(gomega.MatchRegexp(`Created pull request:\s*(?P<url>https:.*\/\d+)`))
				})

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("And I should see cluster status changes to 'Ready'", func() {
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					clusterFound := func() string {
						output, _ := runGitopsCommand(`get clusters`)
						return output
					}
					gomega.Eventually(clusterFound, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(fmt.Sprintf(`%s\s+Ready`, clusterName)), clusterName+" cluster status should be Ready")
				})

				cmd = fmt.Sprintf(`get cluster %s --print-kubeconfig | tee %s`, clusterName, kubeconfigPath)
				ginkgo.By("And I should print/download the kubeconfig for the CAPD capi cluster "+clusterName, func() {
					kubeConfigFound := func() string {
						output, _ := runGitopsCommand(cmd)
						return output

					}
					gomega.Eventually(kubeConfigFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)), "Failed to download kubeconfig for cluster "+clusterName)
				})
			}

			// Parameter values
			clusterName := capdClusters[0].Name
			namespace := capdClusters[0].Namespace
			k8version := "1.23.3"
			profiles := []Application{
				{
					Name:            "cert-manager",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "cert-manager",
					Version:         "0.0.8",
					Values:          createProfileValuesYaml("cert-manager", clusterName),
				},
				{
					Name:            "weave-policy-agent",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "policy-system",
					Version:         "0.5.0",
					Values:          createProfileValuesYaml("weave-policy-agent", clusterName),
				},
			}

			// Creating two capd clusters
			createCluster(clusterName, namespace, k8version, profiles)

			ginkgo.By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", clusterName), func() {
				verifyCapiClusterKubeconfig(kubeconfigPath, clusterName)
			})

			ginkgo.By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", clusterName), func() {
				// verifyCapiClusterHealth(kubeconfigPath, profiles)
				verifyCapiClusterHealth(kubeconfigPath, []Application{})
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
				clusterName, prBranch, prTitle, gitRepositoryUrl, prCommit, prDescription)
			stdOut, _ = runGitopsCommand(cmd)

			ginkgo.By("Then I should see delete pull request created to management cluster", func() {
				re := regexp.MustCompile(`Created pull request for clusters deletion:\s*(?P<url>https:.*\/\d+)`)
				match := re.FindSubmatch([]byte(stdOut))
				gomega.Eventually(match).ShouldNot(gomega.BeNil(), "Failed to Create pull request for deleting cluster")
			})

			var deletePRUrl string
			ginkgo.By("And I should veriyfy the delete pull request in the cluster config repository", func() {
				deletePRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
			})

			ginkgo.By("Then I should merge the delete pull request to delete cluster", func() {
				mergePullRequest(gitProviderEnv, repoAbsolutePath, deletePRUrl)
			})

			ginkgo.By("And the delete pull request manifests are not present in the cluster config repository", func() {
				pullGitRepo(repoAbsolutePath)
				_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, namespace, clusterName+".yaml"))
				gomega.Expect(err).Should(gomega.MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
			})

			ginkgo.By(fmt.Sprintf("And I should see the '%s' cluster (not deleted) status remains unchanged as 'Ready'", clusterName2), func() {
				clusterStatus := func() string {
					output, _ := runGitopsCommand(`get cluster ` + clusterName2)
					return output

				}
				gomega.Eventually(clusterStatus).Should(gomega.MatchRegexp(fmt.Sprintf(`%s\s+Ready`, clusterName2)), clusterName2+" cluster status should be Ready")
			})

			ginkgo.By(fmt.Sprintf("Then I should see the '%s' cluster deleted/disappeared", clusterName), func() {
				clusterFound := func() error {
					return runCommandPassThrough("kubectl", "get", "cluster", clusterName)
				}
				gomega.Eventually(clusterFound, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.HaveOccurred(), clusterName+" cluster should be deleted")
			})
		})
	})
})
