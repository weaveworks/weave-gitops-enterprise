package adapters

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

type HttpTemplateRetriever struct {
	endpoint string
	client   *resty.Client
}

func NewHttpTemplateRetriever(endpoint string, client *resty.Client) *HttpTemplateRetriever {
	return &HttpTemplateRetriever{
		endpoint: endpoint,
		client:   client,
	}
}

func (c *HttpTemplateRetriever) Retrieve() ([]templates.Template, error) {
	var templateList TemplateList
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(templateList).
		Get(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to GET templates from %q: %w", c.endpoint, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", c.endpoint, res.StatusCode())
	}

	var ts []templates.Template
	for _, t := range templateList.Templates {
		ts = append(ts, templates.Template{
			Name:                   t.Name,
			Description:            t.Description,
			Version:                t.Version,
			InfrastructureProvider: t.InfrastructureProvider,
			Author:                 t.Author,
		})
	}

	return ts, nil
}

func (c *HttpTemplateRetriever) Source() string {
	return c.endpoint
}

type TemplateList struct {
	Templates []Template `json:"templates"`
}

type Template struct {
	Name                   string `json:"name"`
	Description            string `json:"description"`
	Version                string `json:"version"`
	InfrastructureProvider string `json:"infrastructureProvider"`
	Author                 string `json:"author"`
}
