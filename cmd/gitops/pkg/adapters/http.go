package adapters

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"k8s.io/client-go/rest"

	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	kubecfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	expiredHeaderName          = "Entitlement-Expired-Message"
	gitProviderTokenHeaderName = "Git-Provider-Token"
	auth_cookie_name           = "id_token"
)

// An HTTP client of the cluster service.
type HTTPClient struct {
	baseURI    *url.URL
	client     *resty.Client
	authFunc   func(*config.Options, *HTTPClient) error
	configured bool
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewHTTPClient returns a new HTTP client for requests to Weave GitOps services.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: resty.New(),
	}
}

// EnableCLIAuth configures client to authenticate automatically with
// with either username or password, or kubeconfig, when a request is executed.
func (c *HTTPClient) EnableCLIAuth() *HTTPClient {
	c.authFunc = configureAuthForClient
	return c
}

// ConfigureClientWithOptions accepts a config.Options object that configures the client
// with the necessary options to make a request.
func (c *HTTPClient) ConfigureClientWithOptions(opts *config.Options, out io.Writer) error {
	u, err := url.ParseRequestURI(opts.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	c.baseURI = u

	c.client = c.client.SetHostURL(u.String()).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() >= http.StatusInternalServerError {
				fmt.Fprintf(out, "Server error: %s\n", r.Body())
				return nil
			}

			if m := r.Header().Get(expiredHeaderName); m != "" {
				fmt.Fprintln(out, m)
			}
			return nil
		})

	if opts.InsecureSkipTLSVerify {
		c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	if c.authFunc != nil {
		err = c.authFunc(opts, c)
		if err != nil {
			return fmt.Errorf("error: could not configure auth for client: %w", err)
		}
	}

	c.configured = true

	return nil
}

func configureAuthForClient(opts *config.Options, httpClient *HTTPClient) error {
	if opts.Username != "" && opts.Password != "" {
		err := httpClient.signIn(opts.Username, opts.Password)
		if err != nil {
			return err
		}

		return nil
	}

	// controller-runtime config getter does not allow us to pass a kubeconfig location
	// but does support the --kubeconfig flag via the `flag` stdlib package. Therefore
	// set this flag with the kubeconfig location if the user has passed one via the CLI.
	_ = flag.Set("kubeconfig", opts.Kubeconfig)

	restConfig, err := kubecfg.GetConfig()
	if err != nil {
		return fmt.Errorf("error: could not load config for kubeconfig: %w", err)
	}

	if opts.InsecureSkipTLSVerify {
		restConfig.TLSClientConfig = rest.TLSClientConfig{Insecure: true}
	}

	roundtripper, err := rest.TransportFor(restConfig)
	if err != nil {
		return err
	}

	httpClient.SetTransport(roundtripper)

	return nil
}

func getAuthCookie(cookies []*http.Cookie) (*http.Cookie, error) {
	for i := range cookies {
		if cookies[i].Name == auth_cookie_name {
			return cookies[i], nil
		}
	}

	return nil, errors.New("unable to find token in auth response")
}

