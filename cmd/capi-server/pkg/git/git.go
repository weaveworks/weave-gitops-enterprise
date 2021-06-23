package git

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var DefaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 20 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

type Provider interface {
	WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req WriteFilesToBranchAndCreatePullRequestRequest) (*WriteFilesToBranchAndCreatePullRequestResponse, error)
}

type GitProviderService struct {
}

func NewGitProviderService() *GitProviderService {
	return &GitProviderService{}
}

type GitProvider struct {
	Token    string
	Type     string
	Hostname string
}

type WriteFilesToBranchAndCreatePullRequestRequest struct {
	GitProvider   GitProvider
	RepositoryURL string
	HeadBranch    string
	BaseBranch    string
	Title         string
	Description   string
	CommitMessage string
	Files         []gitprovider.CommitFile
}

type WriteFilesToBranchAndCreatePullRequestResponse struct {
	WebURL string
}

// WriteFilesToBranchAndCreatePullRequest writes a set of provided files
// to a new branch and creates a new pull request for that branch.
// It returns the URL of the pull request.
func (s *GitProviderService) WriteFilesToBranchAndCreatePullRequest(ctx context.Context,
	req WriteFilesToBranchAndCreatePullRequestRequest) (*WriteFilesToBranchAndCreatePullRequestResponse, error) {
	repo, err := s.getRepository(ctx, req.GitProvider, req.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo: %w", err)
	}

	if err := s.writeFilesToBranch(ctx, writeFilesToBranchRequest{
		Repository:    repo,
		HeadBranch:    req.HeadBranch,
		BaseBranch:    req.BaseBranch,
		CommitMessage: req.CommitMessage,
		Files:         req.Files,
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

func (s *GitProviderService) getRepository(ctx context.Context, gp GitProvider, url string) (gitprovider.OrgRepository, error) {
	c, err := getGitProviderClient(gp)
	if err != nil {
		return nil, fmt.Errorf("unable to get a git provider client for %q: %w", gp.Type, err)
	}

	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse repository URL %q: %w", url, err)
	}

	var repo gitprovider.OrgRepository
	err = retry.OnError(DefaultBackoff,
		func(err error) bool {
			if errors.Is(err, gitprovider.ErrNotFound) {
				return true
			}
			return false
		}, func() error {
			var err error
			repo, err = c.OrgRepositories().Get(ctx, *ref)
			if err != nil {
				log.Warn("Retrying getting the repository")
				return err
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("unable to get repository %q: %w", url, err)
	}

	return repo, nil
}

type writeFilesToBranchRequest struct {
	Repository    gitprovider.OrgRepository
	HeadBranch    string
	BaseBranch    string
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
				log.Warn("Retrying getting the repository")
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

	commit, err := req.Repository.Commits().Create(ctx, req.HeadBranch, req.CommitMessage, req.Files)
	if err != nil {
		return fmt.Errorf("unable to commit changes to %q: %w", req.HeadBranch, err)
	}
	log.WithFields(log.Fields{
		"sha":    commit.Get().Sha,
		"branch": req.HeadBranch,
	}).Info("Files committed")

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
	log.WithFields(log.Fields{
		"pull_request_web_url": pr.Get().WebURL,
	}).Info("Created pull request")

	return &createPullRequestResponse{
		WebURL: pr.Get().WebURL,
	}, nil
}

func getGitProviderClient(gpi GitProvider) (gitprovider.Client, error) {
	var client gitprovider.Client
	var err error

	switch gpi.Type {
	case "github":
		if gpi.Hostname != "github.com" {
			client, err = github.NewClient(
				github.WithOAuth2Token(gpi.Token),
				github.WithDomain(gpi.Hostname),
			)
		} else {
			client, err = github.NewClient(
				github.WithOAuth2Token(gpi.Token),
			)
		}
		if err != nil {
			return nil, err
		}
	case "gitlab":
		if gpi.Hostname != "gitlab.com" {
			client, err = gitlab.NewClient(gpi.Token, "", gitlab.WithDomain(gpi.Hostname), gitlab.WithConditionalRequests(true))
		} else {
			client, err = gitlab.NewClient(gpi.Token, "", gitlab.WithConditionalRequests(true))
		}
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("the Git provider %q is not supported", gpi.Type)
	}
	return client, err
}
