package git

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"k8s.io/apimachinery/pkg/util/wait"
)

var DefaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 20 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

type Provider interface {
	WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req WriteFilesToBranchAndCreatePullRequestRequest) (*WriteFilesToBranchAndCreatePullRequestResponse, error)
	GetRepository(ctx context.Context, gp GitProvider, url string) (*git.Repository, error)
	GetTreeList(ctx context.Context, gp GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*git.TreeEntry, error)
	ListPullRequests(ctx context.Context, gp GitProvider, url string) ([]*git.PullRequest, error)
}

type GitProviderService struct {
	log logr.Logger
}

func NewGitProviderService(log logr.Logger) *GitProviderService {
	return &GitProviderService{
		log: log,
	}
}

type GitProvider struct {
	Token     string
	TokenType string
	Type      string
	Hostname  string
}

type WriteFilesToBranchAndCreatePullRequestRequest struct {
	GitProvider       GitProvider
	RepositoryURL     string
	ReposistoryAPIURL string
	HeadBranch        string
	BaseBranch        string
	Title             string
	Description       string
	CommitMessage     string
	Files             []git.CommitFile
}

type WriteFilesToBranchAndCreatePullRequestResponse struct {
	WebURL string
}

// WriteFilesToBranchAndCreatePullRequest writes a set of provided files
// to a new branch and creates a new pull request for that branch.
// It returns the URL of the pull request.
func (s *GitProviderService) WriteFilesToBranchAndCreatePullRequest(
	ctx context.Context,
	req WriteFilesToBranchAndCreatePullRequestRequest,
) (*WriteFilesToBranchAndCreatePullRequestResponse, error) {
	provider, err := getGitProviderClient(s.log, req.GitProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to create provider: %w", err)
	}

	pr, err := provider.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: req.RepositoryURL,
		Title:         req.Title,
		Body:          req.Description,
		Head:          req.HeadBranch,
		Base:          req.BaseBranch,
		Commits: []git.Commit{{
			CommitMessage: req.CommitMessage,
			Files:         req.Files,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request for branch %q: %w", req.HeadBranch, err)
	}

	return &WriteFilesToBranchAndCreatePullRequestResponse{
		WebURL: pr.Link,
	}, nil
}

type GitRepo struct {
	WorktreeDir string
	Repo        *go_git.Repository
	Auth        *http.BasicAuth
}

func (s *GitProviderService) GetRepository(ctx context.Context, gp GitProvider, url string) (*git.Repository, error) {
	provider, err := getGitProviderClient(s.log, gp)
	if err != nil {
		return nil, fmt.Errorf("unable to get a git provider client for %q: %w", gp.Type, err)
	}

	return provider.GetRepository(ctx, url)
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

// GetTreeList retrieves list of tree files from gitprovider given the sha/branch
func (s *GitProviderService) GetTreeList(ctx context.Context, gp GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*git.TreeEntry, error) {
	provider, err := getGitProviderClient(s.log, gp)
	if err != nil {
		return nil, fmt.Errorf("unable to get a git provider client for %q: %w", gp.Type, err)
	}

	return provider.GetTreeList(ctx, repoUrl, sha, path)
}

func (s *GitProviderService) ListPullRequests(ctx context.Context, gp GitProvider, repoURL string) ([]*git.PullRequest, error) {
	provider, err := getGitProviderClient(s.log, gp)
	if err != nil {
		return nil, fmt.Errorf("unable to get a git provider client for %q: %w", gp.Type, err)
	}

	return provider.ListPullRequests(ctx, repoURL)
}

type Commit struct {
	CommitMessage string
	Files         []gitprovider.CommitFile
}

func getGitProviderClient(log logr.Logger, gpi GitProvider) (git.Provider, error) {
	// quirk of ggp
	hostname := addSchemeToDomain(gpi.Hostname)

	providerFactory := git.NewFactory(log)
	providerOpts := []git.ProviderWithFn{}

	switch gpi.Type {
	case git.GitHubProviderName:
		providerOpts = append(providerOpts, git.WithOAuth2Token(gpi.Token))

		if gpi.Hostname != "github.com" {
			providerOpts = append(providerOpts, git.WithDomain(hostname))
		}
	case git.GitLabProviderName:
		providerOpts = append(providerOpts, git.WithConditionalRequests())
		providerOpts = append(providerOpts, git.WithToken(gpi.TokenType, gpi.Token))

		if gpi.Hostname != "gitlab.com" {
			providerOpts = append(providerOpts, git.WithDomain(hostname))
		}
	case git.BitBucketServerProviderName:
		providerOpts = append(providerOpts, git.WithUsername("git"))
		providerOpts = append(providerOpts, git.WithToken(gpi.TokenType, gpi.Token))
		providerOpts = append(providerOpts, git.WithDomain(hostname))
		providerOpts = append(providerOpts, git.WithConditionalRequests())
	case git.AzureGitOpsProviderName:
		providerOpts = append(providerOpts, git.WithToken(gpi.TokenType, gpi.Token))
	default:
		return nil, fmt.Errorf("the Git provider %q is not supported", gpi.Type)
	}

	provider, err := providerFactory.Create(
		gpi.Type,
		providerOpts...,
	)

	return provider, err
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

func addSchemeToDomain(domain string) string {
	// Fixing https:// again (ggp quirk)
	if domain != "github.com" && domain != "gitlab.com" && !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		return "https://" + domain
	}
	return domain
}
