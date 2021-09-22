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
	git_utils "github.com/weaveworks/weave-gitops-enterprise/pkg/utilities/git"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type UpgradeValues struct {
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
	config := config.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	entitlement, err := PreFlightCheck(clientset)
	if err != nil {
		return err
	}

	repoURL, err := getRepoURL()
	if err != nil {
		return err
	}
	log.Infof("Found repo url as: %v", repoURL)

	githubRepoPath, err := getGithubRepoPath(repoURL)
	if err != nil {
		return err
	}

	gitRepositoryResource := "wego-system/" + strings.TrimSuffix(filepath.Base(repoURL), ".git")

	upgradeValues := UpgradeValues{
		RepositoryURL:  githubRepoPath,
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
		ProfileRepoURL: "git@github.com:weaveworks/weave-gitops-enterprise-profiles.git",
		ProfilePath:    ".",
		GitRepository:  gitRepositoryResource,
	}

	log.Infof("Using values %+v", upgradeValues)

	key := entitlement.Data["deploy-key"]
	localRepo, err := git_utils.CloneToTempDir("/tmp", upgradeValues.ProfileRepoURL, upgradeValues.ProfileBranch, key)
	if err != nil {
		return err
	}

	upgradeValues.ProfileRepoURL = localRepo.WorktreeDir()

	installationDirectory, err := addProfile(upgradeValues)
	if err != nil {
		return err
	}

	if err := createPullRequest(upgradeValues, installationDirectory); err != nil {
		return err

	}

	fmt.Fprintf(w, "Upgrade pull request created\n")

	return nil
}

func getGithubRepoPath(url string) (string, error) {
	repoEndpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	return strings.Trim(strings.TrimSuffix(strings.TrimSpace(repoEndpoint.Path), ".git"), "/"), nil
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
	return strings.TrimSpace(stdout.String()), nil
}

func PreFlightCheck(clientset kubernetes.Interface) (*v1.Secret, error) {
	log.Info("Checking if entitlement exists...")
	var entitlement *v1.Secret

	entitlement, err := clientset.CoreV1().Secrets("wego-system").Get(context.Background(), "weave-gitops-enterprise-credentials", metav1.GetOptions{})
	if err != nil {
		return entitlement, fmt.Errorf("failed to get entitlement: %v", err)
	}

	return entitlement, nil
}

func addProfile(values UpgradeValues) (string, error) {
	var (
		err           error
		catalogClient *client.Client
		catalogName   string
		profileName   string
		version       = "latest"
	)

	url := values.ProfileRepoURL

	catalogClient, err = buildCatalogClient()
	if err != nil {
		return "", err
	}

	branch := values.ProfileBranch
	subName := values.Name
	namespace := values.Namespace
	configMap := values.ConfigMap
	dir := values.Out
	path := values.ProfilePath
	message := values.CommitMessage

	r := &runner.CLIRunner{}
	g := git.NewCLIGit(git.CLIGitConfig{
		Message: message,
	}, r)

	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(values.GitRepository)
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

func buildCatalogClient() (*client.Client, error) {
	home := homedir.HomeDir()
	options := client.ServiceOptions{
		KubeconfigPath: filepath.Join(home, ".kube", "config"),
		Namespace:      "profiles-catalog-namespace",
		ServiceName:    "profiles-catalog-name",
		ServicePort:    "8000",
	}
	return client.NewFromOptions(options)
}

func createPullRequest(values UpgradeValues, installationDirectory string) error {
	branch := values.HeadBranch
	repo := values.RepositoryURL
	base := values.BaseBranch
	remote := values.Remote
	directory := values.Out
	message := values.CommitMessage

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
