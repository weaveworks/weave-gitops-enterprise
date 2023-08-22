package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/manifoldco/promptui"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PromptContent struct {
	ErrorMsg     string
	Label        string
	DefaultValue string
}

func GetPromptStringInput(pc PromptContent) (string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New(pc.ErrorMsg)
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    pc.Label,
		Validate: validate,
		Default:  pc.DefaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", CheckIfError(err, "Prompt failed")
	}

	return result, nil
}

func GetPromptPasswordInput(pc PromptContent) (string, error) {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    pc.Label,
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()

	if err != nil {
		return "", CheckIfError(err, "Prompt failed")
	}

	return result, nil
}

func GetPromptSelect(pc PromptContent, items []string) (string, error) {
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.Select{
			Label: pc.Label,
			Items: items,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		return "", CheckIfError(err, "Prompt failed")
	}

	fmt.Printf("Selected: %s\n", result)

	return result, nil
}

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	// Path to the kubeconfig file. This is typically located at "~/.kube/config".
	// Obtain the user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, CheckIfError(err)
	}

	// Construct the full path to the kubeconfig file.
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, CheckIfError(err)
	}

	// Create a new Kubernetes client using the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, CheckIfError(err)
	}

	return clientset, nil
}

func GetSecret(secretNamespace, secretName string) (*corev1.Secret, error) {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		return nil, CheckIfError(err)
	}

	// Fetch the secret from the Kubernetes cluster.
	secret, err := clientset.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, CheckIfError(err)
	}

	return secret, nil
}

func CreateSecret(secretName string, secretNamespace string, secretData map[string][]byte) error {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		return CheckIfError(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: secretData,
	}

	_, err = clientset.CoreV1().Secrets(secretNamespace).Create(context.TODO(), secret, v1.CreateOptions{
		TypeMeta: secret.TypeMeta,
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return CheckIfError(err)
	}
	return nil
}

const WORKINGDIR = "/tmp/bootstrap-flux"

func CloneRepo() (string, error) {

	err := CleanupRepo()
	if err != nil {
		return "", CheckIfError(err)
	}

	var runner runner.CLIRunner
	repoUrl, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.url}\"")
	if err != nil {
		return "", CheckIfError(err)
	}

	repoUrlParsed := string(repoUrl[1 : len(repoUrl)-1])

	if strings.Contains(repoUrlParsed, "ssh://") {
		repoUrlParsed = strings.TrimPrefix(repoUrlParsed, "ssh://")
		repoUrlParsed = strings.Replace(repoUrlParsed, "/", ":", 1)
	}

	repoBranch, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.ref.branch}\"")
	if err != nil {
		return "", CheckIfError(err)
	}

	repoBranchParsed := string(repoBranch[1 : len(repoBranch)-1])

	repoPath, err := runner.Run("kubectl", "get", "kustomization", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.path}\"")
	if err != nil {
		return "", CheckIfError(err)
	}

	repoPathParsed := strings.TrimPrefix(string(repoPath[1:len(repoPath)-1]), "./")

	out, err := runner.Run("git", "clone", repoUrlParsed, WORKINGDIR, "--depth", "1", "-b", repoBranchParsed)
	if err != nil {
		return "", CheckIfError(err, string(out))
	}

	return repoPathParsed, nil
}

func CreateFileToRepo(filename string, filecontent string, path string, commitmsg string) error {

	repo, err := git.PlainOpen(WORKINGDIR)
	if err != nil {
		return CheckIfError(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return CheckIfError(err)
	}

	filePath := filepath.Join(WORKINGDIR, path, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return CheckIfError(err)
	}

	defer file.Close()
	_, err = file.WriteString(filecontent)
	if err != nil {
		return CheckIfError(err)
	}

	_, err = worktree.Add(filepath.Join(path, filename))
	if err != nil {
		return CheckIfError(err)
	}

	_, err = worktree.Commit(commitmsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Flux Bootstrap CLI",
			Email: "bootstrap@weave.works",
			When:  time.Now(),
		},
	})
	if err != nil {
		return CheckIfError(err)
	}

	err = repo.Push(&git.PushOptions{})
	if err != nil {
		return CheckIfError(err)
	}

	return nil
}

func CheckIfError(err error, extramsg ...string) error {
	if err == nil {
		return nil
	}
	if len(extramsg) > 0 {
		return fmt.Errorf("\x1b[31;1m%s\x1b[0m", fmt.Sprintf("%s\n%s", err.Error(), extramsg[0]))
	}
	return fmt.Errorf("\x1b[31;1m%s\x1b[0m", err.Error())
}

func CleanupRepo() error {
	err := os.RemoveAll(WORKINGDIR)
	return CheckIfError(err)
}

func ReconcileFlux(helmReleaseName ...string) error {

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "source", "git", "flux-system")
	if err != nil {
		return CheckIfError(err, string(out))
	}
	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		CheckIfError(err, string(out))
	}
	if len(helmReleaseName) > 0 {
		out, err = runner.Run("flux", "reconcile", "helmrelease", helmReleaseName[0])
		if err != nil {
			CheckIfError(err, string(out))
		}
	}
	return nil
}
