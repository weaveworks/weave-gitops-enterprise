package profiles

import (
	"context"
	"fmt"
	"io"
	"strings"

	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/controller"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ProfilesRetriever interface {
	Source() string
	RetrieveProfiles() (*pb.ListChartsForRepositoryResponse, error)
}

type GetOptions struct {
	Name      string
	Version   string
	Cluster   string
	Namespace string
	Writer    io.Writer
	Port      string
}

func (s *ProfilesSvc) Get(ctx context.Context, r ProfilesRetriever, w io.Writer) error {
	profiles, err := r.RetrieveProfiles()
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
func (s *ProfilesSvc) GetProfile(ctx context.Context, r ProfilesRetriever, opts GetOptions) (*pb.RepositoryChart, string, error) {
	s.Logger.Actionf("getting available profiles from %s", r.Source())

	profilesList, err := r.RetrieveProfiles()
	if err != nil {
		return nil, "", fmt.Errorf("unable to retrieve profiles from %q: %w", r.Source(), err)
	}

	var version string

	for _, p := range profilesList.Charts {
		if p.Name == opts.Name {
			if len(p.Versions) == 0 {
				return nil, "", fmt.Errorf("no version found for profile '%s' in %s/%s", p.Name, opts.Cluster, opts.Namespace)
			}

			switch {
			case opts.Version == "latest":
				versions, err := controller.ConvertStringListToSemanticVersionList(p.Versions)
				if err != nil {
					return nil, "", err
				}

				controller.SortVersions(versions)
				version = versions[0].String()
			default:
				if !foundVersion(p.Versions, opts.Version) {
					return nil, "", fmt.Errorf("version '%s' not found for profile '%s' in %s/%s", opts.Version, opts.Name, opts.Cluster, opts.Namespace)
				}

				version = opts.Version
			}

			return p, version, nil
		}
	}

	return nil, "", fmt.Errorf("no available profile '%s' found in %s/%s", opts.Name, opts.Cluster, opts.Namespace)
}

func foundVersion(availableVersions []string, version string) bool {
	for _, v := range availableVersions {
		if v == version {
			return true
		}
	}

	return false
}

func printProfiles(profiles *pb.ListChartsForRepositoryResponse, w io.Writer) {
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tAVAILABLE_VERSIONS\n")

	if profiles.Charts != nil && len(profiles.Charts) > 0 {
		for _, p := range profiles.Charts {
			fmt.Fprintf(w, "%s\t%s\t%v", p.Name, "", strings.Join(p.Versions, ","))
			fmt.Fprintln(w, "")
		}
	}
}
