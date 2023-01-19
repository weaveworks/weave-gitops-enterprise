package templates

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type TemplateKind string

const (
	// CAPITemplateKind defines a CAPI template
	CAPITemplateKind TemplateKind = "CAPITemplate"

	// GitOpsTemplateKind defines a TF-Controller template
	GitOpsTemplateKind TemplateKind = "GitOpsTemplate"

	// Default no template kind
	DefaultTemplateKind TemplateKind = ""
)

var TemplateKinds = []TemplateKind{
	CAPITemplateKind,
	GitOpsTemplateKind,
}

// Return a string representation of all supported template Kinds
func templateKindsString() string {
	var kinds []string
	for _, k := range TemplateKinds {
		kinds = append(kinds, k.String())
	}
	return strings.Join(kinds, ", ")
}

// String returns a string representation of the template Kind.
func (t TemplateKind) String() string {
	return string(t)
}

// Set the value of the template kind object with a string
func (t *TemplateKind) Set(v string) error {
	if v == "" {
		*t = DefaultTemplateKind
		return nil
	}
	if inTemplateKinds(v) {
		*t = TemplateKind(v)
		return nil
	}
	return fmt.Errorf("template kind not found, supported templates: %s", templateKindsString())

}

func inTemplateKinds(str string) bool {
	for _, k := range TemplateKinds {
		if k.String() == str {
			return true
		}
	}
	return false
}

type CreatePullRequestFromTemplateParams struct {
	GitProviderToken  string
	TemplateName      string
	TemplateNamespace string
	TemplateKind      string
	ParameterValues   map[string]string
	RepositoryURL     string
	HeadBranch        string
	BaseBranch        string
	Title             string
	Description       string
	CommitMessage     string
	Credentials       Credentials
	ProfileValues     []ProfileValues
}

// TemplatePullRequester defines the interface that adapters
// need to implement in order to create a pull request from
// a template (e.g. CAPI template, TF-Controller template).
// Implementers should return the web URI of the pull request.
type TemplatePullRequester interface {
	CreatePullRequestFromTemplate(params CreatePullRequestFromTemplateParams) (string, error)
}

// CreatePullRequestFromTemplate uses a TemplatePullRequester
// adapter to create a pull request from a template.
func CreatePullRequestFromTemplate(params CreatePullRequestFromTemplateParams, r TemplatePullRequester, w io.Writer) error {
	res, err := r.CreatePullRequestFromTemplate(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	fmt.Fprintf(w, "Created pull request: %s\n", res)

	return nil
}

// TemplatesRetriever defines the interface that adapters
// need to implement in order to return an array of templates.
type TemplatesRetriever interface {
	Source() string
	RetrieveTemplate(name string, kind TemplateKind, namespace string) (*Template, error)
	RetrieveTemplates(kind TemplateKind) ([]Template, error)
	RetrieveTemplatesByProvider(kind TemplateKind, provider string) ([]Template, error)
	RetrieveTemplateParameters(kind TemplateKind, name string, namespace string) ([]TemplateParameter, error)
	RetrieveTemplateProfiles(name string, namespace string) ([]Profile, error)
}

// TemplateRenderer defines the interface that adapters
// need to implement in order to render a template populated
// with parameter values.
type TemplateRenderer interface {
	RenderTemplateWithParameters(req RenderTemplateRequest) (*RenderTemplateResponse, error)
}

// CredentialsRetriever defines the interface that adapters
// need to implement in order to retrieve CAPI credentials.
type CredentialsRetriever interface {
	Source() string
	RetrieveCredentials() ([]Credentials, error)
}

type Template struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Provider     string `json:"provider"`
	TemplateKind string `json:"templateKind"`
	TemplateType string `json:"templateType"`
	Error        string `json:"error"`
}

type TemplateParameter struct {
	Name        string
	Description string
	Required    bool
	Options     []string
}

type Credentials struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ProfileValues struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Values    string `json:"values"`
	Namespace string `json:"namespace"`
}

type Profile struct {
	Name              string
	Home              string
	Sources           []string
	Description       string
	Keywords          []string
	Maintainers       []Maintainer
	Icon              string
	Annotations       map[string]string
	KubeVersion       string
	HelmRepository    HelmRepository
	AvailableVersions []string
}

type HelmRepository struct {
	Name      string
	Namespace string
}

type Maintainer struct {
	Name  string
	Email string
	Url   string
}

