package cmd

import (
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/templates"
	"k8s.io/cli-runtime/pkg/printers"
)

func templatesRenderCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "render",
		Short:         "Render CAPI template",
		Example:       `mccp templates render <template-name>`,
		RunE:          getTemplatesRenderCmdRun(client),
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&templatesRenderCmdFlags.ListTemplateParameters, "list-parameters", false, "The CAPI templates HTTP API endpoint")
	cmd.PersistentFlags().StringSliceVar(&templatesRenderCmdFlags.ParameterValues, "set", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().BoolVar(&templatesRenderCmdFlags.CreatePullRequest, "create-pr", false, "Indicates whether to create a pull request for the CAPI template")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.RepositoryURL, "pr-repo", "", "The repository to open a pull request against")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.BaseBranch, "pr-base", "", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.HeadBranch, "pr-branch", "", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.Title, "pr-title", "", "The title of the pull request")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.Description, "pr-description", "", "The description of the pull request")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.CommitMessage, "pr-commit-message", "", "The commit message to use when adding the CAPI template")
	cmd.PersistentFlags().BoolVar(&templatesRenderCmdFlags.ListCredentials, "list-credentials", false, "Indicates whether to list existing cluster credentials")
	cmd.PersistentFlags().StringVar(&templatesRenderCmdFlags.Credentials, "set-credentials", "", "Set credentials value on the command line")

	return cmd
}

type templatesRenderFlags struct {
	ListTemplateParameters bool
	ParameterValues        []string
	CreatePullRequest      bool
	RepositoryURL          string
	BaseBranch             string
	HeadBranch             string
	Title                  string
	Description            string
	CommitMessage          string
	ListCredentials        bool
	Credentials            string
}

var templatesRenderCmdFlags templatesRenderFlags

func getTemplatesRenderCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		r, err := adapters.NewHttpClient(endpoint, client)
		if err != nil {
			return err
		}

		if templatesRenderCmdFlags.ListTemplateParameters {
			w := printers.GetNewTabWriter(os.Stdout)
			defer w.Flush()
			return templates.ListTemplateParameters(args[0], r, w)
		}
		if templatesRenderCmdFlags.ListCredentials {
			w := printers.GetNewTabWriter(os.Stdout)
			defer w.Flush()
			return templates.ListCredentials(r, w)
		}

		vals := make(map[string]string)
		for _, v := range templatesRenderCmdFlags.ParameterValues {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 2 {
				vals[kv[0]] = kv[1]
			}
		}
		creds := templates.Credentials{}
		if templatesRenderCmdFlags.Credentials != "" {
			creds, err = r.RetrieveCredentialsByName(templatesRenderCmdFlags.Credentials)
			if err != nil {
				return err
			}
		}
		if templatesRenderCmdFlags.CreatePullRequest {
			return templates.CreatePullRequest(templates.CreatePullRequestForTemplateParams{
				TemplateName:    args[0],
				ParameterValues: vals,
				RepositoryURL:   templatesRenderCmdFlags.RepositoryURL,
				HeadBranch:      templatesRenderCmdFlags.HeadBranch,
				BaseBranch:      templatesRenderCmdFlags.BaseBranch,
				Title:           templatesRenderCmdFlags.Title,
				Description:     templatesRenderCmdFlags.Description,
				CommitMessage:   templatesRenderCmdFlags.CommitMessage,
				Credentials:     creds,
			}, r, os.Stdout)
		}
		return templates.RenderTemplate(args[0], vals, creds, r, os.Stdout)
	}
}
