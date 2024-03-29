package clusters

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/internal"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

type clusterCommandFlags struct {
	DryRun            bool
	Template          string
	TemplateNamespace string
	ParameterValues   []string
	RepositoryURL     string
	BaseBranch        string
	HeadBranch        string
	Title             string
	Description       string
	CommitMessage     string
	Credentials       string
	Profiles          []string
}

var flags clusterCommandFlags

// AddCommand returns a cobra command that provides support for adding a cluster.
func AddCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Add a new cluster using a CAPI template",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster --from-template <template-name> --set key=val

# View a CAPI template populated with parameter values
# without creating a pull request for it
gitops add cluster --from-template <template-name> --set key=val --dry-run

# Add a new cluster supplied with profiles versions, namespaces and values files
gitops add cluster --from-template <template-name> \
--profile 'name=foo-profile,version=0.0.1,namespace=foo-system' --profile 'name=bar-profile,values=bar-values.yaml
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getClusterCmdPreRunE(&opts.Endpoint),
		RunE:          getClusterCmdRunE(opts, client),
	}

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "View the populated template without creating a pull request")
	cmd.Flags().StringVar(&flags.RepositoryURL, "url", "", "URL of remote repository to create the pull request")
	cmd.Flags().StringVar(&flags.Credentials, "set-credentials", "", "The CAPI credentials to use")
	cmd.Flags().StringArrayVar(&flags.Profiles, "profile", []string{}, "Set profiles values files on the command line (--profile 'name=foo-profile,version=0.0.1,namespace=foo-system' --profile 'name=bar-profile,namespace=bar-system,values=bar-values.yaml')")
	internal.AddTemplateFlags(cmd, &flags.Template, &flags.TemplateNamespace, &flags.ParameterValues)
	internal.AddPRFlags(cmd, &flags.HeadBranch, &flags.BaseBranch, &flags.Description, &flags.CommitMessage, &flags.Title)

	return cmd
}

func getClusterCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getClusterCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
		if err != nil {
			return err
		}

		vals := make(map[string]string)

		for _, v := range flags.ParameterValues {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 2 {
				vals[kv[0]] = kv[1]
			}
		}

		creds := templates.Credentials{}
		if flags.Credentials != "" {
			creds, err = client.RetrieveCredentialsByName(flags.Credentials)
			if err != nil {
				return err
			}
		}

		profilesValues, err := templates.ParseProfileFlags(flags.Profiles)
		if err != nil {
			return fmt.Errorf("error parsing profiles: %w", err)
		}

		if flags.DryRun {
			req := templates.RenderTemplateRequest{
				TemplateName:      flags.Template,
				TemplateKind:      templates.CAPITemplateKind,
				TemplateNamespace: flags.TemplateNamespace,
				Values:            vals,
				Credentials:       creds,
				Profiles:          profilesValues,
			}
			return templates.RenderTemplateWithParameters(req, client, os.Stdout)
		}

		if flags.RepositoryURL == "" {
			return cmderrors.ErrNoURL
		}

		url, err := gitproviders.NewRepoURL(flags.RepositoryURL)
		if err != nil {
			return fmt.Errorf("cannot parse url: %w", err)
		}

		token, err := internal.GetToken(url, os.LookupEnv)
		if err != nil {
			return err
		}

		params := templates.CreatePullRequestFromTemplateParams{
			GitProviderToken:  token,
			TemplateKind:      templates.CAPITemplateKind.String(),
			TemplateName:      flags.Template,
			TemplateNamespace: flags.TemplateNamespace,
			ParameterValues:   vals,
			RepositoryURL:     flags.RepositoryURL,
			HeadBranch:        flags.HeadBranch,
			BaseBranch:        flags.BaseBranch,
			Title:             flags.Title,
			Description:       flags.Description,
			CommitMessage:     flags.CommitMessage,
			Credentials:       creds,
			ProfileValues:     profilesValues,
		}

		return templates.CreatePullRequestFromTemplate(params, client, os.Stdout)
	}
}
