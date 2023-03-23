package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
)

type goGitProvider struct{}

func (g goGitProvider) CreateBranch(ctx context.Context, log logr.Logger, repo gitprovider.OrgRepository, base, head string) error {
	var commits []gitprovider.Commit
	err := retry.OnError(DefaultBackoff,
		func(err error) bool {
			// Ideally this should return true only for 404 (gitprovider.ErrNotFound) and 409 errors
			return true
		}, func() error {
			var err error
			commits, err = repo.Commits().ListPage(ctx, base, 1, 1)
			if err != nil {
				log.Info("Retrying getting the repository")
				return err
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("unable to get most recent commit for branch %q: %w", base, err)
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits were found for branch %q, is the repository empty?", base)
	}

	err = repo.Branches().Create(ctx, head, commits[0].Get().Sha)
	if err != nil {
		return fmt.Errorf("unable to create new branch %q from commit %q in branch %q: %w", head, commits[0].Get().Sha, base, err)
	}

	return nil
}

func (g goGitProvider) WriteFilesToBranch(ctx context.Context, log logr.Logger, req writeFilesToBranchRequest, repo gitprovider.OrgRepository) error {
	if req.CreateBranch {
		if err := g.CreateBranch(ctx, log, repo, req.BaseBranch, req.HeadBranch); err != nil {
			return err
		}
	}

	// Loop through all the commits and write the files.
	for _, c := range req.Commits {
		// Adapt for go-git-providers
		adapted := make([]gitprovider.CommitFile, 0)
		for idx := range c.Files {
			adapted = append(adapted, gitprovider.CommitFile{
				Path:    &c.Files[idx].Path,
				Content: c.Files[idx].Content,
			})
		}

		commit, err := repo.Commits().Create(ctx, req.HeadBranch, c.CommitMessage, adapted)
		if err != nil {
			return fmt.Errorf("unable to commit changes to %q: %w", req.HeadBranch, err)
		}
		log.WithValues("sha", commit.Get().Sha, "branch", req.HeadBranch).Info("Files committed")
	}

	return nil
}

func (g goGitProvider) CreatePullRequest(ctx context.Context, log logr.Logger, req createPullRequestRequest, repo gitprovider.OrgRepository) (*createPullRequestResponse, error) {
	pr, err := repo.PullRequests().Create(ctx, req.Title, req.HeadBranch, req.BaseBranch, req.Description)
	if err != nil {
		return nil, fmt.Errorf("unable to create new pull request for branch %q: %w", req.HeadBranch, err)
	}
	log.WithValues("pullRequestURL", pr.Get().WebURL).Info("Created pull request")

	return &createPullRequestResponse{
		WebURL: pr.Get().WebURL,
	}, nil
}

func (g goGitProvider) GetUpdatedFiles(
	ctx context.Context,
	reqFiles []CommitFile,
	client gitprovider.Client,
	repoURL,
	branch string) ([]CommitFile, error) {
	var updatedFiles []CommitFile

	for _, file := range reqFiles {
		// if file content is empty, then it's a delete operation,
		// so we don't need to check if the file exists
		if file.Content == nil {
			continue
		}

		dirPath, _ := filepath.Split(file.Path)

		treeEntries, err := g.GetTreeList(ctx, client, repoURL, branch, dirPath, true)
		if err != nil {
			return nil, fmt.Errorf("error getting list of trees in repo: %s@%s: %w", repoURL, branch, err)
		}

		for _, treeEntry := range treeEntries {
			if treeEntry.Path == file.Path {
				updatedFiles = append(updatedFiles, CommitFile{
					Path:    treeEntry.Path,
					Content: nil,
				})
			}
		}
	}

	return updatedFiles, nil
}

// GetTreeList retrieves list of tree files from gitprovider given the sha/branch
func (g goGitProvider) GetTreeList(ctx context.Context, client gitprovider.Client, repoUrl string, sha string, path string, recursive bool) ([]*gitprovider.TreeEntry, error) {
	repo, err := g.GetRepository(ctx, logr.Discard(), client, repoUrl)
	if err != nil {
		return nil, err
	}

	treePaths, err := repo.Trees().List(ctx, sha, path, recursive)
	if err != nil {
		return nil, err
	}

	return treePaths, nil
}

func (g goGitProvider) GetRepository(ctx context.Context, log logr.Logger, client gitprovider.Client, url string) (gitprovider.OrgRepository, error) {
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %q: %w", url, err)
	}

	ref.Domain = addSchemeToDomain(ref.Domain)
	ref = WithCombinedSubOrgs(*ref)

	var repo gitprovider.OrgRepository
	err = retry.OnError(DefaultBackoff,
		func(err error) bool { return errors.Is(err, gitprovider.ErrNotFound) },
		func() error {
			var err error
			repo, err = client.OrgRepositories().Get(ctx, *ref)
			if err != nil {
				log.Info("Retrying getting the repository")
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository %q: %w, (client domain: %s)", url, err, client.SupportedDomain())
	}

	return repo, nil
}

func (g goGitProvider) GetBitbucketRepository(ctx context.Context, log logr.Logger, client gitprovider.Client, url string) (gitprovider.OrgRepository, error) {
	re := regexp.MustCompile(`://(?P<host>[^/]+)/(.+/)?(?P<key>[^/]+)/(?P<repo>[^/]+)\.git`)
	match := re.FindStringSubmatch(url)
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	if len(result) != 3 {
		return nil, fmt.Errorf("unable to parse repository URL %q using regex %q", url, re.String())
	}

	orgRef := &gitprovider.OrganizationRef{
		Domain:       result["host"],
		Organization: result["key"],
	}
	ref := &gitprovider.OrgRepositoryRef{
		OrganizationRef: *orgRef,
		RepositoryName:  result["repo"],
	}
	ref.SetKey(result["key"])
	ref.Domain = addSchemeToDomain(ref.Domain)

	var repo gitprovider.OrgRepository
	err := retry.OnError(DefaultBackoff,
		func(err error) bool { return errors.Is(err, gitprovider.ErrNotFound) },
		func() error {
			var err error
			repo, err = client.OrgRepositories().Get(ctx, *ref)
			if err != nil {
				log.Info("Retrying getting the repository")
				return err
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("unable to get repository %q: %w, (client domain: %s)", url, err, client.SupportedDomain())
	}

	return repo, nil
}

func (g goGitProvider) ListPullRequests(ctx context.Context, repo gitprovider.OrgRepository) ([]*PullRequest, error) {
	prList, err := repo.PullRequests().List(ctx)
	if err != nil {
		return nil, err
	}

	prs := []*PullRequest{}
	for _, pr := range prList {
		prs = append(prs, &PullRequest{
			Title:       pr.Get().Title,
			Description: pr.Get().Description,
			Link:        pr.Get().WebURL,
			Merged:      pr.Get().Merged,
		})
	}

	return prs, nil
}