type RenderTemplateRequest struct {
	TemplateName      string            `json:"template_name,omitempty"`
	Values            map[string]string `json:"values,omitempty"`
	Credentials       Credentials       `json:"credentials,omitempty"`
	TemplateKind      TemplateKind      `json:"template_kind,omitempty"`
	ClusterNamespace  string            `json:"cluster_namespace,omitempty"`
	Profiles          []ProfileValues   `json:"profiles,omitempty"`
	TemplateNamespace string            `json:"template_namespace,omitempty"`
}

type CommitFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type RenderTemplateResponse struct {
	RenderedTemplate []CommitFile `json:"renderedTemplate"`
	ProfileFiles     []CommitFile `json:"profileFiles"`
}

func (r *RenderTemplateResponse) String() string {
	var output strings.Builder
	for i := range r.RenderedTemplate {
		file := r.RenderedTemplate[i]
		output.WriteString(
			fmt.Sprintf(
				"\n---\n# %s\n\n%s",
				file.Path,
				file.Content,
			),
		)
	}
	for i := range r.ProfileFiles {
		file := r.ProfileFiles[i]
		output.WriteString(
			fmt.Sprintf(
				"\n---\n# %s\n\n%s",
				file.Path,
				file.Content,
			),
		)
	}
	return output.String()
}

