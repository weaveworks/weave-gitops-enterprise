package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RepoCleanupMsg = "Cleaning up repo ..."
)

const (
	workingDir      = "/tmp/bootstrap-flux"
	fluxGitUserName = "Flux Bootstrap CLI"
	fluxGitEmail    = "bootstrap@weave.works"
)

// getGitRepositoryObject get the default source git repository object to be used in cloning
func getGitRepositoryObject(client k8s_client.Client, repoName string, namespace string) (*sourcev1.GitRepository, error) {
	gitRepo := &sourcev1.GitRepository{}

	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Namespace: namespace,
		Name:      repoName,
	}, gitRepo); err != nil {
		return nil, err
	}

	return gitRepo, nil
}

// getRepoUrl get the default repo url for flux installation (flux-system) GitRepository.
func getRepoUrl(gitRepo *sourcev1.GitRepository) string {
	return gitRepo.Spec.URL
}

// getRepoBranch get the branch for flux installation (flux-system) GitRepository.
func getRepoBranch(gitRepo *sourcev1.GitRepository) string {
	return gitRepo.Spec.Reference.Branch
}

// getRepoPath get the path for flux installation (flux-system) Kustomization.
func getRepoPath(client k8s_client.Client, repoName string, namespace string) (string, error) {
	kustomization := &kustomizev1.Kustomization{}

	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Namespace: namespace,
		Name:      repoName,
	}, kustomization); err != nil {
		return "", err
	}

	return kustomization.Spec.Path, nil
}

// CloneRepo shallow clones the user repo's branch under temp and returns the current path.
func CloneRepo(client k8s_client.Client, repoName string, namespace string) (string, error) {
	if err := CleanupRepo(); err != nil {
		return "", err
	}

	gitRepo, err := getGitRepositoryObject(client, repoName, namespace)
	if err != nil {
		return "", err
	}

	repoUrl := getRepoUrl(gitRepo)
	repoBranch := getRepoBranch(gitRepo)

	repoPath, err := getRepoPath(client, repoName, namespace)
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:           repoUrl,
		ReferenceName: plumbing.NewBranchReferenceName(repoBranch),
		SingleBranch:  true,
		Depth:         1,
		Progress:      nil,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository %s: %w", repoName, err)
	}

	return repoPath, nil
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
