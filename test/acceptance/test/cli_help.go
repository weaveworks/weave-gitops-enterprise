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

	By("Then I should see help message printed for gitops", func() {
		Eventually(session).Should(gbytes.Say("Command line utility for managing Kubernetes applications via GitOps"))
	})

	By("And Usage category", func() {
		Eventually(session).Should(gbytes.Say("Usage:"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("gitops [command]"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("To learn more, you can find our documentation at"))
	})

	By("And Available-Commands category", func() {
		Eventually(session).Should(gbytes.Say("Available Commands:"))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`get[\s]+Display one or many Weave GitOps resources`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`upgrade[\s]+Upgrade to Weave GitOps Enterprise`))
	})

	By("And Flags category", func() {
		Eventually(session).Should(gbytes.Say("Flags:"))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string[\s]+.+`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for gitops`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`--namespace string[\s]`))
	})

	By("And command help usage", func() {
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Use "gitops \[command\] --help".+`))
	})

}

func DescribeCliHelp() {
	var _ = Describe("Gitops Help Tests", func() {

		var session *gexec.Session
		var err error

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(fileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
			})
		})

		Context("[CLI] When gitops binary is available", func() {
			It("Verify that gitops displays help text when provided with the wrong flag", func() {

				By("When I run 'gitops foo'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "foo")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see gitops error message", func() {
					Eventually(session.Err).Should(gbytes.Say("Error: unknown command \"foo\" for \"gitops\""))
					// Eventually(session.Err).Should(gbytes.Say("Run 'gitops --help' for usage."))
				})
			})

			It("Verify that gitops help flag prints the help text", func() {

				By("When I run the command 'gitops --help' ", func() {
					command := exec.Command(GITOPS_BIN_PATH, "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				verifyUsageText(session)
			})

			It("Verify that gitops command prints the help text", func() {

				By("When I run the command 'gitops'", func() {
					command := exec.Command(GITOPS_BIN_PATH)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				verifyUsageText(session)

			})

			It("Verify that gitops command prints the help text for get command", func() {

				By("When I run the command 'gitops get --help' ", func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Display one or many Weave GitOps resources"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops get.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops get templates"))
				})

				By("And Available commands category", func() {
					Eventually(session).Should(gbytes.Say("Available Commands:"))
					Eventually(session).Should(gbytes.Say(`cluster[\s]+.+`))
					Eventually(session).Should(gbytes.Say(`credential[\s]+.+`))
					Eventually(session).Should(gbytes.Say(`template[\s]+.+`))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for get`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})
			})

			It("Verify that gitops command prints the sub help text for the get templates command", func() {

				By("When I run the command 'gitops get templates --help'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Display one or many CAPI templates"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops get template.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops get templates --provider.+"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`--list-parameters[\s]+.+`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})

			})

			It("Verify that gitops command prints the sub help text for the get credentials command", func() {

				By("When I run the command 'gitops get credentials --help'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "credentials", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Get CAPI credentials"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops get credential.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops get credentials"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for credential`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-v, --verbose[\s]+.+`))
				})

			})

			It("Verify that gitops command prints the sub help text for the get clusters command", func() {

				By("When I run the command 'gitops get clusters --help'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "clusters", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Display one or many CAPI clusters"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops get cluster.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops get cluster <cluster-name>.+"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`--kubeconfig[\s]+.+`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})
			})

			It("Verify that gitops command prints the help text for add command", func() {

				By("When I run the command 'gitops add --help' ", func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Add a new Weave GitOps resource"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops add.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops add cluster"))
				})

				By("And Available commands category", func() {
					Eventually(session).Should(gbytes.Say("Available Commands:"))
					Eventually(session).Should(gbytes.Say(`cluster[\s]+.+`))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for add`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`namespace string\s+.+`))
				})
			})

			It("Verify that gitops command prints the sub help text for the add cluster command", func() {

				By("When I run the command 'gitops add cluster --help'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster", "--help")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(session).Should(gbytes.Say("Add a new cluster using a CAPI template"))
				})

				By("And Usage category", func() {
					Eventually(session).Should(gbytes.Say("Usage:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp("gitops add cluster.+"))
				})

				By("And Examples category", func() {
					Eventually(session).Should(gbytes.Say("Examples:"))
					Eventually(session).Should(gbytes.Say("gitops add cluster --from-template.+"))
				})

				By("And Flags category", func() {
					Eventually(session).Should(gbytes.Say("Flags:"))

					output := string(session.Wait().Out.Contents())

					Eventually(output).Should(MatchRegexp(`--base string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--branch string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--commit-message string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--description string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--dry-run[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--from-template string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`-h, --help[\s]+help for cluster`))
					Eventually(output).Should(MatchRegexp(`--set strings[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--set-credentials string[\s].+`))
					Eventually(output).Should(MatchRegexp(`--title string[\s]+.+`))
					Eventually(output).Should(MatchRegexp(`--url string[\s]+.+`))
				})

				By("And  Global Flags category", func() {
					Eventually(session).Should(gbytes.Say("Global Flags:"))
					Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-e, --endpoint string\s+.+`))
				})

			})
		})

		Context("[CLI] When gitops command required parameters are missing", func() {
			It("Verify that gitops displays error text when listing parameters without specifying a template", func() {

				By(fmt.Sprintf("When I run 'gitops get templates --list-parameters --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--list-parameters", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see gitops error message", func() {
					Eventually(session.Err).Should(gbytes.Say("Error: template name is required"))
				})
			})

			It("Verify that gitops displays error text when listing templates without specifying a provider name", func() {

				By(fmt.Sprintf("When I run 'gitops get templates --provider --endpoint %s'", CAPI_ENDPOINT_URL), func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--provider", "--endpoint", CAPI_ENDPOINT_URL)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see gitops error message", func() {
					Eventually(session.Err).Should(gbytes.Say("Error"))
				})
			})

			It("Verify that gitops displays error text when performing actions on resources without specifying api endpoint", func() {

				By("When I run 'gitops get templates'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "get", "templates", "--provider")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see gitops error message", func() {
					Eventually(session.Err).Should(gbytes.Say(`Error.+needs an argument.+`))
				})

				By("When I run 'gitops add cluster'", func() {
					command := exec.Command(GITOPS_BIN_PATH, "add", "cluster")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
				})

				By("Then I should see gitops error message", func() {
					Eventually(session.Err).Should(gbytes.Say("Error"))
				})
			})
		})
	})
}
