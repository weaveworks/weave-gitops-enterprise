package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"
)

const AddCommitMessage = "Add profile manifests"

// Add installs an available profile in a cluster's namespace by appending a HelmRelease to the profile manifest in the config repo,
// provided that such a HelmRelease does not exist with the same profile name and version in the same namespace and cluster.
func (s *ProfilesSvc) Add(ctx context.Context, r ProfilesRetriever, gitProvider gitproviders.GitProvider, opts Options) error {
	configRepoURL, err := gitproviders.NewRepoURL(opts.ConfigRepo)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoURL)
	if err != nil {
		return fmt.Errorf("failed to check whether repository exists: %w", err)
	} else if !repoExists {
		return fmt.Errorf("repository %q could not be found", configRepoURL)
	}

	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, configRepoURL)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	// FIXME: should come from flags
	helmRepo := types.NamespacedName{
		Name:      "weaveworks-charts",
		Namespace: "flux-system",
	}

	version, err := s.discoverHelmRepository(ctx, r, GetOptions{
		Name:      opts.Name,
		Version:   opts.Version,
		Cluster:   opts.Cluster,
		Namespace: opts.Namespace,
		Port:      opts.ProfilesPort,
	})
	if err != nil {
		return fmt.Errorf("failed to discover HelmRepository: %w", err)
	}

	opts.Version = version

	files, err := gitProvider.GetRepoDirFiles(ctx, configRepoURL, git.GetSystemPath(opts.Cluster), defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' for config repository %q: %s", git.GetSystemPath(opts.Cluster), configRepoURL, err)
	}

	fileContent := getGitCommitFileContent(files, git.GetProfilesPath(opts.Cluster, ManifestFileName))

	content, err := addHelmRelease(helmRepo, fileContent, opts.Name, opts.Version, opts.Cluster, opts.Namespace)
	if err != nil {
		return fmt.Errorf("failed to add HelmRelease for profile '%s' to %s: %w", opts.Name, ManifestFileName, err)
	}

	path := git.GetProfilesPath(opts.Cluster, ManifestFileName)
	pr, err := gitProvider.CreatePullRequest(ctx, configRepoURL, prInfo(opts, "add", defaultBranch, gitprovider.CommitFile{
		Path:    &path,
		Content: &content,
	}))

	if err != nil {
		return fmt.Errorf("failed to create pull request: %s", err)
	}

	s.Logger.Actionf("created Pull Request: %s", pr.Get().WebURL)

	if opts.AutoMerge {
		s.Logger.Actionf("auto-merge=true; merging PR number %v", pr.Get().Number)

		if err := gitProvider.MergePullRequest(ctx, configRepoURL, pr.Get().Number, AddCommitMessage); err != nil {
			return fmt.Errorf("error auto-merging PR: %w", err)
		}
	}

	s.printAddSummary(opts)

	return nil
}

func prInfo(opts Options, action, defaultBranch string, commitFile gitprovider.CommitFile) gitproviders.PullRequestInfo {
	title := fmt.Sprintf("GitOps %s %s", action, opts.Name)

	if opts.Title != "" {
		title = opts.Title
	}

	titleCaser := cases.Title(language.AmericanEnglish)

	description := fmt.Sprintf("%s manifest for %s profile", titleCaser.String(action), opts.Name)
	if opts.Description != "" {
		description = opts.Description
	}

	commitMessage := fmt.Sprintf("%s profile manifests", titleCaser.String(action))
	if opts.Message != "" {
		commitMessage = opts.Message
	}

	headBranch := defaultBranch
	if opts.HeadBranch != "" {
		headBranch = opts.HeadBranch
	}

	newBranch := uuid.New().String()
	if opts.BaseBranch != "" {
		newBranch = opts.BaseBranch
	}

	return gitproviders.PullRequestInfo{
		Title:         title,
		Description:   description,
		CommitMessage: commitMessage,
		TargetBranch:  headBranch,
		NewBranch:     newBranch,
		Files:         []gitprovider.CommitFile{commitFile},
	}
}

func (s *ProfilesSvc) printAddSummary(opts Options) {
	s.Logger.Println("Adding profile:\n")
	s.Logger.Println("Name: %s", opts.Name)
	s.Logger.Println("Version: %s", opts.Version)
	s.Logger.Println("Cluster: %s", opts.Cluster)
	s.Logger.Println("Namespace: %s\n", opts.Namespace)
}

func addHelmRelease(helmRepo types.NamespacedName, fileContent, name, version, cluster, ns string) (string, error) {
	existingReleases, err := helm.SplitHelmReleaseYAML([]byte(fileContent))
	if err != nil {
		return "", fmt.Errorf("error splitting into YAML: %w", err)
	}

	newRelease := helm.MakeHelmRelease(name, version, cluster, ns, helmRepo)

	if releaseIsInNamespace(existingReleases, newRelease.Name, ns) {
		return "", fmt.Errorf("found another HelmRelease for profile '%s' in namespace %s", name, ns)
	}

	return helm.AppendHelmReleaseToString(fileContent, newRelease)
}

func releaseIsInNamespace(existingReleases []*helmv2beta1.HelmRelease, name, ns string) bool {
	for _, r := range existingReleases {
		if r.Name == name && r.Namespace == ns {
			return true
		}
	}

	return false
}
