package upgrade

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/pctl/pkg/bootstrap"
	"github.com/weaveworks/pctl/pkg/catalog"
	"github.com/weaveworks/pctl/pkg/client"
	"github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/pctl/pkg/runner"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type UpgradeParams struct {
	RepositoryURL  string
	Remote         string
	HeadBranch     string
	BaseBranch     string
	Title          string
	Description    string
	CommitMessage  string
	Name           string
	Namespace      string
	ProfileBranch  string
	ConfigMap      string
	Out            string
	ProfileRepoURL string
	ProfilePath    string
	GitRepository  string
	Args           []string
}

func Upgrade(w io.Writer) error {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)
	kubeClientConfig := config.GetConfigOrDie()
	kubeClient, err := client.New(kubeClientConfig, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	err = PreFlightCheck(kubeClient)
	if err != nil {
		return err
	}

	err = removeWEGO(kubeClient)
	if err != nil {
		return err
	}

	repoURL, err := getRepoURL()
	if err != nil {
		return err
	}

	gitRepository, err := getGitRepo()
	if err != nil {
		return err
	}

	params := UpgradeParams{
		RepositoryURL:  repoURL,
		Remote:         "origin",
		HeadBranch:     "tier-upgrade-enterprise",
		BaseBranch:     "main",
		Title:          "Upgrade to WGE",
		Description:    "Upgrade to WGE",
		CommitMessage:  "Upgrade to WGE",
		Name:           "wge-profile",
		Namespace:      "wego-system",
		ProfileBranch:  "main",
		ConfigMap:      "",
		Out:            ".",
		ProfileRepoURL: "https://github.com/weaveworks/weave-gitops-enterprise-profiles",
		ProfilePath:    ".",
		GitRepository:  gitRepository,
	}

	installationDirectory, err := addProfile(params)
	if err != nil {
		return err
	}

	if err := createPullRequest(params, installationDirectory); err != nil {
		return err

	}

	fmt.Fprintf(w, "Upgrade pull request created\n")

	return nil
}

func getRepoURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = "."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return stdout.String(), nil
}

func getGitRepo() (string, error) {
	cmd := exec.Command("basename", "`git rev-parse --show-toplevel`")
	cmd.Dir = "."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return stdout.String(), nil
}

func PreFlightCheck(c client.Client) error {
	secs := &v1.SecretList{}
	name := "weave-gitops-enterprise-credentials"
	foundEntitlement := false

	c.List(context.Background(), secs)

	for _, item := range secs.Items {
		if item.Name == name {
			foundEntitlement = true
			break
		}
	}

	if !foundEntitlement {
		return fmt.Errorf("failed to get entitlement: %v", name)
	}

	return nil
}

func removeWEGO(c client.Client) error {
	cmdItems := []string{"kubectl", "delete", "--all", "services", "--namespace", "wego-system"}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete wego services: %v", err)
	}

	cmdItems = []string{"kubectl", "delete", "--all", "deployments", "--namespace", "wego-system"}
	cmd = exec.Command(cmdItems[0], cmdItems[1:]...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to wego deployments: %v", err)
	}

	return nil
}

func addProfile(params UpgradeParams) (string, error) {
	var (
		err           error
		catalogClient catalog.CatalogClient
		catalogName   string
		profileName   string
		version       = "latest"
	)

	url := params.ProfileRepoURL

	catalogClient, err = buildCatalogClient()
	if err != nil {
		return "", err
	}

	branch := params.ProfileBranch
	subName := params.Name
	namespace := params.Namespace
	configMap := params.ConfigMap
	dir := params.Out
	path := params.ProfilePath
	message := params.CommitMessage

	r := &runner.CLIRunner{}
	g := git.NewCLIGit(git.CLIGitConfig{
		Message: message,
	}, r)

	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(params.GitRepository)
	if err != nil {
		return "", err
	}

	installationDirectory := filepath.Join(dir, subName)
	installer := install.NewInstaller(install.Config{
		GitClient:        g,
		RootDir:          installationDirectory,
		GitRepoNamespace: gitRepoNamespace,
		GitRepoName:      gitRepoName,
	})

	cfg := catalog.InstallConfig{
		Clients: catalog.Clients{
			CatalogClient: catalogClient,
			Installer:     installer,
		},
		Profile: catalog.Profile{
			ProfileConfig: catalog.ProfileConfig{
				CatalogName:   catalogName,
				ConfigMap:     configMap,
				Namespace:     namespace,
				Path:          path,
				ProfileBranch: branch,
				ProfileName:   profileName,
				SubName:       subName,
				URL:           url,
				Version:       version,
			},
			GitRepoConfig: catalog.GitRepoConfig{
				Namespace: gitRepoNamespace,
				Name:      gitRepoName,
			},
		},
	}
	manager := &catalog.Manager{}
	err = manager.Install(cfg)

	return installationDirectory, err
}

func getGitRepositoryNamespaceAndName(gitRepository string) (string, string, error) {
	if gitRepository != "" {
		split := strings.Split(gitRepository, "/")
		if len(split) != 2 {
			return "", "", fmt.Errorf("git-repository must in format <namespace>/<name>; was: %s", gitRepository)
		}
		return split[0], split[1], nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch current working directory: %w", err)
	}
	config, err := bootstrap.GetConfig(wd)
	if err == nil && config != nil {
		return config.GitRepository.Namespace, config.GitRepository.Name, nil
	}
	return "", "", fmt.Errorf("flux git repository not provided, please provide the --git-repository flag or use the pctl bootstrap functionality")
}

func buildCatalogClient() (catalog.CatalogClient, error) {
	home := homedir.HomeDir()
	options := client.ServiceOptions{
		KubeconfigPath: filepath.Join(home, ".kube", "config"),
		Namespace:      "profiles-catalog-namespace",
		ServiceName:    "profiles-catalog-name",
		ServicePort:    "8000",
	}
	return client.NewFromOptions(options)
}

func createPullRequest(params UpgradeParams, installationDirectory string) error {
	branch := params.HeadBranch
	repo := params.RepositoryURL
	base := params.BaseBranch
	remote := params.Remote
	directory := params.Out
	message := params.CommitMessage

	r := &runner.CLIRunner{}
	g := git.NewCLIGit(git.CLIGitConfig{
		Directory: directory,
		Branch:    branch,
		Remote:    remote,
		Base:      base,
		Message:   message,
	}, r)
	scmClient, err := git.NewClient(git.SCMConfig{
		Branch: branch,
		Base:   base,
		Repo:   repo,
	})
	if err != nil {
		return fmt.Errorf("failed to create scm client: %w", err)
	}
	return catalog.CreatePullRequest(scmClient, g, branch, installationDirectory)
}