func (c *HTTPClient) signIn(username, password string) error {
	endpoint := "oauth2/sign_in"

	type SignInBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	res, err := c.client.R().
		SetBody(SignInBody{Username: username, Password: password}).
		Post(endpoint)

	if err != nil {
		return fmt.Errorf("unable to sign in from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	cookie, err := getAuthCookie(res.Cookies())
	if err != nil {
		return err
	}

	c.client.SetCookie(cookie)

	return nil
}

// Source returns the endpoint of the cluster service.
func (c *HTTPClient) Source() string {
	return c.baseURI.String()
}

// GetClient returns the internal *resty.Client object.
func (c *HTTPClient) GetClient() *resty.Client {
	return c.client
}

// GetBaseClient returns the underlying http.Client object.
func (c *HTTPClient) GetBaseClient() *http.Client {
	return c.client.GetClient()
}

// SetTransport method sets custom `*http.Transport` or any `http.RoundTripper`
// compatible interface implementation in the client.
func (c *HTTPClient) SetTransport(transport http.RoundTripper) {
	c.client.SetTransport(transport)
}

// SetTLSClientConfig method sets TLSClientConfig for underling client Transport.
func (c *HTTPClient) SetTLSClientConfig(config *tls.Config) {
	c.client.SetTLSClientConfig(config)
}

// RetrieveTemplates returns the list of all templates from the cluster service.
func (c *HTTPClient) RetrieveTemplates(kind templates.TemplateKind) ([]templates.Template, error) {
	endpoint := "v1/templates"

	type ListTemplatesResponse struct {
		Templates []*templates.Template
	}

	var templateList ListTemplatesResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"template_kind": kind.String(),
		}).
		SetResult(&templateList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET templates from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var ts []templates.Template
	for _, t := range templateList.Templates {
		ts = append(ts, templates.Template{
			Name:        t.Name,
			Provider:    t.Provider,
			Description: t.Description,
			Error:       t.Error,
		})
	}

	return ts, nil
}

// RetrieveTemplate returns a template from the cluster service.
func (c *HTTPClient) RetrieveTemplate(name string, kind templates.TemplateKind, namespace string) (*templates.Template, error) {
	endpoint := "v1/templates/{template_name}"

	type GetTemplateResponse struct {
		Template *templates.Template
	}

	var template GetTemplateResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"template_name": name,
		}).
		SetQueryParams(map[string]string{
			"template_kind":      kind.String(),
			"template_namespace": namespace,
		}).
		SetResult(&template).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET template from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	return template.Template, nil
}

// RetrieveTemplatesByProvider returns the list of all templates for a given
// provider from the cluster service.
func (c *HTTPClient) RetrieveTemplatesByProvider(kind templates.TemplateKind, provider string) ([]templates.Template, error) {
	endpoint := "v1/templates"

	type ListTemplatesResponse struct {
		Templates []*templates.Template
	}

	var templateList ListTemplatesResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"provider":      provider,
			"template_kind": kind.String(),
		}).
		SetResult(&templateList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET templates from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var ts []templates.Template
	for _, t := range templateList.Templates {
		ts = append(ts, templates.Template{
			Name:        t.Name,
			Provider:    t.Provider,
			Description: t.Description,
		})
	}

	return ts, nil
}

// RetrieveTemplateParameters returns the list of all parameters of the
// specified template.
func (c *HTTPClient) RetrieveTemplateParameters(kind templates.TemplateKind, name string, namespace string) ([]templates.TemplateParameter, error) {
	endpoint := "v1/templates/{template_name}/params"

	type ListTemplateParametersResponse struct {
		Parameters []*templates.TemplateParameter
	}

	var templateParametersList ListTemplateParametersResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"template_name": name,
		}).
		SetQueryParams(map[string]string{
			"template_kind":      kind.String(),
			"template_namespace": namespace,
		}).
		SetResult(&templateParametersList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET template parameters from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var tps []templates.TemplateParameter
	for _, p := range templateParametersList.Parameters {
		tps = append(tps, templates.TemplateParameter{
			Name:        p.Name,
			Description: p.Description,
			Required:    p.Required,
			Options:     p.Options,
		})
	}

	return tps, nil
}

type TemplateParameterValuesAndCredentials struct {
	Values      map[string]string     `json:"values"`
	Credentials templates.Credentials `json:"credentials"`
}

