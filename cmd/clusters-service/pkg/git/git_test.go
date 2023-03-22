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

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	csgit "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

const (
	TestRepositoryNamePrefix = "capi-server-test-repo"
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

	s := csgit.NewGitProviderService(logr.Discard())
	path := "management/cluster-01.yaml"
	content := "---\n"
	res, err := s.WriteFilesToBranchAndCreatePullRequest(ctx, csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITHUB_TOKEN"),
			Type:     git.GitHubProviderName,
			Hostname: "github.com",
		},
		RepositoryURL: repo.GetCloneURL(),
		HeadBranch:    "feature-01",
		BaseBranch:    repo.GetDefaultBranch(),
		Title:         "New cluster",
		Description:   "Creates a cluster through a CAPI template",
		CommitMessage: "Add cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: &content,
			},
		},
	})
	require.NoError(t, err)

	pr, _, err := client.PullRequests.Get(ctx, os.Getenv("GITHUB_ORG"), repo.GetName(), 1) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.WebURL)
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

	s := csgit.NewGitProviderService(logr.Discard())
	path := "management/cluster-01.yaml"
	content := "---\n"
	res, err := s.WriteFilesToBranchAndCreatePullRequest(ctx, csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITHUB_TOKEN"),
			Type:     git.GitHubProviderName,
			Hostname: "github.com",
		},
		RepositoryURL: repo.GetCloneURL(),
		HeadBranch:    "feature-01",
		BaseBranch:    repo.GetDefaultBranch(),
		Title:         "New cluster",
		Description:   "Creates a cluster through a CAPI template",
		CommitMessage: "Add cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: &content,
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.PullRequests.Get(ctx, os.Getenv("GITHUB_USER"), repo.GetName(), 1) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.WebURL)
	assert.Equal(t, pr.GetTitle(), "New cluster")
	assert.Equal(t, pr.GetBody(), "Creates a cluster through a CAPI template")
	assert.Equal(t, pr.GetAdditions(), 1)

	res, err = s.WriteFilesToBranchAndCreatePullRequest(ctx, csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITHUB_TOKEN"),
			Type:     git.GitHubProviderName,
			Hostname: "github.com",
		},
		RepositoryURL: repo.GetCloneURL(),
		HeadBranch:    "feature-02",
		BaseBranch:    "feature-01",
		Title:         "Delete cluster",
		Description:   "Deletes a cluster via gitops",
		CommitMessage: "Remove cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: nil,
			},
		},
	})
	assert.NoError(t, err)
	pr, _, err = client.PullRequests.Get(ctx, os.Getenv("GITHUB_USER"), repo.GetName(), 2) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.GetHTMLURL(), res.WebURL)
	assert.Equal(t, pr.GetTitle(), "Delete cluster")
	assert.Equal(t, pr.GetBody(), "Deletes a cluster via gitops")
	assert.Equal(t, pr.GetDeletions(), 1)
}

func TestCreatePullRequestInGitLab(t *testing.T) {

	// Create a client
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

	s := csgit.NewGitProviderService(logr.Discard())
	path := "management/cluster-01.yaml"
	content := "---\n"
	res, err := s.WriteFilesToBranchAndCreatePullRequest(context.Background(), csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITLAB_TOKEN"),
			Type:     git.GitLabProviderName,
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: repo.HTTPURLToRepo,
		HeadBranch:    "feature-01",
		BaseBranch:    repo.DefaultBranch,
		Title:         "New cluster",
		Description:   "Creates a cluster through a CAPI template",
		CommitMessage: "Add cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: &content,
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil) // #PR is 1 because it is a new repo
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.WebURL)
	assert.Equal(t, pr.Title, "New cluster")
	assert.Equal(t, pr.Description, "Creates a cluster through a CAPI template")
}

