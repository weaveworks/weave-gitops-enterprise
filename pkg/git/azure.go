package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/drone/go-scm/scm/driver/azure"
	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

const (
	AzureGitOpsProviderName string = "azure-gitops"
)

// AzureGitOpsProvider is used to interact with the AzureGitOps API.
type AzureGitOpsProvider struct {
	log    logr.Logger
	client *scm.Client
}

func NewAzureGitOpsProvider(log logr.Logger) (Provider, error) {
	return &AzureGitOpsProvider{
		log: log,
	}, nil
}

func (p *AzureGitOpsProvider) Setup(opts ProviderOption) error {
	if opts.Token == "" {
		return fmt.Errorf("missing required option: Token")
	}

	if opts.Hostname == "" {
		opts.Hostname = "dev.azure.com"
	}

	// ggpOpts := []gitprovider.ClientOption{
	// 	gitprovider.WithDomain(opts.Hostname),
	// }

	var err error

	p.client, err = factory.NewClient("azure", fmt.Sprintf("https://%s", opts.Hostname), opts.Token)

	return err
}

func (p *AzureGitOpsProvider) GetRepository(ctx context.Context, repoURL string) (*Repository, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unbale to parse url %q: %w", repoURL, err)
	}

	dscm := jenkinsSCM{}

	repo, err := dscm.GetRepository(ctx, p.log, p.client, u)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Domain: u.Host,
		Org:    repo.Namespace,
		Name:   repo.Name,
	}, nil
}

func (p *AzureGitOpsProvider) CreatePullRequest(ctx context.Context, input PullRequestInput) (*PullRequest, error) {
	jsmc := jenkinsSCM{}

	u, err := url.Parse(input.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unbale to parse url %q: %w", input.RepositoryURL, err)
	}

	repo, err := jsmc.GetRepository(ctx, p.log, p.client, u)
	if err != nil {
		return nil, err
	}

	headCommit, _ := jsmc.GetCurrentCommitOfBranch(ctx, p.client, repo, input.Head, "")

	if headCommit == "" {
		headCommit, err = jsmc.GetCurrentCommitOfBranch(ctx, p.client, repo, input.Base, "")
		if err != nil {
			return nil, err
		}
		_, _, err = p.client.Git.CreateRef(ctx, repo.FullName, input.Head, headCommit)
		if err != nil {
			return nil, fmt.Errorf("failed to create new branch: %w", err)
		}
	}

	// Note: commits has to be split into a separate update and add commits, Azure
	// does not support updates and additions in the same commit, at least it gave
	// me back an error when I tried.
	for _, commit := range input.Commits {
		request := jsmc.CommitFilesRequest(
			headCommit,
			input.RepositoryURL,
			input.Head,
			commit.CommitMessage,
			commit.Files,
		)

		if _, err := p.sendRawRequest(ctx, request); err != nil {
			return nil, err
		}

		// Fetch the new head commit
		headCommit, _ = jsmc.GetCurrentCommitOfBranch(ctx, p.client, repo, input.Head, "")
	}

	pr, _, err := p.client.PullRequests.Create(ctx, repo.FullName, &scm.PullRequestInput{
		Title: input.Title,
		Head:  input.Head,
		Base:  input.Base,
		Body:  input.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request for branch %q: %w", input.Head, err)
	}

	return &PullRequest{Link: pr.Link}, nil
}

func (p *AzureGitOpsProvider) GetTreeList(ctx context.Context, repoUrl string, sha string, path string) ([]*TreeEntry, error) {
	url, err := GetGitProviderUrl(repoUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	files := []*TreeEntry{}

	fileList, _, err := p.client.Contents.List(ctx, url, path, sha)
	if err != nil {
		return nil, err
	}

	for _, file := range fileList {
		files = append(files, &TreeEntry{
			Name: file.Name,
			Path: file.Path,
			Type: file.Type,
			Size: file.Size,
			SHA:  file.Sha,
			Link: file.Link,
		})
	}

	return files, nil
}

func (p *AzureGitOpsProvider) ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error) {
	url, err := GetGitProviderUrl(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	prList, _, err := p.client.PullRequests.List(ctx, url, &scm.PullRequestListOptions{})
	if err != nil {
		return nil, err
	}

	prs := []*PullRequest{}
	for _, pr := range prList {
		prs = append(prs, &PullRequest{
			Title:       pr.Title,
			Description: pr.Body,
			Link:        pr.Link,
			Merged:      pr.Merged,
		})
	}

	return prs, nil
}

func (p *AzureGitOpsProvider) sendRawRequest(ctx context.Context, request *scm.Request) (*scm.Response, error) {
	resp, err := p.client.Do(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to commit files: %w", err)
	}

	defer resp.Body.Close()

	if resp.Status > 300 {
		err := new(azure.Error)

		_ = json.NewDecoder(resp.Body).Decode(err)

		return resp, err
	}

	return resp, nil
}
