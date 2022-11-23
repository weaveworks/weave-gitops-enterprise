package profiles

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/controller"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ProfilesRetriever interface {
	Source() string
	RetrieveProfiles(GetOptions) ([]Profile, error)
}

type GetOptions struct {
	Name       string
	Version    string
	Cluster    string
	Namespace  string
	Writer     io.Writer
	Port       string
	Kind       string        `json:"kind,omitempty"`
	Repository RepositoryRef `json:"repository,omitempty"`
}

type RepositoryRef struct {
	Name      string     `json:"name,omitempty"`
	Namespace string     `json:"namespace,omitempty"`
	Kind      string     `json:"kind,omitempty"`
	Cluster   ClusterRef `json:"cluster,omitempty"`
}

type ClusterRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Profile struct {
	Name       string        `json:"name,omitempty"`
	Versions   []string      `json:"versions,omitempty"`
	Layer      string        `json:"layer,omitempty"`
	Repository RepositoryRef `json:"repository,omitempty"`
}

func (s *ProfilesSvc) Get(ctx context.Context, r ProfilesRetriever, w io.Writer, opts GetOptions) error {

	profiles, err := r.RetrieveProfiles(opts)
	if err != nil {
		if e, ok := err.(*errors.StatusError); ok {
			return fmt.Errorf("unable to retrieve profiles from %q: status code %d", r.Source(), e.ErrStatus.Code)
		}

		return fmt.Errorf("unable to retrieve profiles from %q: %w", r.Source(), err)
	}

	printProfiles(profiles, w)

	return nil
}

// GetProfile returns a single available profile.
func (s *ProfilesSvc) GetProfile(ctx context.Context, r ProfilesRetriever, opts GetOptions) (Profile, string, error) {
	s.Logger.Actionf("getting available profiles from %s", r.Source())

	profilesList, err := r.RetrieveProfiles(opts)
	if err != nil {
		return Profile{}, "", fmt.Errorf("unable to retrieve profiles from %q: %w", r.Source(), err)
	}

	var version string

	for _, p := range profilesList {
		if p.Name == opts.Name {
			if len(p.Versions) == 0 {
				return Profile{}, "", fmt.Errorf("no version found for profile '%s' in %s/%s", p.Name, opts.Cluster, opts.Namespace)
			}

			switch {
			case opts.Version == "latest":
				versions, err := controller.ConvertStringListToSemanticVersionList(p.Versions)
				if err != nil {
					return Profile{}, "", err
				}

				controller.SortVersions(versions)
				version = versions[0].String()
			default:
				if !foundVersion(p.Versions, opts.Version) {
					return Profile{}, "", fmt.Errorf("version '%s' not found for profile '%s' in %s/%s", opts.Version, opts.Name, opts.Cluster, opts.Namespace)
				}

				version = opts.Version
			}

			if p.Repository.Name == "" || p.Repository.Namespace == "" {
				return Profile{}, "", fmt.Errorf("HelmRepository's name or namespace is empty")
			}

			return p, version, nil
		}
	}

	return Profile{}, "", fmt.Errorf("no available profile '%s' found in %s/%s", opts.Name, opts.Cluster, opts.Namespace)
}

func foundVersion(availableVersions []string, version string) bool {
	for _, v := range availableVersions {
		if v == version {
			return true
		}
	}

	return false
}

func printProfiles(profiles []Profile, w io.Writer) {
	fmt.Fprintf(w, "NAME\tAVAILABLE_VERSIONS\tLAYER\n")

	if len(profiles) > 0 {
		for _, p := range profiles {
			fmt.Fprintf(w, "%s\t%s\t%v", p.Name, strings.Join(p.Versions, ","), p.Layer)
			fmt.Fprintln(w, "")
		}
	}
}
