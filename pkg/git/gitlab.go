package git

import (
	"context"
	"fmt"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
)

const (
	GitLabProviderName       string = "gitlab"
	deleteFilesCommitMessage string = "Delete old files for resources"
)

// GitLabProvider is used to interact with the GitLab API.
type GitLabProvider struct {
	log    logr.Logger
	client gitprovider.Client
}

func NewGitLabProvider(log logr.Logger) (Provider, error) {
	return &GitLabProvider{
		log: log,
	}, nil
}

func (p *GitLabProvider) Setup(opts ProviderOption) error {
	if opts.Token == "" {
		return fmt.Errorf("missing required option: Token")
	}

	ggpOpts := []gitprovider.ClientOption{
		gitprovider.WithConditionalRequests(opts.ConditionalRequests),
	}

	if opts.Hostname != "" {
		ggpOpts = append(ggpOpts, gitprovider.WithDomain(opts.Hostname))
	}

	var err error

	p.client, err = gitlab.NewClient(opts.Token, opts.TokenType, ggpOpts...)

	return err
}

func (p *GitLabProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
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

func (p *GitLabProvider) CreatePullRequest(ctx context.Context, input PullRequestInput) (*PullRequest, error) {
	url, err := GetGitProviderUrl(input.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, url)
	if err != nil {
		return nil, err
	}

	files := []CommitFile{}
	for _, commit := range input.Commits {
		files = append(files, commit.Files...)
	}

	if err := ggp.CreateBranch(ctx, p.log, repo, input.Base, input.Head); err != nil {
		return nil, err
	}

	updatedFiles, err := ggp.GetUpdatedFiles(ctx, files, p.client, input.RepositoryURL, input.Head)
	if err != nil {
		return nil, err
	}

	commits := []Commit{}

	if len(updatedFiles) > 0 {
		for idx := range updatedFiles {
			updatedFiles[idx].Content = nil
		}

		commits = append(commits, Commit{
			CommitMessage: deleteFilesCommitMessage,
			Files:         updatedFiles,
		})
	}

	commits = append(commits, input.Commits...)

	if err := ggp.WriteFilesToBranch(ctx, p.log, writeFilesToBranchRequest{
		HeadBranch:   input.Head,
		BaseBranch:   input.Base,
		Commits:      commits,
		CreateBranch: false,
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

func (p *GitLabProvider) GetTreeList(ctx context.Context, repoUrl string, sha string, path string) ([]*TreeEntry, error) {
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

func (p *GitLabProvider) ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error) {
	ggp := goGitProvider{}

	repo, err := ggp.GetRepository(ctx, p.log, p.client, repoURL)
	if err != nil {
		return nil, err
	}

	return ggp.ListPullRequests(ctx, repo)
}
