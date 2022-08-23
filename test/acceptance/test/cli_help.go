package acceptance

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func verifyUsageText(output string) {

	By("Then I should see help message printed for gitops", func() {
		Eventually(output).Should(MatchRegexp("Command line utility for managing Kubernetes applications via GitOps"))
	})

	By("And Usage category", func() {
		Eventually(output).Should(MatchRegexp("Usage:"))
		Eventually(output).Should(MatchRegexp("gitops [command]"))
		Eventually(output).Should(MatchRegexp("To learn more, you can find our documentation at"))
	})

	By("And Available-Commands category", func() {
		Eventually(output).Should(MatchRegexp("Available Commands:"))
		Eventually(output).Should(MatchRegexp(`get[\s]+Display one or many Weave GitOps resources`))
		Eventually(output).Should(MatchRegexp(`upgrade[\s]+Upgrade to Weave GitOps Enterprise`))
	})

	By("And Flags category", func() {
		Eventually(output).Should(MatchRegexp("Flags:"))
		Eventually(output).Should(MatchRegexp(`-e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL[\s]+.+`))
		Eventually(output).Should(MatchRegexp(`--insecure-skip-tls-verify [\s]+.+`))
		Eventually(output).Should(MatchRegexp(`--namespace string[\s]+.+`))
		Eventually(output).Should(MatchRegexp(`-p, --password WEAVE_GITOPS_PASSWORD[\s]+.+`))
		Eventually(output).Should(MatchRegexp(`-u, --username WEAVE_GITOPS_USERNAME[\s]+.+`))
	})

	By("And command help usage", func() {
		Eventually(output).Should(MatchRegexp(`Use "gitops \[command\] --help".+`))
	})

}

func verifyGlobalFlags(stdOut string) {
	By("And  Global Flags category", func() {
		Eventually(stdOut).Should(MatchRegexp("Global Flags:"))
		Eventually(stdOut).Should(MatchRegexp(`-e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL[\s]+.+`))
		Eventually(stdOut).Should(MatchRegexp(`--insecure-skip-tls-verify [\s]+.+`))
		Eventually(stdOut).Should(MatchRegexp(`--namespace string[\s]+.+`))
		Eventually(stdOut).Should(MatchRegexp(`-p, --password WEAVE_GITOPS_PASSWORD[\s]+.+`))
		Eventually(stdOut).Should(MatchRegexp(`-u, --username WEAVE_GITOPS_USERNAME[\s]+.+`))
	})
}

