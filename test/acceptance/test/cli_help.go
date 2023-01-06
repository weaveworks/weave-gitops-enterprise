package acceptance

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func verifyUsageText(output string) {

	ginkgo.By("Then I should see help message printed for gitops", func() {
		gomega.Eventually(output).Should(gomega.MatchRegexp("Command line utility for managing Kubernetes applications via GitOps"))
	})

	ginkgo.By("And Usage category", func() {
		gomega.Eventually(output).Should(gomega.MatchRegexp("Usage:"))
		gomega.Eventually(output).Should(gomega.MatchRegexp("gitops [command]"))
		gomega.Eventually(output).Should(gomega.MatchRegexp("To learn more, you can find our documentation at"))
	})

	ginkgo.By("And Available-Commands category", func() {
		gomega.Eventually(output).Should(gomega.MatchRegexp("Available Commands:"))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`get[\s]+Display one or many Weave GitOps resources`))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`upgrade[\s]+Upgrade to Weave GitOps Enterprise`))
	})

	ginkgo.By("And Flags category", func() {
		gomega.Eventually(output).Should(gomega.MatchRegexp("Flags:"))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`-e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL[\s]+.+`))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`--insecure-skip-tls-verify [\s]+.+`))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`--namespace string[\s]+.+`))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`-p, --password WEAVE_GITOPS_PASSWORD[\s]+.+`))
		gomega.Eventually(output).Should(gomega.MatchRegexp(`-u, --username WEAVE_GITOPS_USERNAME[\s]+.+`))
	})

	ginkgo.By("And command help usage", func() {
		gomega.Eventually(output).Should(gomega.MatchRegexp(`Use "gitops \[command\] --help".+`))
	})

}

