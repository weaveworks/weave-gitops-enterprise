package git

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
)

const GitHubProviderName string = "github"

// GitHubProvider is used to interact with the GitHub API.
// This implementation delegates most of the work to the
// fluxcd/go-git-providers library.
type GitHubProvider struct {
	log    logr.Logger
	client gitprovider.Client
}

func NewGitHubProvider(log logr.Logger) (Provider, error) {
	return &GitHubProvider{
		log: log,
	}, nil
}

func (p *GitHubProvider) Setup(opts ProviderOption) error {
	if opts.OAuth2Token == "" {
		return fmt.Errorf("missing required option: OAuth2Token")
	}

	ggpOpts := []gitprovider.ClientOption{
		gitprovider.WithOAuth2Token(opts.OAuth2Token),
	}

	if opts.Hostname != "" {
		ggpOpts = append(ggpOpts, gitprovider.WithDomain(opts.Hostname))
	}

	var err error

	p.client, err = github.NewClient(ggpOpts...)

	return err
}

func (p *GitHubProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, url)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Org:  repo.Repository().GetIdentity(),
		Name: repo.Repository().GetRepository(),
	}, nil
}

func (p *GitHubProvider) CreatePullRequest(ctx context.Context, input PullRequestInput) (*PullRequest, error) {
	url, err := GetGitProviderUrl(input.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, url)
	if err != nil {
		return nil, err
	}

	if err := ggp.WriteFilesToBranch(ctx, p.log, writeFilesToBranchRequest{
		HeadBranch:   input.Head,
		BaseBranch:   input.Base,
		Commits:      input.Commits,
		CreateBranch: true,
	}, repo); err != nil {
		return nil, fmt.Errorf("unable to write files to branch %q: %w", input.Head, err)
	}

	res, err := ggp.CreatePullRequest(ctx, p.log, createPullRequestRequest{
		HeadBranch:  input.Head,
		BaseBranch:  input.Base,
		Title:       input.Title,
		Description: input.Body,
	}, repo)
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request for branch %q: %w", input.Head, err)
	}

	return &PullRequest{
		Link: res.WebURL,
	}, nil
}

func (p *GitHubProvider) GetTreeList(ctx context.Context, repoUrl string, sha string, path string) ([]*TreeEntry, error) {
	url, err := GetGitProviderUrl(repoUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, url)
	if err != nil {
		return nil, err
	}

	files := []*TreeEntry{}

	treePaths, err := repo.Trees().List(ctx, sha, path, true)
	if err != nil {
		return nil, err
	}

	for _, file := range treePaths {
		files = append(files, &TreeEntry{
			Path: file.Path,
			Type: file.Type,
			Size: file.Size,
			SHA:  file.SHA,
			Link: file.URL,
		})
	}

	return files, nil
}

func (p *GitHubProvider) ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error) {
	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, repoURL)
	if err != nil {
		return nil, err
	}

	return ggp.ListPullRequests(ctx, repo)
}
