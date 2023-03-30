package gitproviders

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type RepositoryURLProtocol string

const (
	RepositoryURLProtocolHTTPS RepositoryURLProtocol = "https"
	RepositoryURLProtocolSSH   RepositoryURLProtocol = "ssh"
	// AzureDevOpsHTTPDefaultDomain is used for HTTP clone URLs
	AzureDevOpsHTTPDefaultDomain = "dev.azure.com"
	// AzureDevOpsSSHDefaultDomain is used for SSH clone URLs
	AzureDevOpsSSHDefaultDomain = "ssh.dev.azure.com"
)

type RepoURL struct {
	repoName   string
	owner      string
	url        *url.URL
	normalized string
	provider   GitProviderName
	protocol   RepositoryURLProtocol
}

func NewRepoURL(uri string) (RepoURL, error) {
	providerName, err := detectGitProviderFromURL(uri, ViperGetStringMapString("git-host-types"))
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not get provider name from URL %s: %w", uri, err)
	}

	normalized, err := normalizeRepoURLString(uri)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not normalize repo URL %s: %w", uri, err)
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not create normalized repo URL %s: %w", uri, err)
	}

	owner, err := getOwnerFromURL(*u, providerName)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not get owner name from URL %s: %w", uri, err)
	}

	protocol := RepositoryURLProtocolSSH
	if u.Scheme == "https" {
		protocol = RepositoryURLProtocolHTTPS
	}

	return RepoURL{
		repoName:   utils.URLToRepoName(uri),
		owner:      owner,
		url:        u,
		normalized: normalized,
		provider:   providerName,
		protocol:   protocol,
	}, nil
}

func (n RepoURL) String() string {
	return n.normalized
}

func (n RepoURL) URL() *url.URL {
	return n.url
}

func (n RepoURL) Owner() string {
	return n.owner
}

func (n RepoURL) RepositoryName() string {
	return n.repoName
}

func (n RepoURL) Provider() GitProviderName {
	return n.provider
}

func (n RepoURL) Protocol() RepositoryURLProtocol {
	return n.protocol
}

func getOwnerFromURL(url url.URL, providerName GitProviderName) (string, error) {
	url.Path = strings.TrimPrefix(url.Path, "/")
	parts := strings.Split(url.Path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %v", url.String())
	}

	// Examples of Azure DevOps URLs:
	// 	- https://weaveworks@dev.azure.com/weaveworks/weave-gitops-integration/_git/config
	// 	- git@ssh.dev.azure.com:v3/weaveworks/weave-gitops-integration/config
	if providerName == GitProviderAzureDevOps {
		if parts[len(parts)-2] == "_git" {
			return strings.Join(parts[:2], "/"), nil
		} else {
			return strings.Join(parts[1:3], "/"), nil
		}
	}

	return strings.Join(parts[:len(parts)-1], "/"), nil
}

func GitHostTypes(gitHostTypesConfig map[string]string) map[string]string {
	defaults := map[string]string{
		github.DefaultDomain:         string(GitProviderGitHub),
		gitlab.DefaultDomain:         string(GitProviderGitLab),
		AzureDevOpsHTTPDefaultDomain: string(GitProviderAzureDevOps),
		AzureDevOpsSSHDefaultDomain:  string(GitProviderAzureDevOps),
	}

	// add in the user defined git host types
	for k, v := range gitHostTypesConfig {
		defaults[k] = v
	}

	return defaults
}

// detectGitProviderFromURL accepts a url related to a git repo and
// returns the name of the provider associated.
func detectGitProviderFromURL(raw string, gitHostTypesConfig map[string]string) (GitProviderName, error) {
	u, err := parseGitURL(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse git repo url %q: %w", raw, err)
	}

	gitHostTypes := GitHostTypes(gitHostTypesConfig)

	provider := gitHostTypes[u.Host]
	if provider == "" {
		return "", fmt.Errorf("no git providers found for %q", raw)
	}

	return GitProviderName(provider), nil
}

// Hacks around "scp" formatted urls ($user@$host:$path)
// the `:` delimiter between host and path throws off the std. url parser
func parseGitURL(raw string) (*url.URL, error) {
	if strings.HasPrefix(raw, "git@") {
		// The first occurrence of `:` should be the host:path delimiter.
		raw = strings.Replace(raw, ":", "/", 1)
		raw = "ssh://" + raw
	}

	return url.Parse(raw)
}

// normalizeRepoURLString accepts a url like git@github.com:someuser/podinfo.git and converts it into
// a string like ssh://git@github.com/someuser/podinfo.git. This helps standardize the different
// user inputs that might be provided.
func normalizeRepoURLString(url string) (string, error) {
	// https://github.com/weaveworks/weave-gitops/issues/878
	// A trailing slash causes problems when naming secrets.
	url = strings.TrimSuffix(url, "/")

	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	u, err := parseGitURL(url)
	if err != nil {
		return "", fmt.Errorf("could not parse git repo url while normalizing %q: %w", url, err)
	}

	return fmt.Sprintf("ssh://git@%s%s", u.Host, u.Path), nil
}

// ViperGetStringMapString looks up a command line flag or env var in the format "foo=1,bar=2"
// GetStringMapString tries to JSON decode the env var
// If that fails (silently), try and decode the classic "foo=1,bar=2" form.
// https://github.com/spf13/viper/issues/911
func ViperGetStringMapString(key string) map[string]string {
	sms := viper.GetStringMapString(key)
	if len(sms) > 0 {
		return sms
	}

	ss := viper.GetStringSlice(key)
	out := map[string]string{}

	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}

	return out
}
