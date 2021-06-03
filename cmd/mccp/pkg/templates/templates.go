package templates

import (
	"fmt"
	"io"
)

type TemplateRetriever interface {
	Source() string
	Retrieve() ([]Template, error)
}

type Template struct {
	Name                   string
	Description            string
	Version                string
	InfrastructureProvider string
	Author                 string
}

func ListTemplates(r TemplateRetriever, w io.Writer) error {
	ts, err := r.Retrieve()
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "Retrieved templates from %q.\n", r.Source())

		for _, t := range ts {
			fmt.Fprintf(w, "Name: %s\n", t.Name)
			if t.Description != "" {
				fmt.Fprintf(w, "Description: %s\n", t.Description)
			}
			if t.InfrastructureProvider != "" {
				fmt.Fprintf(w, "Infrastructure Provider: %s\n", t.InfrastructureProvider)
			}
			if t.Version != "" {
				fmt.Fprintf(w, "Version: %s\n", t.Version)
			}
			if t.Author != "" {
				fmt.Fprintf(w, "Author: %s\n", t.Author)
			}
		}

		return nil
	}

	fmt.Fprintf(w, "No templates were found in %q\n", r.Source())

	return nil
}
