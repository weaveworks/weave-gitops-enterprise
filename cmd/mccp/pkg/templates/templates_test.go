package templates_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

func TestListTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No templates were found in \"In-memory fake\"\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name: "template-a",
				},
				{
					Name: "template-b",
				},
			},
			expected: "Retrieved templates from \"In-memory fake\".\nName: template-a\nName: template-b\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:                   "template-a",
					Description:            "a desc",
					Version:                "1.0.1",
					InfrastructureProvider: "Amazon EKS",
					Author:                 "Nigel Potter",
				},
				{
					Name:                   "template-b",
					Description:            "b desc",
					Version:                "1.0.2",
					InfrastructureProvider: "Azure AKS",
					Author:                 "Nigel Potter",
				},
			},
			expected: "Retrieved templates from \"In-memory fake\".\nName: template-a\nDescription: a desc\nInfrastructure Provider: Amazon EKS\nVersion: 1.0.1\nAuthor: Nigel Potter\nName: template-b\nDescription: b desc\nInfrastructure Provider: Azure AKS\nVersion: 1.0.2\nAuthor: Nigel Potter\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFakeTemplateRetriever(tt.ts, tt.err)
			w := new(bytes.Buffer)
			err := templates.ListTemplates(r, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeTemplateRetriever struct {
	ts  []templates.Template
	err error
}

func NewFakeTemplateRetriever(ts []templates.Template, err error) FakeTemplateRetriever {
	return FakeTemplateRetriever{
		ts:  ts,
		err: err,
	}
}

func (r FakeTemplateRetriever) Source() string {
	return "In-memory fake"
}

func (r FakeTemplateRetriever) Retrieve() ([]templates.Template, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.ts, nil
}
