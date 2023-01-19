package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	go_git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const deleteFilesCommitMessage = "Delete old files for resources"

var DefaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 20 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

type Provider interface {
	WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req WriteFilesToBranchAndCreatePullRequestRequest) (*WriteFilesToBranchAndCreatePullRequestResponse, error)
	CloneRepoToTempDir(req CloneRepoToTempDirRequest) (*CloneRepoToTempDirResponse, error)
	GetRepository(ctx context.Context, gp GitProvider, url string) (gitprovider.OrgRepository, error)
	GetTreeList(ctx context.Context, gp GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*gitprovider.TreeEntry, error)
	ListPullRequests(ctx context.Context, gp GitProvider, url string) ([]gitprovider.PullRequest, error)
}

type GitProviderService struct {
	log logr.Logger
}

func NewGitProviderService(log logr.Logger) *GitProviderService {
	return &GitProviderService{
		log: log,
	}
}

type GitProvider struct {
	Token     string
	TokenType string
	Type      string
	Hostname  string
}

type WriteFilesToBranchAndCreatePullRequestRequest struct {
	GitProvider       GitProvider
	RepositoryURL     string
	ReposistoryAPIURL string
	HeadBranch        string
	BaseBranch        string
	Title             string
	Description       string
	CommitMessage     string
	Files             []gitprovider.CommitFile
}

type WriteFilesToBranchAndCreatePullRequestResponse struct {
	WebURL string
}

type CloneRepoToTempDirRequest struct {
	GitProvider   GitProvider
	RepositoryURL string
	BaseBranch    string
	ParentDir     string
}

type CloneRepoToTempDirResponse struct {
	Repo *GitRepo
}

