package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RepoCleanupMsg = "Cleaning up repo ..."
)

const (
	workingDir      = "/tmp/bootstrap-flux"
	fluxGitUserName = "Flux Bootstrap CLI"
	fluxGitEmail    = "bootstrap@weave.works"
)

func getGitRepository(repoName string, namespace string, opts config.Options) (*v1beta2.GitRepository, error) {
	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return nil, err
	}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Get the GitRepository custom resource
	gitRepo := &v1beta2.GitRepository{}
	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      repoName,
	}, gitRepo); err != nil {
		return nil, err
	}

	return gitRepo, nil
}

// GetRepoUrl get the default repo url for flux installation (flux-system) GitRepository.
func GetRepoUrl(repoName string, namespace string, opts config.Options) (string, error) {
	gitRepo, err := getGitRepository(repoName, namespace, opts)
	if err != nil {
		return "", err
	}

	// Parse the URL
	repoUrlParsed := gitRepo.Spec.URL
	if strings.Contains(repoUrlParsed, "ssh://") {
		repoUrlParsed = strings.TrimPrefix(repoUrlParsed, "ssh://")
		repoUrlParsed = strings.Replace(repoUrlParsed, "/", ":", 1)
	}

	return repoUrlParsed, nil
}

// GetRepoBranch get the branch for flux installation (flux-system) GitRepository.
func GetRepoBranch(repoName string, namespace string, opts config.Options) (string, error) {
	gitRepo, err := getGitRepository(repoName, namespace, opts)
	if err != nil {
		return "", err
	}

	// Extract the branch
	return gitRepo.Spec.Reference.Branch, nil
}

// GetRepoPath get the path for flux installation (flux-system) Kustomization.
func GetRepoPath(repoName string, namespace string, opts config.Options) (string, error) {

	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return "", err
	}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	kustomization := &kustomizev1.Kustomization{}

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      repoName,
	}, kustomization); err != nil {
		return "", err
	}

	repoPath := kustomization.Spec.Path

	repoPathParsed := strings.TrimPrefix(string(repoPath[1:len(repoPath)-1]), "./")

	return repoPathParsed, nil
}

// CloneRepo shallow clones the user repo's branch under temp and returns the current path.
func CloneRepo(repoName string, namespace string) (string, error) {
	if err := CleanupRepo(); err != nil {
		return "", err
	}

	var runner runner.CLIRunner

	repoUrlParsed, err := GetRepoUrl(repoName, namespace, config.Options{})
	if err != nil {
		return "", err
	}

	repoBranchParsed, err := GetRepoBranch(repoName, namespace, config.Options{})
	if err != nil {
		return "", err
	}

	repoPathParsed, err := GetRepoPath(repoName, namespace, config.Options{})
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
