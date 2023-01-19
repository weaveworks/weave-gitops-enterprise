package gitfakes

import (
	"context"
	"sort"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func NewFakeGitProvider(url string, repo *git.GitRepo, err error, originalFilesPaths []string, prs []gitprovider.PullRequest) git.Provider {
	return &FakeGitProvider{
		url:           url,
		repo:          repo,
		err:           err,
		OriginalFiles: originalFilesPaths,
		pullRequests:  prs,
	}
}

type FakeGitProvider struct {
	url            string
	repo           *git.GitRepo
	err            error
	CommittedFiles []gitprovider.CommitFile
	OriginalFiles  []string
	pullRequests   []gitprovider.PullRequest
}

func (p *FakeGitProvider) WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req git.WriteFilesToBranchAndCreatePullRequestRequest) (*git.WriteFilesToBranchAndCreatePullRequestResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.CommittedFiles = append(p.CommittedFiles, req.Files...)
	return &git.WriteFilesToBranchAndCreatePullRequestResponse{WebURL: p.url}, nil
}

func (p *FakeGitProvider) CloneRepoToTempDir(req git.CloneRepoToTempDirRequest) (*git.CloneRepoToTempDirResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &git.CloneRepoToTempDirResponse{Repo: p.repo}, nil
}

func (p *FakeGitProvider) GetRepository(ctx context.Context, gp git.GitProvider, url string) (gitprovider.OrgRepository, error) {
	if p.err != nil {
		return nil, p.err
	}
	return nil, nil
}

func (p *FakeGitProvider) GetCommittedFiles() []*capiv1_protos.CommitFile {
	var committedFiles []*capiv1_protos.CommitFile
	for _, f := range p.CommittedFiles {
		content := ""
		if f.Content != nil {
			content = *f.Content
		}

		committedFiles = append(committedFiles, &capiv1_protos.CommitFile{
			Path:    *f.Path,
			Content: content,
		})
	}
	sortCommitFiles(committedFiles)
	return committedFiles
}

func (p *FakeGitProvider) GetTreeList(ctx context.Context, gp git.GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*gitprovider.TreeEntry, error) {
	if p.err != nil {
		return nil, p.err
	}

	var treeEntries []*gitprovider.TreeEntry
	for _, filePath := range p.OriginalFiles {
		if path == "" || (path != "" && strings.HasPrefix(filePath, path)) {
			treeEntries = append(treeEntries, &gitprovider.TreeEntry{
				Path:    filePath,
				Mode:    "",
				Type:    "",
				Size:    0,
				SHA:     "",
				Content: "",
				URL:     "",
			})
		}

	}
	return treeEntries, nil
}

func (p *FakeGitProvider) ListPullRequests(ctx context.Context, gp git.GitProvider, url string) ([]gitprovider.PullRequest, error) {
	return p.pullRequests, nil
}

func NewPullRequest(id int, title string, description string, url string, merged bool, sourceBranch string) gitprovider.PullRequest {
	return &pullrequest{
		id:           id,
		title:        title,
		description:  description,
		url:          url,
		merged:       merged,
		sourceBranch: sourceBranch,
	}
}

type pullrequest struct {
	id           int
	title        string
	description  string
	url          string
	merged       bool
	sourceBranch string
}

func (pr *pullrequest) Get() gitprovider.PullRequestInfo {
	return gitprovider.PullRequestInfo{
		Title:        pr.title,
		Description:  pr.description,
		WebURL:       pr.url,
		Number:       pr.id,
		Merged:       pr.merged,
		SourceBranch: pr.sourceBranch,
	}
}

func (pr *pullrequest) APIObject() interface{} {
	return &pr
}

func sortCommitFiles(files []*capiv1_protos.CommitFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}
