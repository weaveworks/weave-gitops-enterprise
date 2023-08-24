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

const WORKINGDIR = "/tmp/bootstrap-flux"

// GetRepoUrl get the default repo url for flux installation
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

// GetRepoBranch get the branch for flux installation
func GetRepoBranch() (string, error) {
	var runner runner.CLIRunner

	repoBranch, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.ref.branch}\"")
	if err != nil {
		return "", err
	}

	repoBranchParsed := string(repoBranch[1 : len(repoBranch)-1])

	return repoBranchParsed, nil
}

// GetRepoPath get the path for flux installation
func GetRepoPath() (string, error) {
	var runner runner.CLIRunner

	repoPath, err := runner.Run("kubectl", "get", "kustomization", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.path}\"")
	if err != nil {
		return "", err
	}

	repoPathParsed := strings.TrimPrefix(string(repoPath[1:len(repoPath)-1]), "./")

	return repoPathParsed, nil
}

// CloneRepo shallow clones the user repo's branch under temp
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

	out, err := runner.Run("git", "clone", repoUrlParsed, WORKINGDIR, "--depth", "1", "-b", repoBranchParsed)
	if err != nil {
		return "", fmt.Errorf("%s%s", err.Error(), string(out))
	}

	return repoPathParsed, nil
}

// CreateFileToRepo create a file and add to the repo
func CreateFileToRepo(filename string, filecontent string, path string, commitmsg string) error {
	repo, err := git.PlainOpen(WORKINGDIR)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	filePath := filepath.Join(WORKINGDIR, path, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.WriteString(filecontent)
	if err != nil {
		return err
	}

	_, err = worktree.Add(filepath.Join(path, filename))
	if err != nil {
		return err
	}

	_, err = worktree.Commit(commitmsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Flux Bootstrap CLI",
			Email: "bootstrap@weave.works",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CleanupRepo delete the temp repo
func CleanupRepo() error {
	return os.RemoveAll(WORKINGDIR)
}
