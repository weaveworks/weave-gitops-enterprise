package acceptance

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func DescribeMccpCliList(mccpTestRunner MCCPTestRunner) {
	var _ = Describe("MCCP Template List Tests", func() {

		MCCP_BIN_PATH := GetMCCBinPath()
		CAPI_ENDPOINT_URL := GetCapiEndpointUrl()

		templateFiles := []string{}
		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a mccp binary installed on my local machine", func() {
				Expect(FileExists(MCCP_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", MCCP_BIN_PATH))
			})
		})

		AfterEach(func() {
			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
		})

		Context("When no Capi Templates are available in the cluster", func() {
			It("Verify mccp lists no templates", func() {

				By("And I run 'mccp templates list --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then mccp lists no templates", func() {
					Eventually(session).Should(gbytes.Say("No templates found"))
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify mccp can list templates from template library", func() {

				noOfTemplates := 50
				templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-capitemplate.yaml")

				By("And I run 'mccp templates list --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})
				By("Then I should see template table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+DESCRIPTION`))
				})

				By("And I should see template rows", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`cluster-template-[\d]+\s+This is test template [\d]+`)
					matched_list := re.FindAllString(output, noOfTemplates)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")

					// Testing templates are ordered
					expected_list := make([]string, noOfTemplates)
					for i := 0; i < noOfTemplates; i++ {
						expected_list[i] = fmt.Sprintf("cluster-template-%d", i)
					}
					sort.Strings(expected_list)

					for i := 0; i < noOfTemplates; i++ {
						Expect(matched_list[i]).Should(ContainSubstring(expected_list[i]))
					}
				})
			})
		})

		Context("When only invalid Capi Template(s) are available in the cluster", func() {
			XIt("Verify mccp outputs an error message related to an invalid template(s)", func() {

				By("Apply/Insall invalid CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
				})

				By("And I run 'mccp templates list --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see error message related to invalid template", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(fmt.Sprintf(`Error: unable to retrieve templates from "%s":*`, CAPI_ENDPOINT_URL)))
				})
			})
		})

		Context("When both valid and invalid Capi Templates are available in the cluster", func() {
			XIt("Verify mccp outputs an error message related to an invalid template and lists the valid template", func() {

				noOfTemplates := 3
				By("Apply/Insall valid CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-template-eks-fargate.yaml")
				})

				By("Apply/Insall invalid CAPITemplate", func() {
					invalid_captemplate := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
					templateFiles = append(templateFiles, invalid_captemplate...)
				})

				By("And I run 'mccp templates list --endpoint <capi-http-endpoint-url>'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "list", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see error message related to invalid template", func() {
					Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(fmt.Sprintf(`Error: unable to retrieve templates from "%s":*`, CAPI_ENDPOINT_URL)))
				})

				By("Then I should see template table header", func() {
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`NAME\s+DESCRIPTION`))
				})

				By("And I should see template rows", func() {
					output := string(session.Wait().Out.Contents())
					re := regexp.MustCompile(`cluster-template-[\d]+\s+This is test template [\d]+`)
					matched_list := re.FindAllString(output, 3)
					Eventually(len(matched_list)).Should(Equal(noOfTemplates), "The number of listed templates should be equal to number of templates created")
				})
			})
		})

	})
}
