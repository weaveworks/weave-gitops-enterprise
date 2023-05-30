package git

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/go-git-providers/stash"
	"github.com/go-logr/logr"
)

const BitBucketServerProviderName string = "bitbucket-server"

// BitBucketServerProvider is used to interact with the BitBucket Server (stash) API.
type BitBucketServerProvider struct {
	log    logr.Logger
	client gitprovider.Client
}

func NewBitBucketServerProvider(log logr.Logger) (Provider, error) {
	return &BitBucketServerProvider{
		log: log,
	}, nil
}

func (p *BitBucketServerProvider) Setup(opts ProviderOption) error {
	if opts.Username == "" {
		opts.Username = "git"
	}

	if opts.Token == "" {
		return fmt.Errorf("missing required option: Token")
	}

	if opts.Hostname == "" {
		return fmt.Errorf("missing required option: Hostname")
	}

	ggpOpts := []gitprovider.ClientOption{
		gitprovider.WithDomain(opts.Hostname),
		gitprovider.WithConditionalRequests(opts.ConditionalRequests),
		gitprovider.WithLogger(&p.log),
	}

	var err error

	p.client, err = stash.NewStashClient(opts.Username, opts.Token, ggpOpts...)

	return err
}

func (p *BitBucketServerProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	ggp := goGitProvider{}

	repo, err := ggp.GetBitbucketRepository(ctx, p.log, p.client, url)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Org:  repo.Repository().GetIdentity(),
		Name: repo.Repository().GetRepository(),
	}, nil
}

func (p *BitBucketServerProvider) CreatePullRequest(ctx context.Context, input PullRequestInput) (*PullRequest, error) {
	url, err := GetGitProviderUrl(input.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	ggp := goGitProvider{}

	repo, err := ggp.GetBitbucketRepository(ctx, p.log, p.client, url)
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

func (p *BitBucketServerProvider) GetTreeList(ctx context.Context, repoUrl string, sha string, path string) ([]*TreeEntry, error) {
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

func (p *BitBucketServerProvider) ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error) {
	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, repoURL)
	if err != nil {
		return nil, err
	}

	return ggp.ListPullRequests(ctx, repo)
}