func verifyGlobalFlags(stdOut string) {
	ginkgo.By("And  Global Flags category", func() {
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Global Flags:"))
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL[\s]+.+`))
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--insecure-skip-tls-verify [\s]+.+`))
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--namespace string[\s]+.+`))
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-p, --password WEAVE_GITOPS_PASSWORD[\s]+.+`))
		gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-u, --username WEAVE_GITOPS_USERNAME[\s]+.+`))
	})
}

var _ = ginkgo.Describe("Gitops Help Tests", ginkgo.Label("cli", "help"), func() {
	var stdOut string
	var stdErr string

	ginkgo.BeforeEach(func() {

	})

	ginkgo.Context("When gitops binary is available", func() {
		ginkgo.It("Verify that gitops displays help text when provided with the wrong flag", func() {

			ginkgo.By("When I run 'gitops foo'", func() {
				stdOut, stdErr = runGitopsCommand("foo")
			})

			ginkgo.By("Then I should see gitops error message", func() {
				gomega.Eventually(stdErr).Should(gomega.MatchRegexp("Error: unknown command \"foo\" for \"gitops\""))
				// gomega.Eventually(stdErr).Should(gomega.MatchRegexp("Run 'gitops --help' for usage."))
			})
		})

		ginkgo.It("Verify that gitops help flag prints the help text", func() {

			ginkgo.By("When I run the command 'gitops --help' ", func() {
				stdOut, stdErr = runGitopsCommand("--help")
			})

			verifyUsageText(stdOut)
		})

		ginkgo.It("Verify that gitops command prints the help text", func() {

			ginkgo.By("When I run the command 'gitops'", func() {
				stdOut, stdErr = runGitopsCommand("")
			})

			verifyUsageText(stdOut)

		})

		ginkgo.It("Verify that gitops command prints the help text for get command", func() {

			ginkgo.By("When I run the command 'gitops get --help' ", func() {
				stdOut, stdErr = runGitopsCommand("get --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Display one or many Weave GitOps resources"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get templates"))
			})

			ginkgo.By("And Available commands category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Available Commands:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`bcrypt-hash[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`cluster[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`credential[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`profile[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`template[\s]+.+`))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+help for get`))
			})

			verifyGlobalFlags(stdOut)
		})

		ginkgo.It("Verify that gitops command prints the sub help text for the get templates command", func() {

			ginkgo.By("When I run the command 'gitops get templates --help'", func() {
				stdOut, stdErr = runGitopsCommand("get templates --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Display one or many CAPI templates"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get template.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get templates --provider.+"))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--list-parameters[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--list-profiles [\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--provider string[\s]+.+`))
			})

			verifyGlobalFlags(stdOut)

		})

		ginkgo.It("Verify that gitops command prints the sub help text for the get credentials command", func() {

			ginkgo.By("When I run the command 'gitops get credentials --help'", func() {
				stdOut, stdErr = runGitopsCommand("get credentials --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Get CAPI credentials"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get credential.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get credentials"))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+help for credential`))
			})

			verifyGlobalFlags(stdOut)

		})

		ginkgo.It("Verify that gitops command prints the sub help text for the get clusters command", func() {

			ginkgo.By("When I run the command 'gitops get clusters --help'", func() {
				stdOut, stdErr = runGitopsCommand("get clusters --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Display one or many CAPI clusters"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get cluster.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get cluster <cluster-name>.+"))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+help for cluster`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`--print-kubeconfig[\s]+.+`))
			})

			verifyGlobalFlags(stdOut)
		})

		ginkgo.It("Verify that gitops command prints the sub help text for the get profile command", func() {

			ginkgo.By("When I run the command 'gitops get profile --help'", func() {
				stdOut, stdErr = runGitopsCommand("get profiles --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Show information about available profiles"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get profile.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops get profiles"))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+help for profile`))
			})

			verifyGlobalFlags(stdOut)
		})

		ginkgo.It("Verify that gitops command prints the help text for add command", func() {

			ginkgo.By("When I run the command 'gitops add --help' ", func() {
				stdOut, stdErr = runGitopsCommand("add --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Add a new Weave GitOps resource"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add cluster"))
			})

			ginkgo.By("And Available commands category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Available Commands:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`cluster[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`profile[\s]+.+`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`terraform[\s]+.+`))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`-h, --help[\s]+help for add`))
			})

			verifyGlobalFlags(stdOut)
		})

		ginkgo.It("Verify that gitops command prints the sub help text for the add cluster command", func() {

			ginkgo.By("When I run the command 'gitops add cluster --help'", func() {
				stdOut, stdErr = runGitopsCommand("add cluster --help")
			})

			ginkgo.By("Then I should see help message printed with the command discreption", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Add a new cluster using a CAPI template"))
			})

			ginkgo.By("And Usage category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add cluster.+"))
			})

			ginkgo.By("And Examples category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add cluster --from-template.+"))
			})

			ginkgo.By("And Flags category", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))

				output := stdOut

				gomega.Eventually(output).Should(gomega.MatchRegexp(`--base string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--branch string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--commit-message string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--description string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--dry-run[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--from-template string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`-h, --help[\s]+help for cluster`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--set strings[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--set-credentials string[\s].+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--title string[\s]+.+`))
				gomega.Eventually(output).Should(gomega.MatchRegexp(`--url string[\s]+.+`))
			})

			verifyGlobalFlags(stdOut)
		})
	})

	ginkgo.It("Verify that gitops command prints the sub help text for the add profile command", func() {

		ginkgo.By("When I run the command 'gitops add profile --help'", func() {
			stdOut, stdErr = runGitopsCommand("add profile --help")
		})

		ginkgo.By("Then I should see help message printed with the command discreption", func() {
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Add a profile to a cluster"))
		})

		ginkgo.By("And Usage category", func() {
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Usage:"))
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add profile.+"))
		})

		ginkgo.By("And Examples category", func() {
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Examples:"))
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("gitops add profile --name=.+"))
		})

		ginkgo.By("And Flags category", func() {
			gomega.Eventually(stdOut).Should(gomega.MatchRegexp("Flags:"))

			output := stdOut

			gomega.Eventually(output).Should(gomega.MatchRegexp(`--auto-merge[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--base string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--branch string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--cluster string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--commit-message string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--config-repo string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--description string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`-h, --help[\s]+help for profile`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--name string[\s]+.+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--title string[\s].+`))
			gomega.Eventually(output).Should(gomega.MatchRegexp(`--version string[\s]+.+`))
		})

		verifyGlobalFlags(stdOut)
	})

	ginkgo.Context("When gitops command required parameters are missing", func() {

		ginkgo.It("Verify that gitops displays error text when listing parameters without specifying a template", func() {
			stdOut, stdErr = runGitopsCommand("get templates --list-parameters")

			ginkgo.By("Then I should see gitops error message", func() {
				gomega.Eventually(stdErr).Should(gomega.MatchRegexp("Error: template name is required"))
			})
		})

		ginkgo.It("Verify that gitops displays error text when listing templates without specifying a provider name", func() {

			ginkgo.By(fmt.Sprintf("When I run 'gitops get templates --provider --endpoint %s'", wgeEndpointUrl), func() {
				stdOut, stdErr = runGitopsCommand("get templates --provider")
			})

			ginkgo.By("Then I should see gitops error message", func() {
				gomega.Eventually(stdErr).Should(gomega.MatchRegexp("Error"))
			})
		})

		ginkgo.It("Verify that gitops displays error text when performing actions on resources without specifying api endpoint", func() {

			ginkgo.By("When I run 'gitops get templates'", func() {
				stdOut, stdErr = runGitopsCommand("get templates --provider")
			})

			ginkgo.By("Then I should see gitops error message", func() {
				gomega.Eventually(stdErr).Should(gomega.MatchRegexp(`Error.+needs an argument.+`))
			})

			ginkgo.By("When I run 'gitops add cluster'", func() {
				stdOut, stdErr = runGitopsCommand("add cluster")
			})

			ginkgo.By("Then I should see gitops error message", func() {
				gomega.Eventually(stdErr).Should(gomega.MatchRegexp("Error"))
			})
		})
	})
})
