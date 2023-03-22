//go:build integration
// +build integration

package git_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"golang.org/x/oauth2"
)

const (
	TestRepositoryNamePrefix = "wge-integration-test-repo"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCreatePullRequestInGitHubOrganization(t *testing.T) {
	// Create a client
	ctx := context.Background()
	client := github.NewClient(
		oauth2.NewClient(ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			),
		),
	)

	// Create a repository using a name that doesn't exist already
	repoName := fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	repos, _, err := client.Repositories.ListByOrg(ctx, os.Getenv("GITHUB_ORG"), nil)
	assert.NoError(t, err)
	for findGitHubRepo(repos, repoName) != nil {
		repoName = fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	}
	repo, _, err := client.Repositories.Create(ctx, os.Getenv("GITHUB_ORG"), &github.Repository{
		Name:     github.String(repoName),
		Private:  github.Bool(true),
		AutoInit: github.Bool(true),
	})
	require.NoError(t, err)
	defer func() {
		_, err = client.Repositories.Delete(ctx, os.Getenv("GITHUB_ORG"), repo.GetName())
		require.NoError(t, err)
	}()

	p, err := git.NewFactory(logr.Discard()).Create(git.GitHubProviderName, git.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")))
	require.NoError(t, err)
	content := "---\n"
	res, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.GetCloneURL(),
		Head:          "feature-01",
		Base:          repo.GetDefaultBranch(),
		Title:         "New cluster",
		Body:          "Creates a cluster through a CAPI template",
		Commits: []git.Commit{
			{
				CommitMessage: "Add cluster manifest",
				Files: []git.CommitFile{
					{
						Path:    "management/cluster-01.yaml",
						Content: &content,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	pr, _, err := client.PullRequests.Get(ctx, os.Getenv("GITHUB_ORG"), repo.GetName(), 1) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.Link)
	assert.Equal(t, pr.GetTitle(), "New cluster")
	assert.Equal(t, pr.GetBody(), "Creates a cluster through a CAPI template")
	assert.Equal(t, pr.GetChangedFiles(), 1)
}

func TestCreatePullRequestInGitHubUser(t *testing.T) {
	// Create a client
	ctx := context.Background()
	client := github.NewClient(
		oauth2.NewClient(ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			),
		),
	)
	// Create a repository using a name that doesn't exist already
	repoName := fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	repos, _, err := client.Repositories.List(ctx, os.Getenv("GITHUB_USER"), nil)
	assert.NoError(t, err)
	for findGitHubRepo(repos, repoName) != nil {
		repoName = fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	}
	repo, _, err := client.Repositories.Create(ctx, "", &github.Repository{
		Name:     github.String(repoName),
		Private:  github.Bool(true),
		AutoInit: github.Bool(true),
	})
	require.NoError(t, err)
	defer func() {
		_, err = client.Repositories.Delete(ctx, os.Getenv("GITHUB_USER"), repo.GetName())
		require.NoError(t, err)
	}()

	p, err := git.NewFactory(logr.Discard()).Create(git.GitHubProviderName, git.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")))
	require.NoError(t, err)
	content := "---\n"
	res, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.GetCloneURL(),
		Head:          "feature-01",
		Base:          repo.GetDefaultBranch(),
		Title:         "New cluster",
		Body:          "Creates a cluster through a CAPI template",
		Commits: []git.Commit{
			{
				CommitMessage: "Add cluster manifest",
				Files: []git.CommitFile{
					{
						Path:    "management/cluster-01.yaml",
						Content: &content,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	pr, _, err := client.PullRequests.Get(ctx, os.Getenv("GITHUB_USER"), repo.GetName(), 1) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.Link)
	assert.Equal(t, pr.GetTitle(), "New cluster")
	assert.Equal(t, pr.GetBody(), "Creates a cluster through a CAPI template")
	assert.Equal(t, pr.GetAdditions(), 1)

	res, err = p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.GetCloneURL(),
		Head:          "feature-02",
		Base:          "feature-01",
		Title:         "Delete cluster",
		Body:          "Deletes a cluster via gitops",
		Commits: []git.Commit{
			{
				CommitMessage: "Remove cluster manifest",
				Files: []git.CommitFile{
					{
						Path:    "management/cluster-01.yaml",
						Content: nil,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	pr, _, err = client.PullRequests.Get(ctx, os.Getenv("GITHUB_USER"), repo.GetName(), 2) // #PR is 2 because it is the 2nd PR for a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.Link)
	assert.Equal(t, pr.GetTitle(), "Delete cluster")
	assert.Equal(t, pr.GetBody(), "Deletes a cluster via gitops")
	assert.Equal(t, pr.GetDeletions(), 1)
}

func findGitHubRepo(repos []*github.Repository, name string) *github.Repository {
	if name == "" {
		return nil
	}
	for _, repo := range repos {
		if repo.GetName() == name {
			return repo
		}
	}
	return nil
}
