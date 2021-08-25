package clusters

import (
	"fmt"
	"io"
)

type ClustersRetriever interface {
	Source() string
	RetrieveClusters() ([]Cluster, error)
	GetClusterKubeconfig(name string) (string, error)
	DeleteClusters(params DeleteClustersParams) (string, error)
}

type Cluster struct {
	Name            string
	Status          string
	PullRequestType string
}

func ListClusters(r ClustersRetriever, w io.Writer) error {
	cs, err := r.RetrieveClusters()
	if err != nil {
		return fmt.Errorf("unable to retrieve clusters from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tSTATUS\n")
		for _, c := range cs {
			if c.PullRequestType == "create" {
				c.Status = "Creation PR"
			} else if c.PullRequestType == "delete" {
				c.Status = "Deletion PR"
			}

			fmt.Fprintf(w, "%s\t%s", c.Name, c.Status)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No clusters found.\n")

	return nil
}

func GetClusterKubeconfig(name string, r ClustersRetriever, w io.Writer) error {
	k, err := r.GetClusterKubeconfig(name)
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster %q from %q: %w", name, r.Source(), err)
	}

	fmt.Fprint(w, k)
	return nil
}

func DeleteClusters(params DeleteClustersParams, r ClustersRetriever, w io.Writer) error {
	pr, err := r.DeleteClusters(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request for cluster deletion: %w", err)
	}

	fmt.Fprintf(w, "Created pull request for clusters deletion: %s\n", pr)
	return nil
}

type DeleteClustersParams struct {
	RepositoryURL string
	HeadBranch    string
	BaseBranch    string
	Title         string
	Description   string
	ClustersNames []string
	CommitMessage string
}
