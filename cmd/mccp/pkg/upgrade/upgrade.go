package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaveworks/pctl/pkg/bootstrap"
	"github.com/weaveworks/pctl/pkg/catalog"
	"github.com/weaveworks/pctl/pkg/client"
	"github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/pctl/pkg/runner"
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
}

func Upgrade(params UpgradeParams) error {
	installationDirectory, err := addProfile(params)
	if err != nil {
		return err
	}

	if err := createPullRequest(params, installationDirectory); err != nil {
		return err

	}
	return nil
}

func removeProfile() error {
	return nil
}

func addProfile(params UpgradeParams) (string, error) {
	var (
		err           error
		catalogClient *client.Client
		profilePath   string
		catalogName   string
		profileName   string
		version       = "latest"
	)

	url := params.ProfileRepoURL
	branch := params.ProfileBranch
	subName := params.Name
	namespace := params.Namespace
	configMap := params.ConfigMap
	dir := params.Out
	path := params.ProfilePath
	message := params.CommitMessage

	parts := strings.Split(profilePath, "/")

	if len(parts) == 3 {
		version = parts[2]
	}
	catalogName, profileName = parts[0], parts[1]

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
