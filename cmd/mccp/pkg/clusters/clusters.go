package clusters

import (
	"fmt"
	"io"
)

type ClustersRetriever interface {
	Source() string
	RetrieveClusters() ([]Cluster, error)
}

type Cluster struct {
	Name   string
	Status string
}

func ListClusters(r ClustersRetriever, w io.Writer) error {
	cs, err := r.RetrieveClusters()
	if err != nil {
		return fmt.Errorf("unable to retrieve clusters from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tSTATUS\n")
		for _, c := range cs {
			fmt.Fprintf(w, "%s\t%s", c.Name, c.Status)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No clusters found.\n")

	return nil
}
