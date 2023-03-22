package git

import (
	"context"
)

// Provider defines the interface that WGE will use to interact
// with a git provider.
type Provider interface {
	// CreatePullRequest pushes a set of changes to a branch
	// and then creates a pull request from it. Typically this
	// is a two-step process that involves making multiple API
	// requests.
	CreatePullRequest(context.Context, PullRequestInput) (*PullRequest, error)

	// Setup configures the provider from ProverOption.
	Setup(ProviderOption) error

	GetRepository(ctx context.Context, repoURL string) (*Repository, error)
	GetTreeList(ctx context.Context, repoUrl, sha, path string) ([]*TreeEntry, error)
	ListPullRequests(ctx context.Context, repoURL string) ([]*PullRequest, error)
}

// CommitFile represents the contents of file in the repository.
type CommitFile struct {
	Path string
	// Content represents the content of the change. If nil,
	// this results in a deletion.
	Content *string
}

// TODO get rid of this, intermediate structure
type GitProvider struct {
	Token     string
	TokenType string
	Type      string
	Hostname  string
}

// TODO get rid of this, intermediate structure
type Commit struct {
	CommitMessage string
	Files         []CommitFile
}