func DescribeCliHelp() {
	var _ = Describe("Gitops Help Tests", func() {
		var stdOut string
		var stdErr string

		BeforeEach(func() {

		})

		Context("[CLI] When gitops binary is available", func() {
			It("Verify that gitops displays help text when provided with the wrong flag", func() {

				By("When I run 'gitops foo'", func() {
					stdOut, stdErr = runGitopsCommand("foo")
				})

				By("Then I should see gitops error message", func() {
					Eventually(stdErr).Should(MatchRegexp("Error: unknown command \"foo\" for \"gitops\""))
					// Eventually(stdErr).Should(MatchRegexp("Run 'gitops --help' for usage."))
				})
			})

			It("Verify that gitops help flag prints the help text", func() {

				By("When I run the command 'gitops --help' ", func() {
					stdOut, stdErr = runGitopsCommand("--help")
				})

				verifyUsageText(stdOut)
			})

			It("Verify that gitops command prints the help text", func() {

				By("When I run the command 'gitops'", func() {
					stdOut, stdErr = runGitopsCommand("")
				})

				verifyUsageText(stdOut)

			})

			It("Verify that gitops command prints the help text for get command", func() {

				By("When I run the command 'gitops get --help' ", func() {
					stdOut, stdErr = runGitopsCommand("get --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Display one or many Weave GitOps resources"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get templates"))
				})

				By("And Available commands category", func() {
					Eventually(stdOut).Should(MatchRegexp("Available Commands:"))
					Eventually(stdOut).Should(MatchRegexp(`bcrypt-hash[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`cluster[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`credential[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`profile[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`template[\s]+.+`))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+help for get`))
				})

				verifyGlobalFlags(stdOut)
			})

			It("Verify that gitops command prints the sub help text for the get templates command", func() {

				By("When I run the command 'gitops get templates --help'", func() {
					stdOut, stdErr = runGitopsCommand("get templates --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Display one or many CAPI templates"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get template.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get templates --provider.+"))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`--list-parameters[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`--list-profiles [\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`--provider string[\s]+.+`))
				})

				verifyGlobalFlags(stdOut)

			})

			It("Verify that gitops command prints the sub help text for the get credentials command", func() {

				By("When I run the command 'gitops get credentials --help'", func() {
					stdOut, stdErr = runGitopsCommand("get credentials --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Get CAPI credentials"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get credential.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get credentials"))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+help for credential`))
				})

				verifyGlobalFlags(stdOut)

			})

			It("Verify that gitops command prints the sub help text for the get clusters command", func() {

				By("When I run the command 'gitops get clusters --help'", func() {
					stdOut, stdErr = runGitopsCommand("get clusters --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Display one or many CAPI clusters"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get cluster.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get cluster <cluster-name>.+"))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+help for cluster`))
					Eventually(stdOut).Should(MatchRegexp(`--print-kubeconfig[\s]+.+`))
				})

				verifyGlobalFlags(stdOut)
			})

			It("Verify that gitops command prints the sub help text for the get profile command", func() {

				By("When I run the command 'gitops get profile --help'", func() {
					stdOut, stdErr = runGitopsCommand("get profiles --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Show information about available profiles"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get profile.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops get profiles"))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+help for profile`))
				})

				verifyGlobalFlags(stdOut)
			})

			It("Verify that gitops command prints the help text for add command", func() {

				By("When I run the command 'gitops add --help' ", func() {
					stdOut, stdErr = runGitopsCommand("add --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Add a new Weave GitOps resource"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops add.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops add cluster"))
				})

				By("And Available commands category", func() {
					Eventually(stdOut).Should(MatchRegexp("Available Commands:"))
					Eventually(stdOut).Should(MatchRegexp(`cluster[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`profile[\s]+.+`))
					Eventually(stdOut).Should(MatchRegexp(`terraform[\s]+.+`))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))
					Eventually(stdOut).Should(MatchRegexp(`-h, --help[\s]+help for add`))
				})

				verifyGlobalFlags(stdOut)
			})

			It("Verify that gitops command prints the sub help text for the add cluster command", func() {

				By("When I run the command 'gitops add cluster --help'", func() {
					stdOut, stdErr = runGitopsCommand("add cluster --help")
				})

				By("Then I should see help message printed with the command discreption", func() {
					Eventually(stdOut).Should(MatchRegexp("Add a new cluster using a CAPI template"))
				})

				By("And Usage category", func() {
					Eventually(stdOut).Should(MatchRegexp("Usage:"))
					Eventually(stdOut).Should(MatchRegexp("gitops add cluster.+"))
				})

				By("And Examples category", func() {
					Eventually(stdOut).Should(MatchRegexp("Examples:"))
					Eventually(stdOut).Should(MatchRegexp("gitops add cluster --from-template.+"))
				})

				By("And Flags category", func() {
					Eventually(stdOut).Should(MatchRegexp("Flags:"))

					output := stdOut

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

				verifyGlobalFlags(stdOut)
			})
		})

		It("Verify that gitops command prints the sub help text for the add profile command", func() {

			By("When I run the command 'gitops add profile --help'", func() {
				stdOut, stdErr = runGitopsCommand("add profile --help")
			})

			By("Then I should see help message printed with the command discreption", func() {
				Eventually(stdOut).Should(MatchRegexp("Add a profile to a cluster"))
			})

			By("And Usage category", func() {
				Eventually(stdOut).Should(MatchRegexp("Usage:"))
				Eventually(stdOut).Should(MatchRegexp("gitops add profile.+"))
			})

			By("And Examples category", func() {
				Eventually(stdOut).Should(MatchRegexp("Examples:"))
				Eventually(stdOut).Should(MatchRegexp("gitops add profile --name=.+"))
			})

			By("And Flags category", func() {
				Eventually(stdOut).Should(MatchRegexp("Flags:"))

				output := stdOut

				Eventually(output).Should(MatchRegexp(`--auto-merge[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--base string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--branch string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--cluster string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--commit-message string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--config-repo string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--description string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`-h, --help[\s]+help for profile`))
				Eventually(output).Should(MatchRegexp(`--name string[\s]+.+`))
				Eventually(output).Should(MatchRegexp(`--title string[\s].+`))
				Eventually(output).Should(MatchRegexp(`--version string[\s]+.+`))
			})

			verifyGlobalFlags(stdOut)
		})

		Context("[CLI] When gitops command required parameters are missing", func() {
			It("Verify that gitops displays error text when listing parameters without specifying a template", func() {
				stdOut, stdErr = runGitopsCommand("get templates --list-parameters")

				By("Then I should see gitops error message", func() {
					Eventually(stdErr).Should(MatchRegexp("Error: template name is required"))
				})
			})

			It("Verify that gitops displays error text when listing templates without specifying a provider name", func() {

				By(fmt.Sprintf("When I run 'gitops get templates --provider --endpoint %s'", capi_endpoint_url), func() {
					stdOut, stdErr = runGitopsCommand("get templates --provider")
				})

				By("Then I should see gitops error message", func() {
					Eventually(stdErr).Should(MatchRegexp("Error"))
				})
			})

			It("Verify that gitops displays error text when performing actions on resources without specifying api endpoint", func() {

				By("When I run 'gitops get templates'", func() {
					stdOut, stdErr = runGitopsCommand("get templates --provider")
				})

				By("Then I should see gitops error message", func() {
					Eventually(stdErr).Should(MatchRegexp(`Error.+needs an argument.+`))
				})

				By("When I run 'gitops add cluster'", func() {
					stdOut, stdErr = runGitopsCommand("add cluster")
				})

				By("Then I should see gitops error message", func() {
					Eventually(stdErr).Should(MatchRegexp("Error"))
				})
			})
		})
	})
}