// WriteFilesToBranchAndCreatePullRequest writes a set of provided files
// to a new branch and creates a new pull request for that branch.
// It returns the URL of the pull request.
func (s *GitProviderService) WriteFilesToBranchAndCreatePullRequest(ctx context.Context,
	req WriteFilesToBranchAndCreatePullRequestRequest) (*WriteFilesToBranchAndCreatePullRequestResponse, error) {
	commits := []Commit{}

	repoURL, err := GetGitProviderUrl(req.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get git porivder url: %w", err)
	}

	repo, err := s.GetRepository(ctx, req.GitProvider, repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo: %w", err)
	}

	// Gitlab doesn't support createOrUpdate, so we need to check if the file exists
	// and if it does, we need to create a commit to delete the file.
	if req.GitProvider.Type == "gitlab" {
		deletedFiles, err := s.getUpdatedFiles(ctx, req.Files, req.GitProvider, repoURL, req.BaseBranch)
		if err != nil {
			return nil, fmt.Errorf("unable to get files from tree list: %w", err)
		}

		// If there are files to delete, append them to the map of changes to be deleted.
		if len(deletedFiles) > 0 {
			commits = append(commits, Commit{
				CommitMessage: deleteFilesCommitMessage,
				Files:         deletedFiles,
			})
		}
	}

	// Add the files to be created to the map of changes.
	commits = append(commits, Commit{
		CommitMessage: req.CommitMessage,
		Files:         req.Files,
	})

	if err := s.writeFilesToBranch(ctx, writeFilesToBranchRequest{
		Repository: repo,
		HeadBranch: req.HeadBranch,
		BaseBranch: req.BaseBranch,
		Commits:    commits,
	}); err != nil {
		return nil, fmt.Errorf("unable to write files to branch %q: %w", req.HeadBranch, err)
	}

	res, err := s.createPullRequest(ctx, createPullRequestRequest{
		Repository:  repo,
		HeadBranch:  req.HeadBranch,
		BaseBranch:  req.BaseBranch,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request for branch %q: %w", req.HeadBranch, err)
	}

	return &WriteFilesToBranchAndCreatePullRequestResponse{
		WebURL: res.WebURL,
	}, nil
}

func (s *GitProviderService) CloneRepoToTempDir(req CloneRepoToTempDirRequest) (*CloneRepoToTempDirResponse, error) {
	s.log.Info("Creating a temp directory...")
	gitDir, err := os.MkdirTemp(req.ParentDir, "git-")
	if err != nil {
		return nil, err
	}
	s.log.Info("Temp directory created.", "dir", gitDir)

	s.log.Info("Cloning the Git repository...", "repository", req.RepositoryURL, "dir", gitDir)

	repo, err := go_git.PlainClone(gitDir, false, &go_git.CloneOptions{
		URL: req.RepositoryURL,
		Auth: &http.BasicAuth{
			Username: "abc123",
			Password: req.GitProvider.Token,
		},
		ReferenceName: plumbing.NewBranchReferenceName(req.BaseBranch),

		SingleBranch: true,
		Tags:         go_git.NoTags,
	})
	if err != nil {
		return nil, err
	}

	s.log.Info("Cloned repository", "repository", req.RepositoryURL)

	gitRepo := &GitRepo{
		WorktreeDir: gitDir,
		Repo:        repo,
		Auth: &http.BasicAuth{
			Username: "abc123",
			Password: req.GitProvider.Token,
		},
	}

	return &CloneRepoToTempDirResponse{
		Repo: gitRepo,
	}, nil
}

type GitRepo struct {
	WorktreeDir string
	Repo        *go_git.Repository
	Auth        *http.BasicAuth
}

func (s *GitProviderService) GetRepository(ctx context.Context, gp GitProvider, url string) (gitprovider.OrgRepository, error) {
	c, err := getGitProviderClient(gp)
	if err != nil {
		return nil, fmt.Errorf("unable to get a git provider client for %q: %w", gp.Type, err)
	}

	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse repository URL %q: %w", url, err)
	}

	ref.Domain = addSchemeToDomain(ref.Domain)

	var repo gitprovider.OrgRepository
	err = retry.OnError(DefaultBackoff,
		func(err error) bool { return errors.Is(err, gitprovider.ErrNotFound) },
		func() error {
			var err error
			repo, err = c.OrgRepositories().Get(ctx, *ref)
			if err != nil {
				s.log.Info("Retrying getting the repository")
				return err
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("unable to get repository %q: %w, (client domain: %s)", url, err, c.SupportedDomain())
	}

	return repo, nil
}

// GetTreeList retrieves list of tree files from gitprovider given the sha/branch
func (s *GitProviderService) GetTreeList(ctx context.Context, gp GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*gitprovider.TreeEntry, error) {
	repo, err := s.GetRepository(ctx, gp, repoUrl)
	if err != nil {
		return nil, err
	}

	treePaths, err := repo.Trees().List(ctx, sha, path, recursive)
	if err != nil {
		return nil, err
	}

	return treePaths, nil
}

func (s *GitProviderService) ListPullRequests(ctx context.Context, gp GitProvider, repoURL string) ([]gitprovider.PullRequest, error) {
	repo, err := s.GetRepository(ctx, gp, repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo: %w", err)
	}

	prs, err := repo.PullRequests().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list pull requests for %q: %w", repoURL, err)
	}

	return prs, nil
}

type writeFilesToBranchRequest struct {
	Repository gitprovider.OrgRepository
	HeadBranch string
	BaseBranch string
	Commits    []Commit
}

type Commit struct {
	CommitMessage string
	Files         []gitprovider.CommitFile
}

func (s *GitProviderService) writeFilesToBranch(ctx context.Context, req writeFilesToBranchRequest) error {

	var commits []gitprovider.Commit
	err := retry.OnError(DefaultBackoff,
		func(err error) bool {
			// Ideally this should return true only for 404 (gitprovider.ErrNotFound) and 409 errors
			return true
		}, func() error {
			var err error
			commits, err = req.Repository.Commits().ListPage(ctx, req.BaseBranch, 1, 1)
			if err != nil {
				s.log.Info("Retrying getting the repository")
				return err
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("unable to get most recent commit for branch %q: %w", req.BaseBranch, err)
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits were found for branch %q, is the repository empty?", req.BaseBranch)
	}

	err = req.Repository.Branches().Create(ctx, req.HeadBranch, commits[0].Get().Sha)
	if err != nil {
		return fmt.Errorf("unable to create new branch %q from commit %q in branch %q: %w", req.HeadBranch, commits[0].Get().Sha, req.BaseBranch, err)
	}

	// Loop through all the commits and write the files.
	for _, c := range req.Commits {
		commit, err := req.Repository.Commits().Create(ctx, req.HeadBranch, c.CommitMessage, c.Files)
		if err != nil {
			return fmt.Errorf("unable to commit changes to %q: %w", req.HeadBranch, err)
		}
		s.log.WithValues("sha", commit.Get().Sha, "branch", req.HeadBranch).Info("Files committed")
	}

	return nil
}

type createPullRequestRequest struct {
	Repository  gitprovider.OrgRepository
	HeadBranch  string
	BaseBranch  string
	Title       string
	Description string
}

type createPullRequestResponse struct {
	WebURL string
}

func (s *GitProviderService) createPullRequest(ctx context.Context, req createPullRequestRequest) (*createPullRequestResponse, error) {
	pr, err := req.Repository.PullRequests().Create(ctx, req.Title, req.HeadBranch, req.BaseBranch, req.Description)
	if err != nil {
		return nil, fmt.Errorf("unable to create new pull request for branch %q: %w", req.HeadBranch, err)
	}
	s.log.WithValues("pullRequestURL", pr.Get().WebURL).Info("Created pull request")

	return &createPullRequestResponse{
		WebURL: pr.Get().WebURL,
	}, nil
}

func getGitProviderClient(gpi GitProvider) (gitprovider.Client, error) {
	var client gitprovider.Client
	var err error

	// quirk of ggp
	hostname := addSchemeToDomain(gpi.Hostname)

	switch gpi.Type {
	case "github":
		if gpi.Hostname != "github.com" {
			client, err = github.NewClient(
				gitprovider.WithOAuth2Token(gpi.Token),
				gitprovider.WithDomain(hostname),
			)
		} else {
			client, err = github.NewClient(
				gitprovider.WithOAuth2Token(gpi.Token),
			)
		}
		if err != nil {
			return nil, err
		}
	case "gitlab":
		if gpi.Hostname != "gitlab.com" {
			client, err = gitlab.NewClient(gpi.Token, gpi.TokenType, gitprovider.WithDomain(hostname), gitprovider.WithConditionalRequests(true))
		} else {
			client, err = gitlab.NewClient(gpi.Token, gpi.TokenType, gitprovider.WithConditionalRequests(true))
		}
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("the Git provider %q is not supported", gpi.Type)
	}
	return client, err
}

func GetGitProviderUrl(giturl string) (string, error) {
	repositoryAPIURL := viper.GetString("capi-templates-repository-api-url")
	if repositoryAPIURL != "" {
		return repositoryAPIURL, nil
	}

	ep, err := transport.NewEndpoint(giturl)
	if err != nil {
		return "", err
	}
	if ep.Protocol == "http" || ep.Protocol == "https" {
		return giturl, nil
	}

	httpsEp := transport.Endpoint{Protocol: "https", Host: ep.Host, Path: ep.Path}

	return httpsEp.String(), nil
}

func addSchemeToDomain(domain string) string {
	// Fixing https:// again (ggp quirk)
	if domain != "github.com" && domain != "gitlab.com" && !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		return "https://" + domain
	}
	return domain
}

func (s *GitProviderService) getUpdatedFiles(
	ctx context.Context,
	reqFiles []gitprovider.CommitFile,
	gp GitProvider,
	repoURL,
	branch string) ([]gitprovider.CommitFile, error) {
	var updatedFiles []gitprovider.CommitFile

	for _, file := range reqFiles {
		// if file content is empty, then it's a delete operation
		// so we don't need to check if the file exists
		if file.Content == nil {
			continue
		}

		dirPath, _ := filepath.Split(*file.Path)

		treeEntries, err := s.GetTreeList(ctx, gp, repoURL, branch, dirPath, true)
		if err != nil {
			return nil, fmt.Errorf("error getting list of trees in repo: %s@%s: %w", repoURL, branch, err)
		}

		for _, treeEntry := range treeEntries {
			if treeEntry.Path == *file.Path {
				updatedFiles = append(updatedFiles, gitprovider.CommitFile{
					Path:    &treeEntry.Path,
					Content: nil,
				})
			}
		}
	}

	return updatedFiles, nil
}