// GetTemplate uses a TemplatesRetriever adapter to show print template to the console.
func GetTemplate(name string, kind TemplateKind, namespace string, r TemplatesRetriever, w io.Writer) error {
	t, err := r.RetrieveTemplate(name, kind, namespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	fmt.Fprintf(w, "NAME\tPROVIDER\tDESCRIPTION\tERROR\n")

	fmt.Fprintf(w, "%s", t.Name)
	fmt.Fprintf(w, "\t%s", t.Provider)
	fmt.Fprintf(w, "\t%s", t.Description)
	fmt.Fprintf(w, "\t%s", t.Error)
	fmt.Fprintln(w, "")

	return nil
}

// GetTemplates uses a TemplatesRetriever adapter to show
// a list of templates to the console.
func GetTemplates(kind TemplateKind, r TemplatesRetriever, w io.Writer) error {
	allTemplates := []Template{}
	if kind == "" {
		// get all templates for all supported kinds
		for _, templateKind := range TemplateKinds {
			ts, err := r.RetrieveTemplates(templateKind)
			if err != nil {
				return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
			}
			allTemplates = append(allTemplates, ts...)

		}
	} else {
		// get templates for given kind
		ts, err := r.RetrieveTemplates(kind)
		if err != nil {
			return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
		}

		allTemplates = append(allTemplates, ts...)

	}
	if len(allTemplates) > 0 {
		fmt.Fprintf(w, "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\n")

		for _, t := range allTemplates {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%s", t.Provider)
			fmt.Fprintf(w, "\t%s", t.TemplateType)
			fmt.Fprintf(w, "\t%s", t.Description)
			fmt.Fprintf(w, "\t%s", t.Error)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates were found.\n")

	return nil
}

// GetTemplatesByProvider uses a TemplatesRetriever adapter to show
// a list of templates for a given provider to the console.
func GetTemplatesByProvider(kind TemplateKind, provider string, r TemplatesRetriever, w io.Writer) error {
	allTemplates := []Template{}
	if kind == "" {
		// get all templates for all supported kinds
		for _, templateKind := range TemplateKinds {
			ts, err := r.RetrieveTemplatesByProvider(templateKind, provider)
			if err != nil {
				return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
			}
			allTemplates = append(allTemplates, ts...)
		}
	} else {
		// get templates for given kind
		ts, err := r.RetrieveTemplatesByProvider(kind, provider)
		if err != nil {
			return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
		}
		allTemplates = append(allTemplates, ts...)
	}

	if len(allTemplates) > 0 {
		fmt.Fprintf(w, "NAME\tPROVIDER\tTYPE\tDESCRIPTION\tERROR\n")

		for _, t := range allTemplates {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%s", t.Provider)
			fmt.Fprintf(w, "\t%s", t.TemplateType)
			fmt.Fprintf(w, "\t%s", t.Description)
			fmt.Fprintf(w, "\t%s", t.Error)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates were found for provider %q.\n", provider)

	return nil
}

// GetTemplateParameters uses a TemplatesRetriever adapter
// to show a list of parameters for a given template.
func GetTemplateParameters(kind TemplateKind, name string, namespace string, r TemplatesRetriever, w io.Writer) error {
	allParameters := []TemplateParameter{}
	if kind == "" {
		for _, templateKind := range TemplateKinds {
			ps, err := r.RetrieveTemplateParameters(templateKind, name, namespace)
			if err != nil {
				continue
			}
			allParameters = append(allParameters, ps...)
		}
	} else {
		// get templates for given kind
		ps, err := r.RetrieveTemplateParameters(kind, name, namespace)
		if err != nil {
			return fmt.Errorf("unable to retrieve parameters for template %q from %q: %w", name, r.Source(), err)
		}
		allParameters = append(allParameters, ps...)
	}

	if len(allParameters) > 0 {
		fmt.Fprintf(w, "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\n")

		for _, t := range allParameters {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%t", t.Required)

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

	fmt.Fprintf(w, "No template parameters were found.\n")

	return nil
}

// RenderTemplate uses a TemplateRenderer adapter to show
// a template populated with parameter values.
func RenderTemplateWithParameters(req RenderTemplateRequest, r TemplateRenderer, w io.Writer) error {
	t, err := r.RenderTemplateWithParameters(req)
	if err != nil {
		return fmt.Errorf("unable to render template %q: %w", req.TemplateName, err)
	}

	if t != nil {
		fmt.Fprint(w, t.String())
		return nil
	}

	fmt.Fprintf(w, "No template was found.\n")

	return nil
}

// GetCredentials uses a CredentialsRetriever adapter to show
// a list of CAPI credentials.
func GetCredentials(r CredentialsRetriever, w io.Writer) error {
	cs, err := r.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("unable to retrieve credentials from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tINFRASTRUCTURE PROVIDER\n")

		for _, c := range cs {
			fmt.Fprintf(w, "%s", c.Name)
			// Extract the infra provider name from ClusterKind
			provider := c.Kind[:strings.Index(c.Kind, "Cluster")]
			fmt.Fprintf(w, "\t%s", provider)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No credentials were found.\n")

	return nil
}

// GetTemplateProfiles uses a TemplatesRetriever adapter
// to show a list of profiles for a given template.
func GetTemplateProfiles(name string, namespace string, r TemplatesRetriever, w io.Writer) error {
	ps, err := r.RetrieveTemplateProfiles(name, namespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve profiles for template %q from %q: %w", name, r.Source(), err)
	}

	if len(ps) > 0 {
		fmt.Fprintf(w, "NAME\tLATEST_VERSIONS\n")

		for _, p := range ps {
			if len(p.AvailableVersions) > 5 {
				p.AvailableVersions = p.AvailableVersions[len(p.AvailableVersions)-5:]
			}

			latestVersions := strings.Join(p.AvailableVersions, ", ")

			fmt.Fprintf(w, "%s", p.Name)
			fmt.Fprintf(w, "\t%s", latestVersions)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No template profiles were found.\n")

	return nil
}

func ParseProfileFlags(profiles []string) ([]ProfileValues, error) {
	var profilesValues []ProfileValues

	// Validate values include alphanumeric or - or .
	r := regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`)

	for _, p := range profiles {
		valuesPairs := strings.Split(p, ",")
		profileMap := make(map[string]string)

		for _, pair := range valuesPairs {
			kv := strings.Split(pair, "=")

			if kv[0] != "name" && kv[0] != "version" && kv[0] != "values" && kv[0] != "namespace" {
				return nil, fmt.Errorf("invalid key: %s", kv[0])
			} else if kv[0] == "values" {
				file, err := os.ReadFile(kv[1])
				if err == nil {
					profileMap[kv[0]] = base64.StdEncoding.EncodeToString(file)
				}
			} else if kv[0] == "version" {
				_, err := semver.NewConstraint(kv[1])
				if err != nil {
					return nil, fmt.Errorf("invalid semver for %s: %s, %w", kv[0], kv[1], err)
				}
				profileMap[kv[0]] = kv[1]
			} else if !r.MatchString(kv[1]) {
				return nil, fmt.Errorf("invalid value for %s: %s", kv[0], kv[1])
			} else {
				profileMap[kv[0]] = kv[1]
			}
		}

		if _, ok := profileMap["name"]; !ok {
			return nil, fmt.Errorf("profile name must be specified")
		}

		profileJson, err := json.Marshal(profileMap)
		if err != nil {
			return nil, err
		}

		var profileValues ProfileValues

		err = json.Unmarshal(profileJson, &profileValues)
		if err != nil {
			return nil, err
		}

		profilesValues = append(profilesValues, profileValues)
	}

	return profilesValues, nil
}
