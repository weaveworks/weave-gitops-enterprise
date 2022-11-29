package acceptance

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func DescribeCliGet(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Gitops Get Tests", func() {

		templateFiles := []string{}
		var stdOut string
		var stdErr string

		ginkgo.BeforeEach(func() {

		})

		ginkgo.AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		ginkgo.Context("[CLI] When no Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify gitops lists no templates", func() {
				stdOut, stdErr = runGitopsCommand(`get templates`)

				ginkgo.By("Then gitops lists no templates", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp("No templates were found"))
				})
			})
		})

		ginkgo.Context("[CLI] When only invalid Capi Template(s) are available in the cluster", func() {
			ginkgo.It("Verify gitops outputs an error message related to an invalid template(s)", func() {

				noOfTemplates := 1
				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "templates/miscellaneous/invalid-cluster-template.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				ginkgo.By("Then I should see template table header", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				ginkgo.By("And I should see template rows", func() {
					re := regexp.MustCompile(`invalid-cluster-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(stdOut, 1)
					gomega.Eventually(len(matched_list)).Should(gomega.Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")
				})
			})
		})

		ginkgo.Context("[CLI] When both valid and invalid Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify gitops outputs an error message related to an invalid template and lists the valid template", func() {

				noOfTemplates := 3
				ginkgo.By("Apply/Install valid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "templates/cluster/aws/cluster-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 2
				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					invalid_captemplate := gitopsTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "templates/miscellaneous/invalid-cluster-template.yaml")
					templateFiles = append(templateFiles, invalid_captemplate...)
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				ginkgo.By("Then I should see template table header", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				ginkgo.By("And I should see template rows for invalid template", func() {
					re := regexp.MustCompile(`invalid-cluster-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(stdOut, noOfInvalidTemplates)
					gomega.Eventually(len(matched_list)).Should(gomega.Equal(noOfInvalidTemplates), "The number of listed invalid templates should be equal to number of templates created")
				})

				ginkgo.By("And I should see template rows for valid template", func() {
					re := regexp.MustCompile(`capa-cluster-template-eks-fargate-[\d]+\s+aws\s+This is eks fargate template-[\d]+`)
					matched_list := re.FindAllString(stdOut, noOfTemplates)
					gomega.Eventually(len(matched_list)).Should(gomega.Equal(noOfTemplates), "The number of listed valid templates should be equal to number of templates created")
				})
			})

			ginkgo.It("Verify gitops reports an error when listing template parameters of invalid template from template library", func() {

				noOfTemplates := 1
				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "templates/miscellaneous/invalid-cluster-template.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates invalid-cluster-template-0 --list-parameters`)

				ginkgo.By("Then I should not see template parameter table header", func() {
					gomega.Eventually(stdOut).ShouldNot(gomega.MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				ginkgo.By("And I should see error message related to invalid template", func() {
					gomega.Eventually(stdErr).Should(gomega.MatchRegexp(`Error: unable to retrieve parameters.+`))
				})
			})
		})

		ginkgo.Context("[CLI] When Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify gitops can list templates from template library", func() {

				noOfTemplates := 5
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "templates/cluster/azure/cluster-template-e2e.yaml")

				stdOut, stdErr = runGitopsCommand(`get templates`)

				ginkgo.By("Then I should see template table header", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				ginkgo.By("And I should see template rows", func() {
					re := regexp.MustCompile(`capz-cluster-template-[\d]+\s+azure\s+This is Azure capi quick start template-[\d]+`)
					matched_list := re.FindAllString(stdOut, noOfTemplates)
					gomega.Eventually(len(matched_list)).Should(gomega.Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")

					// Testing templates are ordered
					expected_list := make([]string, noOfTemplates)
					for i := 0; i < noOfTemplates; i++ {
						expected_list[i] = fmt.Sprintf("capz-cluster-template-%d", i)
					}
					sort.Strings(expected_list)

					for i := 0; i < noOfTemplates; i++ {
						gomega.Expect(matched_list[i]).Should(gomega.ContainSubstring(expected_list[i]))
					}
				})

				stdOut, stdErr = runGitopsCommand("get templates --namespace foo")

				// ginkgo.By("Then I should see an error message", func() {
				//		gomega.Eventually(stdErr).Should(gomega.MatchRegexp(`No templates were found`))
				// })
			})

			ginkgo.It("Verify gitops can list filtered templates from template library", func() {
				awsTemplateCount := 2
				eksFargateTemplateCount := 2
				capdTemplateCount := 5
				totalTemplateCount := awsTemplateCount + eksFargateTemplateCount + capdTemplateCount
				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "templates/cluster/docker/cluster-template.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "templates/cluster/aws/cluster-template-ec2.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "templates/cluster/aws/cluster-template-eks-fargate.yaml")...)
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				ginkgo.By("And I should see template list table header", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				ginkgo.By("And I should see template rows", func() {
					re := regexp.MustCompile(`capa-cluster-template-eks-fargate-[\d]+\s+aws\s+This is eks fargate template-[\d]+.+`)
					matched_list := re.FindAllString(stdOut, eksFargateTemplateCount)
					gomega.Eventually(len(matched_list)).Should(gomega.Equal(eksFargateTemplateCount), "The number of listed templates should be equal to number of templates created")
				})

				ginkgo.By("And I should see ordered list of templates", func() {
					expected_list := make([]string, totalTemplateCount)
					for i := 0; i < awsTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capa-cluster-template-%d", i)
					}
					for i := 0; i < eksFargateTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capa-cluster-template-eks-fargate-%d", i)
					}
					for i := 0; i < capdTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capd-cluster-template-%d", i)
					}
					sort.Strings(expected_list)

					for i := 0; i < totalTemplateCount; i++ {
						gomega.Eventually(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`%s\s+.*`, expected_list[i])))
					}

				})

				stdOut, stdErr = runGitopsCommand(`get templates --provider aws`)

				ginkgo.By("And I should see templates list filtered by provider", func() {
					awsCluster_list := make([]string, awsTemplateCount+eksFargateTemplateCount)
					for i := 0; i < awsTemplateCount; i++ {
						awsCluster_list[i] = fmt.Sprintf("capa-cluster-template-%d", i)
					}
					for i := 0; i < eksFargateTemplateCount; i++ {
						awsCluster_list[i] = fmt.Sprintf("capa-cluster-template-eks-fargate-%d", i)
					}
					sort.Strings(awsCluster_list)
					for i := 0; i < awsTemplateCount+eksFargateTemplateCount; i++ {
						gomega.Eventually(stdOut).Should(gomega.MatchRegexp(fmt.Sprintf(`%s\s+.*`, awsCluster_list[i])))
					}

					for i := 0; i < 5; i++ {
						capd_template := fmt.Sprintf("capd-cluster-template-%d", i)
						re := regexp.MustCompile(fmt.Sprintf(`%s\s+.*`, capd_template))
						gomega.Eventually((re.Find([]byte(stdOut)))).Should(gomega.BeNil())
					}

				})

				_, stdErr = runGitopsCommand(`get templates --provider foobar`)

				ginkgo.By("And I should see error message for invalid provider", func() {
					gomega.Eventually(stdErr).Should(gomega.MatchRegexp(`Error:\s+provider "foobar" is not valid.*`))
				})
			})

			ginkgo.It("Verify gitops can list template parameters of a template from template library", func() {

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/aws/cluster-template-eks-fargate.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates capa-cluster-template-eks-fargate-0 --list-parameters`)

				ginkgo.By("Then I should see template parameter table header", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				ginkgo.By("And I should see parameter rows", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`CLUSTER_NAME+\s+true\s+This is used for the cluster naming`))
					gomega.Expect(stdOut).Should(gomega.MatchRegexp(`AWS_REGION+\s+true\s+AWS Region to create cluster`))
					gomega.Expect(stdOut).Should(gomega.MatchRegexp(`AWS_SSH_KEY_NAME+\s+false\s+AWS ssh key name`))
					gomega.Expect(stdOut).Should(gomega.MatchRegexp(`KUBERNETES_VERSION+\s+false`))
				})
			})
		})

		ginkgo.Context("[CLI] When no infrastructure provider credentials are available in the management cluster", func() {
			ginkgo.It("Verify gitops lists no credentials", func() {
				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				ginkgo.By("Then gitops lists no credentials", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp("No credentials were found"))
				})
			})
		})

		ginkgo.Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {
			ginkgo.It("Verify gitops can list credentials present in the management cluster", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AWS")
				defer gitopsTestRunner.DeleteIPCredentials("AZURE")

				ginkgo.By("And create AWS credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
				})

				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				ginkgo.By("Then gitops lists AWS credentials", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`aws-test-identity`))
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`test-role-identity`))
				})

				ginkgo.By("And create AZURE credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				ginkgo.By("Then gitops lists AZURE credentials", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`azure-cluster-identity`))
				})
			})
		})

		ginkgo.Context("[CLI] When no clusters are available in the management cluster", func() {
			ginkgo.It("Verify gitops lists no clusters", func() {
				ginkgo.By("And gitops state is reset", func() {
					gitopsTestRunner.ResetControllers("enterprise")
					gitopsTestRunner.VerifyWegoPodsRunning()
				})

				stdOut, stdErr = runGitopsCommand(`get cluster`)

				ginkgo.By("Then gitops lists no clusters", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`management\s+Ready`))
				})
			})
		})

		ginkgo.Context("[CLI] When profiles are available in the management cluster", func() {
			ginkgo.It("Verify gitops can list profiles from profile repository", func() {
				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
				})

				stdOut, stdErr = runGitopsCommand(`get profiles`)

				ginkgo.By("Then gitops lists profiles", func() {
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`cert-manager\s+A Weaveworks Helm chart for the Certificate Profile[,.\d\w\s]+0.0.8,0.0.7`))
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`weave-policy-agent\s+A Weaveworks Helm chart for Kubernetes to configure the policy agent[,.\d\w\s]+0.4.0`))
					gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`metallb\s+A Weaveworks Helm chart for a network load-balancer implementation[,.\d\w\s]+0.0.2,0.0.1`))
				})
			})
		})

		ginkgo.Context("[CLI] When entitlement is available in the cluster", func() {
			var resourceName string
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					cmd := fmt.Sprintf(`get %s`, resourceName)
					stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_1MINUTE_TIME_OUT)

					msg := stdErr + " " + stdOut

					if typeEntitelment == "expired" {
						re := regexp.MustCompile(`Your entitlement for Weave GitOps Enterprise has expired`)
						return re.MatchString(msg)
					}
					re := regexp.MustCompile(`No entitlement was found for Weave GitOps Enterprise`)
					return re.MatchString(msg)

				}

				matcher := gomega.BeFalse
				if beFound {
					matcher = gomega.BeTrue
				}

				resourceName = "templates"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "credentials"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "clusters"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
			}

			ginkgo.JustBeforeEach(func() {

				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				gitopsTestRunner.CreateIPCredentials("AWS")
			})

			ginkgo.JustAfterEach(func() {
				ginkgo.By("When I apply the valid entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				ginkgo.By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("invalid", false)
				})

				gitopsTestRunner.DeleteIPCredentials("AWS")
				// Login to the dashbord because the logout automatically when the cluster service restarts for entitlement checking
				loginUser()
			})

			ginkgo.It("Verify cluster service acknowledges the entitlement presences", func() {

				ginkgo.By("When I delete the entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlDelete([]string{}, "../../utils/data/entitlement/entitlement-secret.yaml"), "Failed to delete entitlement secret")
				})

				ginkgo.By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)
				})

				ginkgo.By("When I apply the expired entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret-expired.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				ginkgo.By("When I apply the invalid entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret-invalid.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})

	})
}
