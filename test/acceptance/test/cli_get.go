package acceptance

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func DescribeCliGet(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops Get Tests", func() {

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

		Context("[CLI] When no Capi Templates are available in the cluster", func() {
			It("Verify gitops lists no templates", func() {

				By(fmt.Sprintf(`And I run 'gitops get templates --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then gitops lists no templates", func() {
					Eventually(session).Should(gbytes.Say("No templates were found"))
				})
			})
		})

		Context("[CLI] When only invalid Capi Template(s) are available in the cluster", func() {
			It("Verify gitops outputs an error message related to an invalid template(s)", func() {

				noOfTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-invalid-capitemplate.yaml")
				})

				By(fmt.Sprintf(`And I run 'gitops get templates --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`cluster-invalid-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(output, 1)
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

				By(fmt.Sprintf(`And I run 'gitops get templates --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows for invalid template", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`cluster-invalid-template-[\d]+\s+.+Couldn't load template body.+`)
					matched_list := re.FindAllString(output, noOfInvalidTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfInvalidTemplates), "The number of listed invalid templates should be equal to number of templates created")
				})

				By("And I should see template rows for valid template", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`eks-fargate-template-[\d]+\s+aws\s+This is eks fargate template-[\d]+`)
					matched_list := re.FindAllString(output, noOfTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed valid templates should be equal to number of templates created")
				})
			})

			It("Verify gitops reports an error when listing template parameters of invalid template from template library", func() {

				noOfTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-invalid-capitemplate.yaml")
				})

				By(fmt.Sprintf(`And I run 'gitops get template cluster-invalid-template-0 --list-parameters --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "template", "cluster-invalid-template-0", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should not see template parameter table header", func() {
					Eventually(string(session.Wait().Out.Contents())).ShouldNot(MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see error message related to invalid template", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(`Error: unable to retrieve parameters.+`))
				})
			})
		})

		Context("[CLI] When Capi Templates are available in the cluster", func() {
			It("Verify gitops can list templates from template library", func() {

				noOfTemplates := 5
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-template-azure.yaml")

				By(fmt.Sprintf(`And I run 'gitops get templates --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`azure-capi-quickstart-template-[\d]+\s+azure\s+This is Azure capi quick start template-[\d]+`)
					matched_list := re.FindAllString(output, noOfTemplates)
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

				By(fmt.Sprintf(`When I run 'gitops get templates --namespace foo --endpoint %s'`, CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--namespace", "foo", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				// FIXME: issue 209
				// By("Then I should see an error message", func() {
				// 	Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				// })
			})

			It("Verify gitops can list filtered templates from template library", func() {
				awsTemplateCount := 2
				eksFargateTemplateCount := 2
				capdTemplateCount := 5
				totalTemplateCount := awsTemplateCount + eksFargateTemplateCount + capdTemplateCount
				By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "capi-server-v1-template-capd.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-aws.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-eks-fargate.yaml")...)
				})

				By(fmt.Sprintf("Then I run 'gitops get templates --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("And I should see template list table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+PROVIDER\s+DESCRIPTION\s+ERROR`))
				})

				By("And I should see template rows", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`eks-fargate-template-[\d]+\s+aws\s+This is eks fargate template-[\d]+.+`)
					matched_list := re.FindAllString(output, eksFargateTemplateCount)
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
						Eventually(session).Should(gbytes.Say(fmt.Sprintf(`%s\s+.*`, expected_list[i])))
					}

				})

				By(fmt.Sprintf("Then I run 'gitops get templates --provider aws --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--provider", "aws", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

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
						Eventually(session).Should(gbytes.Say(fmt.Sprintf(`%s\s+.*`, awsCluster_list[i])))
					}

					output := session.Wait().Out.Contents()
					for i := 0; i < 5; i++ {
						capd_template := fmt.Sprintf("cluster-template-development-%d", i)
						re := regexp.MustCompile(fmt.Sprintf(`%s\s+.*`, capd_template))
						Eventually((re.Find(output))).Should(BeNil())
					}

				})

				By(fmt.Sprintf("Then I run 'gitops get templates --provider foobar --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--provider", "foobar", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("And I should see error message for invalid provider", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`error:\s+.+provider \\"foobar\\" is not recognised.+`)
					Eventually((re.Match(output))).Should(BeTrue())
				})

			})

			It("Verify gitops can list template parameters of a template from template library", func() {

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				By(fmt.Sprintf("And I run gitops get templates cluster-template-development-0 --list-parameters --endpoint %s", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "cluster-template-development-0", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see template parameter table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+REQUIRED\s+DESCRIPTION\s+OPTIONS`))
				})

				By("And I should see parameter rows", func() {
					output := session.Wait().Out.Contents()
					re := regexp.MustCompile(`CLUSTER_NAME+\s+true\s+This is used for the cluster naming.\s+KUBERNETES_VERSION\s+false\s+Kubernetes version to use for the cluster\s+1.19.7, 1.19.8\s+NAMESPACE\s+false\s+Namespace to create the cluster in`)
					Eventually((re.Find(output))).ShouldNot(BeNil())

				})
			})
		})

		Context("[CLI] When no infrastructure provider credentials are available in the management cluster", func() {
			It("Verify gitops lists no credentials", func() {
				By(fmt.Sprintf("And I run 'gitops get credentials --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "credentials", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then gitops lists no credentials", func() {
					Eventually(session).Should(gbytes.Say("No credentials were found"))
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

				By(fmt.Sprintf("And I run 'gitops get credentials --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "credentials", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then gitops lists AWS credentials", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`aws-test-identity`))
					Eventually(string(output)).Should(MatchRegexp(`test-role-identity`))
				})

				By("And create AZURE credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				By(fmt.Sprintf("And I run 'gitops get credential --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "credential", "--endpoint", CAPI_ENDPOINT_URL)

					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then gitops lists AZURE credentials", func() {
					output := session.Wait().Out.Contents()
					Eventually(string(output)).Should(MatchRegexp(`azure-cluster-identity`))
				})
			})
		})

		Context("[CLI] When no clusters are available in the management cluster", func() {
			It("Verify gitops lists no clusters", func() {
				if getEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
					Skip("This test case runs only with sqlite")
				}

				By("And gitops state is reset", func() {
					_ = gitopsTestRunner.ResetDatabase()
					gitopsTestRunner.VerifyWegoPodsRunning()
					gitopsTestRunner.CheckClusterService(GetCapiEndpointUrl())
				})

				By(fmt.Sprintf("Then I run 'gitops get cluster --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "cluster", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then gitops lists no clusters", func() {
					Eventually(session).Should(gbytes.Say("No clusters found"))
				})
			})
		})

		Context("[CLI] When entitlement is available in the cluster", func() {
			var resourceName string
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					command := exec.Command(GITOPS_BIN_PATH, "get", resourceName, "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					msg := string(session.Wait().Err.Contents()) + " " + string(session.Wait().Out.Contents())

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
				log.Printf("Running 'gitops get %s --endpoint %s'", resourceName, CAPI_ENDPOINT_URL)
				Eventually(checkOutput, ASSERTION_DEFAULT_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "credentials"
				log.Printf("Running 'gitops get %s --endpoint %s'", resourceName, CAPI_ENDPOINT_URL)
				Eventually(checkOutput, ASSERTION_DEFAULT_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "clusters"
				log.Printf("Running 'gitops get %s --endpoint %s'", resourceName, CAPI_ENDPOINT_URL)
				Eventually(checkOutput, ASSERTION_DEFAULT_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
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
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("invalid", false)
				})

				gitopsTestRunner.DeleteIPCredentials("AWS")
			})

			It("Verify cluster service acknowledges the entitlement presences", func() {

				By("When I delete the entitlement", func() {
					Expect(gitopsTestRunner.KubectlDelete([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to delete entitlement secret")
				})

				By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(HaveOccurred(), "Failed restart deployment successfully")
				})

				By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)
				})

				By("When I apply the expired entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-expired.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				By("When I apply the invalid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-invalid.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})

	})
}
