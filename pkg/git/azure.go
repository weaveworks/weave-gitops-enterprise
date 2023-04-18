package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/azure"
	"github.com/jenkins-x/go-scm/scm/factory"
)

const (
	AzureDevOpsProviderName string = "azure-devops"
)

// AzureDevOpsProvider is used to interact with the Azure DevOps API.
type AzureDevOpsProvider struct {
	log    logr.Logger
	client *scm.Client
}

func NewAzureDevOpsProvider(log logr.Logger) (Provider, error) {
	return &AzureDevOpsProvider{
		log: log,
	}, nil
}

func (p *AzureDevOpsProvider) Setup(opts ProviderOption) error {
	if opts.Token == "" {
		return fmt.Errorf("missing required option: Token")
	}

	if opts.Hostname == "" {
		opts.Hostname = "dev.azure.com"
	}

	var err error

	p.client, err = factory.NewClient("azure", fmt.Sprintf("https://%s", opts.Hostname), opts.Token)

	return err
}

func (p *AzureDevOpsProvider) GetRepository(ctx context.Context, repoURL string) (*Repository, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unbale to parse url %q: %w", repoURL, err)
	}

	dscm := JenkinsSCM{}

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

func (p *AzureDevOpsProvider) CreatePullRequest(ctx context.Context, input PullRequestInput) (*PullRequest, error) {
	jsmc := JenkinsSCM{}

	u, err := url.Parse(input.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %q: %w", input.RepositoryURL, err)
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

		if _, err := p.sendRawRequest(ctx, request, nil); err != nil {
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

	return &PullRequest{Title: pr.Title, Description: pr.Body, Link: pr.Link, Merged: pr.Merged}, nil
}

func (p *AzureDevOpsProvider) GetTreeList(ctx context.Context, repoURL string, sha string, path string) ([]*TreeEntry, error) {
	repoURL, err := GetGitProviderUrl(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	jsmc := JenkinsSCM{}

	request, err := jsmc.ListContents(repoURL, path, sha)
	if err != nil {
		return nil, fmt.Errorf("unable to create list request: %w", err)
	}

	list := jscmContentList{}
	if _, err := p.sendRawRequest(ctx, request, &list); err != nil {
		return nil, fmt.Errorf("failed list request: %w", err)
	}

	files := []*TreeEntry{}
	for _, file := range list.Value {
		files = append(files, &TreeEntry{
			Path: file.Path,
			SHA:  file.CommitID,
			Type: file.GitObjectType,
			Link: file.URL,
		})
	}

	return files, nil
}

func (p *AzureDevOpsProvider) ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error) {
	repoURL, err := GetGitProviderUrl(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git provider url: %w", err)
	}

	// At this point GetGitProviderUrl should fail if it's not a valid URL.
	u, _ := url.Parse(repoURL)
	jsmc := JenkinsSCM{}

	repo, err := jsmc.GetRepository(ctx, p.log, p.client, u)
	if err != nil {
		return nil, fmt.Errorf("unable to find repository: %w", err)
	}

	prList, _, err := p.client.PullRequests.List(ctx, repo.FullName, &scm.PullRequestListOptions{})
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

func (p *AzureDevOpsProvider) sendRawRequest(ctx context.Context, request *scm.Request, response interface{}) (*scm.Response, error) {
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

	if response != nil {
		_ = json.NewDecoder(resp.Body).Decode(response)
	}

	return resp, nil
}
