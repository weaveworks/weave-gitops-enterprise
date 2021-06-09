package adapters

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

type HttpTemplateRetriever struct {
	baseURI *url.URL
	client  *resty.Client
}

func NewHttpTemplateRetriever(endpoint string, client *resty.Client) (*HttpTemplateRetriever, error) {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}

	client = client.SetHostURL(u.String())
	return &HttpTemplateRetriever{
		baseURI: u,
		client:  client,
	}, nil
}

func (c *HttpTemplateRetriever) RetrieveTemplates() ([]templates.Template, error) {
	endpoint := "v1/templates"

	var templateList TemplateList
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

func (c *HttpTemplateRetriever) Source() string {
	return c.baseURI.String()
}

func (c *HttpTemplateRetriever) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	endpoint := "v1/templates/{name}/params"

	var templateParameterList TemplateParameterList
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
	for _, p := range templateParameterList.TemplateParameters {
		tps = append(tps, templates.TemplateParameter{
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return tps, nil
}

func (c *HttpTemplateRetriever) RenderTemplateWithParameters(name string, parameters map[string]string) (string, error) {
	endpoint := "v1/templates/{name}/render"

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

type TemplateList struct {
	Templates []Template `json:"templates"`
}

type Template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TemplateParameterList struct {
	TemplateParameters []TemplateParameter `json:"parameters"`
}

type TemplateParameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TemplateParameterValues struct {
	Values map[string]string `json:"values"`
}

type RenderedTemplate struct {
	Template string `json:"renderedTemplate"`
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