// RenderTemplateWithParameters returns a YAML representation of the specified
// template populated with the supplied parameters.
func (c *HTTPClient) RenderTemplateWithParameters(req templates.RenderTemplateRequest) (*templates.RenderTemplateResponse, error) {

	endpoint := "v1/templates/{name}/render"

	resp := &templates.RenderTemplateResponse{}

	var serviceErr *ServiceError

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": req.TemplateName,
		}).
		SetQueryParams(map[string]string{
			"template_kind": req.TemplateKind.String(),
		}).
		SetBody(req).
		SetResult(resp).
		SetError(&serviceErr).
		Post(endpoint)

	if serviceErr != nil {
		return nil, fmt.Errorf("unable to POST parameters and render template from %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to POST parameters and render template from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	return resp, nil
}

// CreatePullRequestFromTemplate commits the YAML template to the specified
// branch and creates a pull request of that branch.
func (c *HTTPClient) CreatePullRequestFromTemplate(params templates.CreatePullRequestFromTemplateParams) (string, error) {
	// POST request payload
	type CreatePullRequestFromTemplateRequest struct {
		RepositoryURL     string                    `json:"repositoryUrl"`
		HeadBranch        string                    `json:"headBranch"`
		BaseBranch        string                    `json:"baseBranch"`
		Title             string                    `json:"title"`
		Description       string                    `json:"description"`
		TemplateName      string                    `json:"templateName"`
		TemplateNamespace string                    `json:"templateNamespace"`
		ParameterValues   map[string]string         `json:"parameter_values"`
		CommitMessage     string                    `json:"commitMessage"`
		Credentials       templates.Credentials     `json:"credentials"`
		ProfileValues     []templates.ProfileValues `json:"values"`
	}

	// POST response payload
	type CreatePullRequestFromTemplateResponse struct {
		WebURL string `json:"webUrl"`
	}

	var (
		endpoint   string
		result     CreatePullRequestFromTemplateResponse
		serviceErr *ServiceError
	)

	endpoint = "v1/clusters"
	if params.TemplateKind == templates.GitOpsTemplateKind.String() {
		endpoint = "v1/tfcontrollers"
	}

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetHeader(gitProviderTokenHeaderName, params.GitProviderToken).
		SetBody(CreatePullRequestFromTemplateRequest{
			RepositoryURL:     params.RepositoryURL,
			HeadBranch:        params.HeadBranch,
			BaseBranch:        params.BaseBranch,
			Title:             params.Title,
			Description:       params.Description,
			TemplateName:      params.TemplateName,
			TemplateNamespace: params.TemplateNamespace,
			ParameterValues:   params.ParameterValues,
			CommitMessage:     params.CommitMessage,
			Credentials:       params.Credentials,
			ProfileValues:     params.ProfileValues,
		}).
		SetResult(&result).
		SetError(&serviceErr).
		Post(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to POST template and create pull request to %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to POST template and create pull request to %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	return result.WebURL, nil
}

// RetrieveCredentials returns a list of all CAPI credentials.
func (c *HTTPClient) RetrieveCredentials() ([]templates.Credentials, error) {
	endpoint := "v1/credentials"

	type ListCredentialsResponse struct {
		Credentials []*templates.Credentials
		Total       int32
	}

	var credentialsList ListCredentialsResponse

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&credentialsList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET credentials from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var creds []templates.Credentials
	for _, c := range credentialsList.Credentials {
		creds = append(creds, templates.Credentials{
			Group:     c.Group,
			Version:   c.Version,
			Kind:      c.Kind,
			Name:      c.Name,
			Namespace: c.Namespace,
		})
	}

	return creds, nil
}

// RetrieveCredentialsByName returns a specific set of CAPI credentials.
func (c *HTTPClient) RetrieveCredentialsByName(name string) (templates.Credentials, error) {
	var creds templates.Credentials

	credsList, err := c.RetrieveCredentials()
	if err != nil {
		return creds, fmt.Errorf("unable to retrieve credentials from %q: %w", c.Source(), err)
	}

	for _, c := range credsList {
		if c.Name == name {
			creds = templates.Credentials{
				Group:     c.Group,
				Version:   c.Version,
				Kind:      c.Kind,
				Name:      c.Name,
				Namespace: c.Namespace,
			}
		}
	}

	return creds, nil
}

// RetrieveClusters returns the list of all clusters from the cluster service.
func (c *HTTPClient) RetrieveClusters() ([]clusters.Cluster, error) {
	endpoint := "/v1/clusters"

	type ClusterView struct {
		Name       string                `json:"name"`
		Conditions []*clusters.Condition `json:"conditions"`
	}

	type ClustersResponse struct {
		Clusters []ClusterView `json:"gitopsClusters"`
	}

	var clustersResponse ClustersResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&clustersResponse).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET clusters from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var cs []clusters.Cluster

	for _, c := range clustersResponse.Clusters {
		var conditions []clusters.Condition
		for _, condition := range c.Conditions {
			conditions = append(conditions, clusters.Condition{
				Type:    condition.Type,
				Status:  condition.Status,
				Message: condition.Message,
			})
		}

		cs = append(cs, clusters.Cluster{
			Name:       c.Name,
			Conditions: conditions,
		})
	}

	return cs, nil
}

func (c *HTTPClient) GetClusterKubeconfig(name string) (string, error) {
	endpoint := "v1/clusters/{name}/kubeconfig"

	type GetKubeconfigResponse struct {
		Kubeconfig string
	}

	var result GetKubeconfigResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetResult(&result).
		Get(endpoint)

	if err != nil {
		return "", fmt.Errorf("unable to GET cluster kubeconfig from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	b, err := base64.StdEncoding.DecodeString(result.Kubeconfig)
	if err != nil {
		return "", fmt.Errorf("unable to base64 decode the cluster kubeconfig: %w", err)
	}

	return string(b), nil
}

func (c *HTTPClient) RetrieveProfiles() (*pb.GetProfilesResponse, error) {
	endpoint := "/v1/profiles"

	result := &pb.GetProfilesResponse{}

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(result).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET profiles from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	return result, nil
}

// DeleteClusters deletes CAPI cluster using its name
func (c *HTTPClient) DeleteClusters(params clusters.DeleteClustersParams) (string, error) {
	endpoint := "v1/clusters"

	type DeleteClustersPullRequestRequest struct {
		RepositoryUrl string                `json:"repositoryUrl"`
		HeadBranch    string                `json:"headBranch"`
		BaseBranch    string                `json:"baseBranch"`
		Title         string                `json:"title"`
		Description   string                `json:"description"`
		ClusterNames  []string              `json:"clusterNames"`
		CommitMessage string                `json:"commitMessage"`
		Credentials   templates.Credentials `json:"credentials"`
	}

	type DeleteClustersResponse struct {
		WebURL string `json:"webUrl"`
	}

	var result DeleteClustersResponse

	var serviceErr *ServiceError

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetHeader(gitProviderTokenHeaderName, params.GitProviderToken).
		SetBody(DeleteClustersPullRequestRequest{
			HeadBranch:    params.HeadBranch,
			BaseBranch:    params.BaseBranch,
			Title:         params.Title,
			Description:   params.Description,
			ClusterNames:  params.ClustersNames,
			CommitMessage: params.CommitMessage,
		}).
		SetResult(&result).
		SetError(&serviceErr).
		Delete(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to Delete cluster and create pull request to %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to Delete cluster and create pull request to %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for Delete %q was %d", res.Request.URL, res.StatusCode())
	}

	return result.WebURL, nil
}

// RetrieveTemplateProfiles returns the list of all profiles of the
// specified template.
func (c *HTTPClient) RetrieveTemplateProfiles(name string, namespace string) ([]templates.Profile, error) {
	endpoint := "v1/templates/{name}/profiles"

	type ListTemplatePResponse struct {
		Profiles []*templates.Profile
	}

	var templateProfilesList ListTemplatePResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetQueryParams(map[string]string{
			"template_namespace": namespace,
		}).
		SetResult(&templateProfilesList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET template profiles from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var tps []templates.Profile
	for _, p := range templateProfilesList.Profiles {
		tps = append(tps, templates.Profile{
			Name:              p.Name,
			Home:              p.Home,
			Sources:           p.Sources,
			Description:       p.Description,
			Maintainers:       p.Maintainers,
			Icon:              p.Icon,
			KubeVersion:       p.KubeVersion,
			HelmRepository:    p.HelmRepository,
			AvailableVersions: p.AvailableVersions,
		})
	}

	return tps, nil
}
