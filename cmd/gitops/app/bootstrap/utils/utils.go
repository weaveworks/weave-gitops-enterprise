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

func GetPromptStringInput(pc PromptContent) string {
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
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func GetPromptPasswordInput(pc PromptContent) string {
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
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func GetPromptSelect(pc PromptContent, items []string) string {
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
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected: %s\n", result)

	return result
}

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	// Path to the kubeconfig file. This is typically located at "~/.kube/config".
	// Obtain the user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Construct the full path to the kubeconfig file.
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create a new Kubernetes client using the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetSecret(secretNamespace, secretName string) (*corev1.Secret, error) {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	// Fetch the secret from the Kubernetes cluster.
	secret, err := clientset.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func CreateSecret(secretName string, secretNamespace string, secretData map[string][]byte) {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		panic(err.Error())
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
		panic(err.Error())
	}
}

const WORKINGDIR = "/tmp/bootstrap-flux"

func CloneRepo() (string, error) {

	CleanupRepo()

	var runner runner.CLIRunner
	repoUrl, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.url}\"")
	CheckIfError(err)

	repoUrlParsed := string(repoUrl[1 : len(repoUrl)-1])

	if strings.Contains(repoUrlParsed, "ssh://") {
		repoUrlParsed = strings.TrimPrefix(repoUrlParsed, "ssh://")
		repoUrlParsed = strings.Replace(repoUrlParsed, "/", ":", 1)
	}

	repoBranch, err := runner.Run("kubectl", "get", "gitrepository", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.ref.branch}\"")
	CheckIfError(err)

	repoBranchParsed := string(repoBranch[1 : len(repoBranch)-1])

	repoPath, err := runner.Run("kubectl", "get", "kustomization", "flux-system", "-n", "flux-system", "-o", "jsonpath=\"{.spec.path}\"")
	CheckIfError(err)

	repoPathParsed := strings.TrimPrefix(string(repoPath[1:len(repoPath)-1]), "./")

	out, err := runner.Run("git", "clone", repoUrlParsed, WORKINGDIR, "--depth", "1", "-b", repoBranchParsed)
	if err != nil {
		fmt.Printf("An error occured cloning repo %s\n%v", repoUrlParsed, string(out))
		os.Exit(1)
	}

	return repoPathParsed, nil
}

func CreateFileToRepo(filename string, filecontent string, path string, commitmsg string) error {

	repo, err := git.PlainOpen(WORKINGDIR)
	CheckIfError(err)

	worktree, err := repo.Worktree()
	CheckIfError(err)

	filePath := filepath.Join(WORKINGDIR, path, filename)

	file, err := os.Create(filePath)
	CheckIfError(err)

	defer file.Close()
	_, err = file.WriteString(filecontent)
	CheckIfError(err)

	_, err = worktree.Add(filepath.Join(path, filename))
	CheckIfError(err)

	_, err = worktree.Commit(commitmsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Flux Bootstrap CLI",
			Email: "bootstrap@weave.works",
			When:  time.Now(),
		},
	})
	CheckIfError(err)

	return nil
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func CleanupRepo() {
	err := os.RemoveAll(WORKINGDIR)
	CheckIfError(err)
}

// func CreateConfigMap(Name string, Namespace string, Data map[string]string) {
// 	// Create a new Kubernetes client using the config.
// 	clientset, err := getKubernetesClient()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	configMap := &corev1.ConfigMap{
// 		ObjectMeta: v1.ObjectMeta{
// 			Name:      Name,
// 			Namespace: Namespace,
// 		},
// 		Data: Data,
// 	}

// 	_, err = clientset.CoreV1().ConfigMaps(Namespace).Create(context.TODO(), configMap, v1.CreateOptions{
// 		TypeMeta: configMap.TypeMeta,
// 	})
// 	if err != nil && !strings.Contains(err.Error(), "already exists") {
// 		panic(err.Error())
// 	}
// }
