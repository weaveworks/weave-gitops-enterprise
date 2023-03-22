package gitfakes

import (
	"context"
	"sort"
	"strings"

	csgit "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
)

func NewFakeGitProvider(url string, repo *csgit.GitRepo, err error, originalFilesPaths []string, prs []*git.PullRequest) csgit.Provider {
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
	repo           *csgit.GitRepo
	err            error
	CommittedFiles []git.CommitFile
	OriginalFiles  []string
	pullRequests   []*git.PullRequest
}

func (p *FakeGitProvider) WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req csgit.WriteFilesToBranchAndCreatePullRequestRequest) (*csgit.WriteFilesToBranchAndCreatePullRequestResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.CommittedFiles = append(p.CommittedFiles, req.Files...)
	return &csgit.WriteFilesToBranchAndCreatePullRequestResponse{WebURL: p.url}, nil
}

func (p *FakeGitProvider) GetRepository(ctx context.Context, gp csgit.GitProvider, url string) (*git.Repository, error) {
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
			Path:    f.Path,
			Content: content,
		})
	}
	sortCommitFiles(committedFiles)
	return committedFiles
}

func (p *FakeGitProvider) GetTreeList(ctx context.Context, gp csgit.GitProvider, repoUrl string, sha string, path string, recursive bool) ([]*git.TreeEntry, error) {
	if p.err != nil {
		return nil, p.err
	}

	var treeEntries []*git.TreeEntry
	for _, filePath := range p.OriginalFiles {
		if path == "" || (path != "" && strings.HasPrefix(filePath, path)) {
			treeEntries = append(treeEntries, &git.TreeEntry{
				Path: filePath,
				Type: "",
				Size: 0,
				SHA:  "",
				Link: "",
			})
		}

	}
	return treeEntries, nil
}

func (p *FakeGitProvider) ListPullRequests(ctx context.Context, gp csgit.GitProvider, url string) ([]*git.PullRequest, error) {
	return p.pullRequests, nil
}

func NewPullRequest(id int, title string, description string, url string, merged bool, sourceBranch string) *git.PullRequest {
	return &git.PullRequest{
		Title:       title,
		Description: description,
		Link:        url,
		Merged:      merged,
	}
}

func sortCommitFiles(files []*capiv1_protos.CommitFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}