func TestCreatePullRequestInGitLab_UpdateFiles(t *testing.T) {
	// Create a client
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

	s := csgit.NewGitProviderService(logr.Discard())
	path := "management/cluster-01.yaml"
	content := "---\n\n"
	res, err := s.WriteFilesToBranchAndCreatePullRequest(context.Background(), csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITLAB_TOKEN"),
			Type:     git.GitLabProviderName,
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: repo.HTTPURLToRepo,
		HeadBranch:    "feature-01",
		BaseBranch:    repo.DefaultBranch,
		Title:         "Update cluster",
		Description:   "Updates a cluster through a CAPI template",
		CommitMessage: "Update cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: &content,
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.WebURL)
	assert.Equal(t, pr.Title, "Update cluster")
	assert.Equal(t, pr.Description, "Updates a cluster through a CAPI template")
}

func TestCreatePullRequestInGitLab_DeleteFiles(t *testing.T) {
	// Create a client
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

	s := csgit.NewGitProviderService(logr.Discard())
	path := "management/cluster-01.yaml"
	res, err := s.WriteFilesToBranchAndCreatePullRequest(context.Background(), csgit.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: csgit.GitProvider{
			Token:    os.Getenv("GITLAB_TOKEN"),
			Type:     git.GitLabProviderName,
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: repo.HTTPURLToRepo,
		HeadBranch:    "feature-01",
		BaseBranch:    repo.DefaultBranch,
		Title:         "Delete cluster",
		Description:   "Deletes a cluster",
		CommitMessage: "Delete cluster manifest",
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: nil,
			},
		},
	})
	assert.NoError(t, err)

	pr, _, err := client.MergeRequests.GetMergeRequest(repo.ID, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, pr.WebURL, res.WebURL)
	assert.Equal(t, pr.Title, "Delete cluster")
	assert.Equal(t, pr.Description, "Deletes a cluster")
}

func TestGetGitProviderUrl(t *testing.T) {
	expected := "https://github.com/user/repo.git"

	viper.SetDefault("capi-templates-repository-api-url", "https://github.com/user/repo.git")
	dummyUrl := "example.com"
	repoURL, err := csgit.GetGitProviderUrl(dummyUrl)
	require.NoError(t, err)
	assert.Equal(t, expected, repoURL)

	viper.Reset()

	gitUrl := "git@github.com:user/repo.git"
	repoURL, err = csgit.GetGitProviderUrl(gitUrl)
	require.NoError(t, err)
	assert.Equal(t, expected, repoURL)

	httpsUrl := "https://github.com/user/repo.git"
	repoURL, err = csgit.GetGitProviderUrl(httpsUrl)
	require.NoError(t, err)
	assert.Equal(t, expected, repoURL)
}

func TestWithCombinedSubOrg(t *testing.T) {
	ref, err := gitprovider.ParseOrgRepositoryURL("https://gitlab.com/org/sub1/sub2/sub3/repo")
	assert.NoError(t, err)
	newRef := csgit.WithCombinedSubOrgs(*ref)

	// Where they are the still the same
	assert.Equal(t, "repo", ref.RepositoryName)
	assert.Equal(t, "repo", newRef.RepositoryName)

	assert.Equal(t, "https://gitlab.com/org/sub1/sub2/sub3/repo.git", ref.GetCloneURL("https"))
	assert.Equal(t, "https://gitlab.com/org/sub1/sub2/sub3/repo.git", newRef.GetCloneURL("https"))

	assert.Equal(t, "https://gitlab.com/org/sub1/sub2/sub3/repo", ref.String())
	assert.Equal(t, "https://gitlab.com/org/sub1/sub2/sub3/repo", newRef.String())

	// Where they now differ
	assert.Equal(t, "org", ref.Organization)
	assert.Equal(t, "org/sub1/sub2/sub3", newRef.Organization)

	assert.Equal(t, []string{"sub1", "sub2", "sub3"}, ref.SubOrganizations)
	assert.Equal(t, []string(nil), newRef.SubOrganizations)
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
