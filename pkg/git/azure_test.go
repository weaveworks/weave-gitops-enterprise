//go:build integration
// +build integration

package git_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	adgit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
)

func TestCreatePullRequestInAzureDevOps(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx := context.Background()
	organisation := "weaveworks"
	project := "weave-gitops-integration"
	repository := fmt.Sprintf("integration-testing-%d", r.Int31n(1000))
	organisationURL := fmt.Sprintf("https://dev.azure.com/%s", organisation)
	branch := fmt.Sprintf("new-branch-%d", r.Int31n(1000))

	// Create an Azure DevOps client
	con := azuredevops.NewPatConnection(organisationURL, os.Getenv("AZURE_DEVOPS_TOKEN"))
	client, err := adgit.NewClient(ctx, con)
	require.NoError(t, err)
	repo, err := client.CreateRepository(ctx, adgit.CreateRepositoryArgs{
		Project: &project,
		GitRepositoryToCreate: &adgit.GitRepositoryCreateOptions{
			Name: &repository,
		},
	})
	require.NoError(t, err)

	// Create a new Azure DevOps repository
	var gitChanges []interface{}
	gitChanges = append(gitChanges, adgit.GitChange{
		ChangeType: &adgit.VersionControlChangeTypeValues.Add,
		Item: &adgit.GitItemDescriptor{
			Path: strPtr("/README.md"),
		},
		NewContent: &adgit.ItemContent{
			Content:     strPtr(fmt.Sprintf("# %s", repository)),
			ContentType: &adgit.ItemContentTypeValues.RawText,
		},
	})
	_, err = client.CreatePush(ctx, adgit.CreatePushArgs{
		Project:      &project,
		RepositoryId: &repository,
		Push: &adgit.GitPush{
			RefUpdates: &[]adgit.GitRefUpdate{
				{
					Name:        strPtr("refs/heads/main"),
					OldObjectId: strPtr("0000000000000000000000000000000000000000"),
				},
			},
			Commits: &[]adgit.GitCommitRef{
				{
					Comment: strPtr("Initial commit"),
					Changes: &gitChanges,
				},
			},
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = client.DeleteRepository(ctx, adgit.DeleteRepositoryArgs{
			RepositoryId: repo.Id,
		})
	})

	// Create a PR using our wrapper
	p, err := git.NewFactory(logr.Discard()).Create(git.AzureDevOpsProviderName, git.WithToken("pat", os.Getenv("AZURE_DEVOPS_TOKEN")))
	require.NoError(t, err)
	content1 := "---\napiVersion: v1\nkind: Namespace\nmetadata:\nname: cert-manager\n"
	content2 := "---\napiVersion: v1\nkind: Namespace\nmetadata:\nname: ingress-nginx\n"
	pr, err := p.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: fmt.Sprintf("%s/%s/_git/%s", organisationURL, project, repository),
		Title:         "Testing PR creation",
		Body:          "Adding namespaces",
		Head:          branch,
		Base:          "main",
		Commits: []git.Commit{
			{
				CommitMessage: "Add cert-manager ns",
				Files: []git.CommitFile{
					{
						Path:    "cert-manager/namespace.yaml",
						Content: &content1,
					},
				},
			},
			{
				CommitMessage: "Add ingress-nginx ns",
				Files: []git.CommitFile{
					{
						Path:    "ingress-nginx/namespace.yaml",
						Content: &content2,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, pr)

	re := regexp.MustCompile(`(\d+)$`)
	id := re.FindString(pr.Link)
	prID, _ := strconv.Atoi(id)
	actual, err := client.GetPullRequest(ctx, adgit.GetPullRequestArgs{
		Project:       &project,
		RepositoryId:  &repository,
		PullRequestId: &prID,
	})
	require.NoError(t, err)

	assert.Equal(t, pr.Title, *actual.Title)
	assert.Equal(t, pr.Description, *actual.Description)
	assert.Equal(t, pr.Merged, false)

	commits, err := client.GetPullRequestCommits(ctx, adgit.GetPullRequestCommitsArgs{
		Project:       &project,
		RepositoryId:  &repository,
		PullRequestId: &prID,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(commits.Value))

	iterations, err := client.GetPullRequestIterations(ctx, adgit.GetPullRequestIterationsArgs{
		Project:        &project,
		RepositoryId:   &repository,
		PullRequestId:  &prID,
		IncludeCommits: boolPtr(true),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(*iterations))

	changes, err := client.GetPullRequestIterationChanges(ctx, adgit.GetPullRequestIterationChangesArgs{
		Project:       &project,
		RepositoryId:  &repository,
		PullRequestId: &prID,
		IterationId:   (*iterations)[0].Id,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(*changes.ChangeEntries))
	assert.Equal(t, "/cert-manager/namespace.yaml", (*changes.ChangeEntries)[0].Item.(map[string]interface{})["path"])
	assert.Equal(t, "/ingress-nginx/namespace.yaml", (*changes.ChangeEntries)[1].Item.(map[string]interface{})["path"])
}

func strPtr(v string) *string { return &v }

func boolPtr(v bool) *bool { return &v }
