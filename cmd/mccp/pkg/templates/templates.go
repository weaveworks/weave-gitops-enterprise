package templates

import (
	"fmt"
	"io"
	"strings"
)

type TemplatesRetriever interface {
	Source() string
	RetrieveTemplates() ([]Template, error)
}

type TemplateParametersRetriever interface {
	Source() string
	RetrieveTemplateParameters(name string) ([]TemplateParameter, error)
}

type TemplateRenderer interface {
	RenderTemplateWithParameters(name string, parameters map[string]string, creds Credentials) (string, error)
}

// TemplatePullRequester implementers must return the web URI of the pull
// request.
type TemplatePullRequester interface {
	CreatePullRequestForTemplate(params CreatePullRequestForTemplateParams) (string, error)
}

type CredentialsRetriever interface {
	Source() string
	RetrieveCredentials() ([]Credentials, error)
}

type Template struct {
	Name        string
	Description string
}

type TemplateParameter struct {
	Name        string
	Description string
	Options     []string
}

type CreatePullRequestForTemplateParams struct {
	TemplateName    string
	ParameterValues map[string]string
	RepositoryURL   string
	HeadBranch      string
	BaseBranch      string
	Title           string
	Description     string
	CommitMessage   string
	Credentials     Credentials
}

type Credentials struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func ListTemplates(r TemplatesRetriever, w io.Writer) error {
	ts, err := r.RetrieveTemplates()
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "NAME\tDESCRIPTION\n")
		for _, t := range ts {
			fmt.Fprintf(w, "%s", t.Name)
			if t.Description != "" {
				fmt.Fprintf(w, "\t%s", t.Description)
			}
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates found.\n")

	return nil
}

func ListTemplateParameters(name string, r TemplateParametersRetriever, w io.Writer) error {
	ts, err := r.RetrieveTemplateParameters(name)
	if err != nil {
		return fmt.Errorf("unable to retrieve template parameters from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "NAME\tDESCRIPTION\tOPTIONS\n")

		for _, t := range ts {
			fmt.Fprintf(w, "%s", t.Name)
			if t.Description != "" {
				fmt.Fprintf(w, "\t%s", t.Description)
			}
			if t.Options != nil {
				optionsStr := strings.Join(t.Options, ", ")
				fmt.Fprintf(w, "\t%s", optionsStr)
			}
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No template parameters found.")

	return nil
}

func RenderTemplate(name string, parameters map[string]string, creds Credentials, r TemplateRenderer, w io.Writer) error {
	s, err := r.RenderTemplateWithParameters(name, parameters, creds)
	if err != nil {
		return fmt.Errorf("unable to render template: %w", err)
	}

	if s != "" {
		fmt.Fprint(w, s)
		return nil
	}

	fmt.Fprintf(w, "No template found.")

	return nil
}

func CreatePullRequest(params CreatePullRequestForTemplateParams, r TemplatePullRequester, w io.Writer) error {
	res, err := r.CreatePullRequestForTemplate(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	fmt.Fprintf(w, "Created pull request: %s\n", res)

	return nil
}

func ListCredentials(r CredentialsRetriever, w io.Writer) error {
	creds, err := r.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("unable to retrieve credentials from %q: %w", r.Source(), err)
	}

	if len(creds) > 0 {
		fmt.Fprintf(w, "NAME\tINFRASTRUCTURE PROVIDER\n")

		for _, c := range creds {
			fmt.Fprintf(w, "%s", c.Name)
			// Extract the infra provider name from ClusterKind
			provider := c.Kind[:strings.Index(c.Kind, "Cluster")]
			fmt.Fprintf(w, "\t%s", provider)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No credentials found.")

	return nil
}
