package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func verifyUsageText(session *gexec.Session) {

	By("Then I should see help message printed with the product name", func() {
		Eventually(session).Should(gbytes.Say("MCCP CLI"))
	})

	By("And Usage category", func() {
		Eventually(session).Should(gbytes.Say("Usage:"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("mccp [command]"))
	})

	By("And Available-Commands category", func() {
		Eventually(session).Should(gbytes.Say("Available Commands:"))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`clusters[\s]+Interact with Kubernetes clusters`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`help[\s]+Help about any command`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`templates[\s]+Interact with CAPI templates`))
	})

	By("And Flags category", func() {
		Eventually(session).Should(gbytes.Say("Flags:"))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string[\s]+.+`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for mccp`))
	})

	By("And command help usage", func() {
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Use "mccp \[command\] --help".+`))
	})

}

func DescribeMccpCliHelp() {
	var _ = FDescribe("MCCP Help Tests", func() {

		MCCP_BIN_PATH := GetMCCBinPath()
		CAPI_ENDPOINT_URL := GetCapiEndpointUrl()

		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a mccp binary installed on my local machine", func() {
				Expect(FileExists(MCCP_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", MCCP_BIN_PATH))
			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp displays help text when provided with the wrong flag", func() {

				By("When I run 'mccp foo'", func() {
					command := exec.Command(MCCP_BIN_PATH, "foo")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				verifyUsageText(session)
			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp help flag prints the help text", func() {

				By("When I run the command 'mccp --help' ", func() {
					command := exec.Command(MCCP_BIN_PATH, "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				verifyUsageText(session)
			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp command prints the help text", func() {

				By("When I run the command 'mccp'", func() {
					command := exec.Command(MCCP_BIN_PATH)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				verifyUsageText(session)

			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp command prints the sub help text for the templates list command", func() {

				By("When I run the command 'mccp templates list help'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "list", "--help", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("List CAPI templates"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("mccp templates list [flags]"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("mccp templates list"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for list`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})

			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp command prints the sub help text for the render command", func() {

				By("When I run the command 'mccp templates render help'", func() {
					command := exec.Command(MCCP_BIN_PATH, "templates", "render", "--help", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Render CAPI template"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("mccp templates render [flags]"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("mccp templates render <template-name>"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))

					output := string(session.Wait().Out.Contents())

					Eventually(output).Should(MatchRegexp(`-h, --help[\s]+help for render`))
					Eventually(output).Should(MatchRegexp(`--create-pr[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--list-parameters[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--list-credentials [\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-base string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-branch string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-commit-message string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-description string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-repo string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--pr-title string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--set strings[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--set-credentials string[\s].+`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+The MCCP HTTP API endpoint`))
				})

			})
		})

		Context("[CLI] When mccp binary is available", func() {
			It("Verify that mccp command prints the sub help text for the clusters list command", func() {

				By("When I run the command 'mccp templates list help'", func() {
					command := exec.Command(MCCP_BIN_PATH, "clusters", "list", "--help", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("List Kubernetes clusters"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("mccp clusters list [flags]"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("mccp clusters list"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for list`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})
			})
		})
	})
}
