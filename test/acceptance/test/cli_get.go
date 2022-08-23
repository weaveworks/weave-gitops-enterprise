package acceptance

import (
	"fmt"
	"regexp"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func DescribeCliGet(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops Get Tests", func() {

		templateFiles := []string{}
		var stdOut string
		var stdErr string

		BeforeEach(func() {

		})

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		Context("[CLI] When no Capi Templates are available in the cluster", func() {
			It("Verify gitops lists no templates", func() {
				stdOut, stdErr = runGitopsCommand(`get templates`)

				By("Then gitops lists no templates", func() {
					Eventually(stdOut).Should(MatchRegexp("No templates were found"))
				})
			})
		})

		Context("[CLI] When only invalid Capi Template(s) are available in the cluster", func() {
			It("Verify gitops outputs an error message related to an invalid template(s)", func() {

				noOfTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-invalid-capitemplate.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				By("Then I should see template table header", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					re := regexp.MustCompile(`cluster-invalid-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(stdOut, 1)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")
				})
			})
		})

		Context("[CLI] When both valid and invalid Capi Templates are available in the cluster", func() {
			It("Verify gitops outputs an error message related to an invalid template and lists the valid template", func() {

				noOfTemplates := 3
				By("Apply/Install valid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 2
				By("Apply/Install invalid CAPITemplate", func() {
					invalid_captemplate := gitopsTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "capi-server-v1-invalid-capitemplate.yaml")
					templateFiles = append(templateFiles, invalid_captemplate...)
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				By("Then I should see template table header", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows for invalid template", func() {
					re := regexp.MustCompile(`cluster-invalid-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(stdOut, noOfInvalidTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfInvalidTemplates), "The number of listed invalid templates should be equal to number of templates created")
				})

				By("And I should see template rows for valid template", func() {
					re := regexp.MustCompile(`eks-fargate-template-[\d]+\s+aws\s+This is eks fargate template-[\d]+`)
					matched_list := re.FindAllString(stdOut, noOfTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed valid templates should be equal to number of templates created")
				})
			})

			It("Verify gitops reports an error when listing template parameters of invalid template from template library", func() {

				noOfTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-invalid-capitemplate.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates cluster-invalid-template-0 --list-parameters`)

				By("Then I should not see template parameter table header", func() {
					Eventually(stdOut).ShouldNot(MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see error message related to invalid template", func() {
					Eventually(stdErr).Should(MatchRegexp(`Error: unable to retrieve parameters.+`))
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify gitops can list templates from template library", func() {

				noOfTemplates := 5
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-template-azure.yaml")

				stdOut, stdErr = runGitopsCommand(`get templates`)

				By("Then I should see template table header", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					re := regexp.MustCompile(`azure-capi-quickstart-template-[\d]+\s+azure\s+This is Azure capi quick start template-[\d]+`)
					matched_list := re.FindAllString(stdOut, noOfTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")

					// Testing templates are ordered
					expected_list := make([]string, noOfTemplates)
					for i := 0; i < noOfTemplates; i++ {
						expected_list[i] = fmt.Sprintf("azure-capi-quickstart-template-%d", i)
					}
					sort.Strings(expected_list)

					for i := 0; i < noOfTemplates; i++ {
						Expect(matched_list[i]).Should(ContainSubstring(expected_list[i]))
					}
				})

				stdOut, stdErr = runGitopsCommand("get templates --namespace foo")

				// By("Then I should see an error message", func() {
				// 	Eventually(stdErr).Should(MatchRegexp(`No templates were found`))
				// })
			})

			It("Verify gitops can list filtered templates from template library", func() {
				awsTemplateCount := 2
				eksFargateTemplateCount := 2
				capdTemplateCount := 5
				totalTemplateCount := awsTemplateCount + eksFargateTemplateCount + capdTemplateCount
				By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "capi-template-capd.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-aws.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-eks-fargate.yaml")...)
				})

				stdOut, stdErr = runGitopsCommand(`get templates`)

				By("And I should see template list table header", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					re := regexp.MustCompile(`eks-fargate-template-[\d]+\s+aws\s+This is eks fargate template-[\d]+.+`)
					matched_list := re.FindAllString(stdOut, eksFargateTemplateCount)
					Eventually(len(matched_list)).Should(Equal(eksFargateTemplateCount), "The number of listed templates should be equal to number of templates created")
				})

				By("And I should see ordered list of templates", func() {
					expected_list := make([]string, totalTemplateCount)
					for i := 0; i < awsTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("aws-cluster-template-%d", i)
					}
					for i := 0; i < capdTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("cluster-template-development-%d", i)
					}
					for i := 0; i < eksFargateTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("eks-fargate-template-%d", i)
					}
					sort.Strings(expected_list)

					for i := 0; i < totalTemplateCount; i++ {
						Eventually(stdOut).Should(MatchRegexp(fmt.Sprintf(`%s\s+.*`, expected_list[i])))
					}

				})

				stdOut, stdErr = runGitopsCommand(`get templates --provider aws`)

				By("And I should see templates list filtered by provider", func() {
					awsCluster_list := make([]string, awsTemplateCount+eksFargateTemplateCount)
					for i := 0; i < awsTemplateCount; i++ {
						awsCluster_list[i] = fmt.Sprintf("aws-cluster-template-%d", i)
					}
					for i := 0; i < eksFargateTemplateCount; i++ {
						awsCluster_list[i] = fmt.Sprintf("eks-fargate-template-%d", i)
					}
					sort.Strings(awsCluster_list)
					for i := 0; i < awsTemplateCount+eksFargateTemplateCount; i++ {
						Eventually(stdOut).Should(MatchRegexp(fmt.Sprintf(`%s\s+.*`, awsCluster_list[i])))
					}

					for i := 0; i < 5; i++ {
						capd_template := fmt.Sprintf("cluster-template-development-%d", i)
						re := regexp.MustCompile(fmt.Sprintf(`%s\s+.*`, capd_template))
						Eventually((re.Find([]byte(stdOut)))).Should(BeNil())
					}

				})

				_, stdErr = runGitopsCommand(`get templates --provider foobar`)

				By("And I should see error message for invalid provider", func() {
					Eventually(stdErr).Should(MatchRegexp(`Error:\s+provider "foobar" is not valid.*`))
				})
			})

			It("Verify gitops can list template parameters of a template from template library", func() {

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
				})

				stdOut, stdErr = runGitopsCommand(`get templates eks-fargate-template-0 --list-parameters`)

				By("Then I should see template parameter table header", func() {
					Eventually(stdOut).Should(MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see parameter rows", func() {
					Eventually(stdOut).Should(MatchRegexp(`CLUSTER_NAME+\s+true\s+This is used for the cluster naming`))
					Expect(stdOut).Should(MatchRegexp(`AWS_REGION+\s+true\s+AWS Region to create cluster`))
					Expect(stdOut).Should(MatchRegexp(`AWS_SSH_KEY_NAME+\s+false\s+AWS ssh key name`))
					Expect(stdOut).Should(MatchRegexp(`KUBERNETES_VERSION+\s+false`))
				})
			})
		})

		Context("[CLI] When no infrastructure provider credentials are available in the management cluster", func() {
			It("Verify gitops lists no credentials", func() {
				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				By("Then gitops lists no credentials", func() {
					Eventually(stdOut).Should(MatchRegexp("No credentials were found"))
				})
			})
		})

		Context("[CLI] When infrastructure provider credentials are available in the management cluster", func() {
			It("Verify gitops can list credentials present in the management cluster", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AWS")
				defer gitopsTestRunner.DeleteIPCredentials("AZURE")

				By("And create AWS credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
				})

				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				By("Then gitops lists AWS credentials", func() {
					Eventually(stdOut).Should(MatchRegexp(`aws-test-identity`))
					Eventually(stdOut).Should(MatchRegexp(`test-role-identity`))
				})

				By("And create AZURE credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				stdOut, stdErr = runGitopsCommand(`get credentials`, ASSERTION_1MINUTE_TIME_OUT)

				By("Then gitops lists AZURE credentials", func() {
					Eventually(stdOut).Should(MatchRegexp(`azure-cluster-identity`))
				})
			})
		})

		Context("[CLI] When no clusters are available in the management cluster", func() {
			It("Verify gitops lists no clusters", func() {
				By("And gitops state is reset", func() {
					gitopsTestRunner.ResetControllers("enterprise")
					gitopsTestRunner.VerifyWegoPodsRunning()
				})

				stdOut, stdErr = runGitopsCommand(`get cluster`)

				By("Then gitops lists no clusters", func() {
					Eventually(stdOut).Should(MatchRegexp(`management\s+Ready`))
				})
			})
		})

		Context("[CLI] When profiles are available in the management cluster", func() {
			It("Verify gitops can list profiles from profile repository", func() {
				stdOut, stdErr = runGitopsCommand(`get profiles`)

				By("Then gitops lists no clusters", func() {
					Eventually(stdOut).Should(MatchRegexp(`cert-manager\s+A Weaveworks Helm chart for the Certificate Profile\s+0.0.7`))
					Eventually(stdOut).Should(MatchRegexp(`weave-policy-agent\s+A Weaveworks Helm chart for Kubernetes to configure the policy agent\s+0.3.1`))
					Eventually(stdOut).Should(MatchRegexp(`podinfo\s+Podinfo Helm chart for Kubernetes\s+6.0.1,6.0`))
				})
			})
		})

		Context("[CLI] When entitlement is available in the cluster", func() {
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

				matcher := BeFalse
				if beFound {
					matcher = BeTrue
				}

				resourceName = "templates"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "credentials"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "clusters"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
			}

			JustBeforeEach(func() {

				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
				gitopsTestRunner.CreateIPCredentials("AWS")
			})

			JustAfterEach(func() {
				By("When I apply the valid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("invalid", false)
				})

				// Login to the dashbord because the logout automatically when the cluster service restarts for entitlement checking
				loginUser()
			})

			It("Verify cluster service acknowledges the entitlement presences", func() {

				By("When I delete the entitlement", func() {
					Expect(gitopsTestRunner.KubectlDelete([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to delete entitlement secret")
				})

				By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(HaveOccurred(), "Failed restart deployment successfully")
				})

				By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)
				})

				By("When I apply the expired entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-expired.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				By("When I apply the invalid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-invalid.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})

	})
}
