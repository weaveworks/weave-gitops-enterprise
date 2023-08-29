package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	workingDir      = "/tmp/bootstrap-flux"
	fluxGitUserName = "Flux Bootstrap CLI"
	fluxGitEmail    = "bootstrap@weave.works"
)

// GetRepoUrl get the default repo url for flux installation (flux-system) GitRepository.
func GetRepoUrl() (string, error) {
	var runner runner.CLIRunner
	repoUrl, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.url}\"")
	if err != nil {
		return "", err
	}

	repoUrlParsed := string(repoUrl[1 : len(repoUrl)-1])

	if strings.Contains(repoUrlParsed, "ssh://") {
		repoUrlParsed = strings.TrimPrefix(repoUrlParsed, "ssh://")
		repoUrlParsed = strings.Replace(repoUrlParsed, "/", ":", 1)
	}
	return repoUrlParsed, nil
}

// GetRepoBranch get the branch for flux installation (flux-system) GitRepository.
func GetRepoBranch() (string, error) {
	var runner runner.CLIRunner

	repoBranch, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.ref.branch}\"")
	if err != nil {
		return "", err
	}

	repoBranchParsed := string(repoBranch[1 : len(repoBranch)-1])

	return repoBranchParsed, nil
}

// GetRepoPath get the path for flux installation (flux-system) Kustomization.
func GetRepoPath() (string, error) {
	var runner runner.CLIRunner

	repoPath, err := runner.Run("kubectl", "get", "kustomization", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.path}\"")
	if err != nil {
		return "", err
	}

	repoPathParsed := strings.TrimPrefix(string(repoPath[1:len(repoPath)-1]), "./")

	return repoPathParsed, nil
}

// CloneRepo shallow clones the user repo's branch under temp and returns the current path.
func CloneRepo() (string, error) {
	if err := CleanupRepo(); err != nil {
		return "", err
	}

	var runner runner.CLIRunner

	repoUrlParsed, err := GetRepoUrl()
	if err != nil {
		return "", err
	}

	repoBranchParsed, err := GetRepoBranch()
	if err != nil {
		return "", err
	}

	repoPathParsed, err := GetRepoPath()
	if err != nil {
		return "", err
	}

	out, err := runner.Run("git", "clone", repoUrlParsed, workingDir, "--depth", "1", "-b", repoBranchParsed)
	if err != nil {
		return "", fmt.Errorf("%s%s", err.Error(), string(out))
	}

	return repoPathParsed, nil
}

// CreateFileToRepo create a file and add to the repo.
func CreateFileToRepo(filename string, filecontent string, path string, commitmsg string) error {
	repo, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	filePath := filepath.Join(workingDir, path, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	if _, err := file.WriteString(filecontent); err != nil {
		return err
	}

	if _, err := worktree.Add(filepath.Join(path, filename)); err != nil {
		return err
	}

	if _, err := worktree.Commit(commitmsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  fluxGitUserName,
			Email: fluxGitEmail,
			When:  time.Now(),
		},
	}); err != nil {
		return err
	}

	if err := repo.Push(&git.PushOptions{}); err != nil {
		return err
	}

	return nil
}

// CleanupRepo delete the temp repo.
func CleanupRepo() error {
	return os.RemoveAll(workingDir)
}
