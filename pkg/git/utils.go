package git

import (
	"path"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/wait"
)

var DefaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 20 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

func GetGitProviderUrl(giturl string) (string, error) {
	repositoryAPIURL := viper.GetString("capi-templates-repository-api-url")
	if repositoryAPIURL != "" {
		return repositoryAPIURL, nil
	}

	ep, err := transport.NewEndpoint(giturl)
	if err != nil {
		return "", err
	}
	if ep.Protocol == "http" || ep.Protocol == "https" {
		return giturl, nil
	}

	httpsEp := transport.Endpoint{Protocol: "https", Host: ep.Host, Path: ep.Path}

	return httpsEp.String(), nil
}

// WithCombinedSubOrgs combines the subgroups into the organization field of the reference
// This is to work around a bug in the go-git-providers library where it doesn't handle subgroups correctly.
// https://github.com/fluxcd/go-git-providers/issues/183
func WithCombinedSubOrgs(ref gitprovider.OrgRepositoryRef) *gitprovider.OrgRepositoryRef {
	orgsWithSubGroups := append([]string{ref.Organization}, ref.SubOrganizations...)
	ref.Organization = path.Join(orgsWithSubGroups...)
	ref.SubOrganizations = nil
	return &ref
}

func addSchemeToDomain(domain string) string {
	// Fixing https:// again (ggp quirk)
	if domain != "github.com" && domain != "gitlab.com" && !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		return "https://" + domain
	}
	return domain
}

type writeFilesToBranchRequest struct {
	HeadBranch   string
	BaseBranch   string
	Commits      []Commit
	CreateBranch bool
}

type createPullRequestRequest struct {
	HeadBranch  string
	BaseBranch  string
	Title       string
	Description string
}

type createPullRequestResponse struct {
	WebURL string
}
