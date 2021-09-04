package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/spf13/cobra"
	"github.com/weaveworks/libgitops/pkg/serializer"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/entitlement"
	"k8s.io/kubectl/pkg/scheme"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an entitlement for WeGO EE",
	PreRun: func(cmd *cobra.Command, args []string) {
		if durationYears <= 0 && durationMonths <= 0 && durationDays <= 0 {
			log.Fatal("The duration of the entitlements has not been set. Please use the flags '-y', '-m' or '-d' to specify a duration.")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		token, err := entitlement.CreateEntitlement(privateKeyFilename, durationYears, durationMonths, durationDays, customerEmail)
		cobra.CheckErr(err)
		secret := entitlement.WrapEntitlementInSecret(token, secretName, secretNamespace)

		var buf bytes.Buffer
		s := serializer.NewSerializer(scheme.Scheme, nil)
		fw := serializer.NewYAMLFrameWriter(&buf)
		s.Encoder().Encode(fw, secret)

		gs := git.NewGitProviderService()

		path := fmt.Sprintf("entitlements/%s.yaml", customerEmail)
		content := buf.String()
		res, err := gs.WriteFilesToBranchAndCreatePullRequest(cmd.Context(), git.WriteFilesToBranchAndCreatePullRequestRequest{
			GitProvider: git.GitProvider{
				Token:    githubToken,
				Type:     "github",
				Hostname: "github.com",
			},
			RepositoryURL: githubRepo,
			HeadBranch:    customerEmail,
			BaseBranch:    githubRepoBaseBranch,
			Title:         fmt.Sprintf("New entitlement for %s", customerEmail),
			Description:   fmt.Sprintf("Adds a new entitlement for %s", customerEmail),
			CommitMessage: fmt.Sprintf("Add new entitlement for %s", customerEmail),
			Files: []gitprovider.CommitFile{
				{
					Path:    &path,
					Content: &content,
				},
			},
		})
		cobra.CheckErr(err)
		fmt.Printf("New entitlement PR created at: %s\n", res.WebURL)
	},
}

var (
	privateKeyFilename   string
	customerEmail        string
	durationYears        int
	durationMonths       int
	durationDays         int
	secretName           string
	secretNamespace      string
	githubToken          string
	githubRepo           string
	githubRepoBaseBranch string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&privateKeyFilename, "private-key-filename", "p", "", "The private key to use for signing the entitlement")
	generateCmd.MarkFlagRequired("private-key-filename")
	generateCmd.Flags().StringVarP(&customerEmail, "customer-email", "c", "", "The email of the customer that the entitlement is generated for")
	generateCmd.MarkFlagRequired("customer-email")
	generateCmd.Flags().IntVarP(&durationYears, "duration-years", "y", 0, "Number of years to use for the duration of the entitlement")
	generateCmd.Flags().IntVarP(&durationMonths, "duration-months", "m", 0, "Number of months to use for the duration of the entitlement")
	generateCmd.Flags().IntVarP(&durationDays, "duration-days", "d", 0, "Number of days to use for the duration of the entitlement")
	generateCmd.Flags().StringVarP(&secretName, "name", "n", "wego-ee-entitlement", "The name of the secret")
	generateCmd.Flags().StringVar(&secretNamespace, "namespace", "default", "The namespace of the secret")
	generateCmd.Flags().StringVar(&githubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub token to use to create a PR that adds the generated entitlement to the entitlements repository")
	generateCmd.Flags().StringVar(&githubRepo, "github-repo", "https://github.com/weaveworks/weave-gitops-enterprise-entitlements", "The GitHub entitlements repository")
	generateCmd.Flags().StringVar(&githubRepoBaseBranch, "github-repo-base-branch", "main", "The GitHub entitlements repository base branch")
}
