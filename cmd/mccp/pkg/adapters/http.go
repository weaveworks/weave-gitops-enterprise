package adapters

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	capiv1_protos "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// An HTTP client of the cluster service.
type HttpClient struct {
	baseURI *url.URL
	client  *resty.Client
}

// NewHttpClient creates a new HTTP client of the cluster service. The endpoint
// is expected to be an absolute HTTP URI.
func NewHttpClient(endpoint string, client *resty.Client) (*HttpClient, error) {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}

	client = client.SetHostURL(u.String())
	return &HttpClient{
		baseURI: u,
		client:  client,
	}, nil
}

// Source returns the endpoint of the cluster service
func (c *HttpClient) Source() string {
	return c.baseURI.String()
}

// RetrieveTemplates returns the list of all templates of the cluster service.
func (c *HttpClient) RetrieveTemplates() ([]templates.Template, error) {
	endpoint := "v1/templates"

	var templateList capiv1_protos.ListTemplatesResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
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
			Description: t.Description,
		})
	}

	return ts, nil
}

// RetrieveTemplateParameters returns the list of all parameters of the
// specified template.
func (c *HttpClient) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	endpoint := "v1/templates/{name}/params"

	var templateParameterList capiv1_protos.ListTemplateParamsResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetResult(&templateParameterList).
		Get(endpoint)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("unable to GET template parameters from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var tps []templates.TemplateParameter
	for _, p := range templateParameterList.Parameters {
		tps = append(tps, templates.TemplateParameter{
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return tps, nil
}

// RenderTemplateWithParameters returns a YAML representation of the specified
// template populated with the supplied parameters.
func (c *HttpClient) RenderTemplateWithParameters(name string, parameters map[string]string) (string, error) {
	endpoint := "v1/templates/{name}/render"

	// POST request payload
	type TemplateParameterValues struct {
		Values map[string]string `json:"values"`
	}

	// POST response payload
	type RenderedTemplate struct {
		Template string `json:"renderedTemplate"`
	}

	var renderedTemplate RenderedTemplate
	var serviceErr *ServiceError
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetBody(TemplateParameterValues{Values: parameters}).
		SetResult(&renderedTemplate).
		SetError(&serviceErr).
		Post(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to POST parameters and render template from %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to POST parameters and render template from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	return renderedTemplate.Template, nil
}

// CreatePullRequestForTemplate commits the YAML template to the specified
// branch and creates a pull request of that branch.
func (c *HttpClient) CreatePullRequestForTemplate(params templates.CreatePullRequestForTemplateParams) (string, error) {
	endpoint := "v1/pulls"

	// POST request payload
	type CreatePullRequestForTemplateRequest struct {
		RepositoryURL   string            `json:"repositoryUrl"`
		HeadBranch      string            `json:"headBranch"`
		BaseBranch      string            `json:"baseBranch"`
		Title           string            `json:"title"`
		Description     string            `json:"description"`
		TemplateName    string            `json:"templateName"`
		ParameterValues map[string]string `json:"parameter_values"`
		CommitMessage   string            `json:"commitMessage"`
	}

	// POST response payload
	type CreatePullRequestForTemplateResponse struct {
		WebURL string `json:"webUrl"`
	}

	var result CreatePullRequestForTemplateResponse
	var serviceErr *ServiceError
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetBody(CreatePullRequestForTemplateRequest{
			RepositoryURL:   params.RepositoryURL,
			HeadBranch:      params.HeadBranch,
			BaseBranch:      params.BaseBranch,
			Title:           params.Title,
			Description:     params.Description,
			TemplateName:    params.TemplateName,
			ParameterValues: params.ParameterValues,
			CommitMessage:   params.CommitMessage,
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
