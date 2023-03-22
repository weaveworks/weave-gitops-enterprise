//go:build integration
// +build integration

package git_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/xanzy/go-gitlab"
)

func TestCreatePullRequestInGitLab(t *testing.T) {
	// Create a client
	ctx := context.Background()
	gitlabHost := fmt.Sprintf("https://%s", os.Getenv("GIT_PROVIDER_HOSTNAME"))
	client, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(gitlabHost))
	require.NoError(t, err)

	repoName := TestRepositoryNamePrefix + "-group-test"

	// Create a group using a name that doesn't exist already
	groupName := fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	groups, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	assert.NoError(t, err)
	for findGitLabGroup(groups, groupName) != nil {
		groupName = fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	}

	parentGroup, _, err := client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Path:       gitlab.String(groupName),
		Name:       gitlab.String(groupName),
		Visibility: gitlab.Visibility(gitlab.PrivateVisibility),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = client.Groups.DeleteGroup(parentGroup.ID)
		require.NoError(t, err)
	})

	fooGroup, _, err := client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Path:       gitlab.String("foo"),
		Name:       gitlab.String("foo group"),
		ParentID:   gitlab.Int(parentGroup.ID),
		Visibility: gitlab.Visibility(gitlab.PrivateVisibility),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = client.Groups.DeleteGroup(fooGroup.ID)
		require.NoError(t, err)
	})

	repo, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:                 gitlab.String(repoName),
		MergeRequestsEnabled: gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		InitializeWithReadme: gitlab.Bool(true),
		NamespaceID:          gitlab.Int(fooGroup.ID),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = client.Projects.DeleteProject(repo.ID)
		require.NoError(t, err)
	})

	p, err := git.NewFactory(logr.Discard()).Create(
		git.GitLabProviderName,
		git.WithConditionalRequests(),
		git.WithToken("key", os.Getenv("GITLAB_TOKEN")),
		git.WithDomain(os.Getenv("GIT_PROVIDER_HOSTNAME")),
	)
	require.NoError(t, err)
	content := "---\n"
	res, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.HTTPURLToRepo,
		Head:          "feature-01",
		Base:          repo.DefaultBranch,
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
					{
						Path:    "management/cluster-02.yaml",
						Content: &content,
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.Link)
	assert.Equal(t, pr.Title, "New cluster")
	assert.Equal(t, pr.Description, "Creates a cluster through a CAPI template")
	assert.Equal(t, pr.ChangesCount, "2")
}

func TestCreatePullRequestInGitLab_UpdateFiles(t *testing.T) {
	// Create a client
	ctx := context.Background()
	gitlabHost := fmt.Sprintf("https://%s", os.Getenv("GIT_PROVIDER_HOSTNAME"))
	client, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(gitlabHost))
	require.NoError(t, err)
	// Create a repository using a name that doesn't exist already
	repoName := fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	repos, _, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Owned: gitlab.Bool(true),
	})
	assert.NoError(t, err)
	for findGitLabRepo(repos, repoName) != nil {
		repoName = fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	}
	repo, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:                 gitlab.String(repoName),
		MergeRequestsEnabled: gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		InitializeWithReadme: gitlab.Bool(true),
	})
	require.NoError(t, err)
	defer func() {
		_, err = client.Projects.DeleteProject(repo.ID)
		require.NoError(t, err)
	}()

	_, _, err = client.RepositoryFiles.CreateFile(repo.ID, "management/cluster-01.yaml", &gitlab.CreateFileOptions{
		Branch:        gitlab.String(repo.DefaultBranch),
		Content:       gitlab.String("---\n"),
		CommitMessage: gitlab.String("Add cluster manifest"),
	})
	require.NoError(t, err)

	p, err := git.NewFactory(logr.Discard()).Create(
		git.GitLabProviderName,
		git.WithToken("key", os.Getenv("GITLAB_TOKEN")),
		git.WithDomain(os.Getenv("GIT_PROVIDER_HOSTNAME")),
	)
	require.NoError(t, err)
	content := "---\n\n"
	res, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.HTTPURLToRepo,
		Head:          "feature-01",
		Base:          repo.DefaultBranch,
		Title:         "Update cluster",
		Body:          "Updates a cluster through a CAPI template",
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
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.Link)
	assert.Equal(t, pr.Title, "Update cluster")
	assert.Equal(t, pr.Description, "Updates a cluster through a CAPI template")
}

func TestCreatePullRequestInGitLab_DeleteFiles(t *testing.T) {
	// Create a client
	ctx := context.Background()
	gitlabHost := fmt.Sprintf("https://%s", os.Getenv("GIT_PROVIDER_HOSTNAME"))
	client, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(gitlabHost))
	require.NoError(t, err)
	// Create a repository using a name that doesn't exist already
	repoName := fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	repos, _, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Owned: gitlab.Bool(true),
	})
	assert.NoError(t, err)
	for findGitLabRepo(repos, repoName) != nil {
		repoName = fmt.Sprintf("%s-%03d", TestRepositoryNamePrefix, rand.Intn(1000))
	}
	repo, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:                 gitlab.String(repoName),
		MergeRequestsEnabled: gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		InitializeWithReadme: gitlab.Bool(true),
	})
	require.NoError(t, err)
	defer func() {
		_, err = client.Projects.DeleteProject(repo.ID)
		require.NoError(t, err)
	}()

	_, _, err = client.RepositoryFiles.CreateFile(repo.ID, "management/cluster-01.yaml", &gitlab.CreateFileOptions{
		Branch:        gitlab.String(repo.DefaultBranch),
		Content:       gitlab.String("---\n"),
		CommitMessage: gitlab.String("Add cluster manifest"),
	})
	require.NoError(t, err)

	p, err := git.NewFactory(logr.Discard()).Create(
		git.GitLabProviderName,
		git.WithToken("key", os.Getenv("GITLAB_TOKEN")),
		git.WithDomain(os.Getenv("GIT_PROVIDER_HOSTNAME")),
	)
	require.NoError(t, err)
	res, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: repo.HTTPURLToRepo,
		Head:          "feature-01",
		Base:          repo.DefaultBranch,
		Title:         "Delete cluster",
		Body:          "Deletes a cluster",
		Commits: []git.Commit{
			{
				CommitMessage: "Delete cluster manifest",
				Files: []git.CommitFile{
					{
						Path:    "management/cluster-01.yaml",
						Content: nil,
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.Link)
	assert.Equal(t, pr.Title, "Delete cluster")
	assert.Equal(t, pr.Description, "Deletes a cluster")
}

func findGitLabGroup(groups []*gitlab.Group, name string) *gitlab.Group {
	if name == "" {
		return nil
	}
	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}
	return nil
}

func findGitLabRepo(repos []*gitlab.Project, name string) *gitlab.Project {
	if name == "" {
		return nil
	}
	for _, repo := range repos {
		if repo.Name == name {
			return repo
		}
	}
	return nil
}
