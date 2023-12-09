package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

const (
	GitProviderGitHub = "github"
	GitProviderGitLab = "gitlab"
	TokenTypeOauth    = "oauth2"
	TmpRepoPath       = "/tmp/wge-git-repos"
)

type GitProviderEnv struct {
	Type           string
	Token          string
	Username       string
	Password       string
	TokenType      string
	Hostname       string
	Org            string
	Repo           string
	HostTypes      string
	ClientId       string
	ClientSecret   string
	GitlabHostname string
}

func initGitProviderData() GitProviderEnv {

	if GetEnv("GIT_PROVIDER", GitProviderGitHub) == GitProviderGitHub {
		return GitProviderEnv{
			Type:      GitProviderGitHub,
			Hostname:  GetEnv("GIT_PROVIDER_HOSTNAME", github.DefaultDomain),
			TokenType: TokenTypeOauth,
			Token:     GetEnv("GITHUB_TOKEN", ""),
			Org:       GetEnv("GITHUB_ORG", ""),
			Repo:      GetEnv("CLUSTER_REPOSITORY", ""),
			Username:  GetEnv("GITHUB_USER", ""),
			Password:  GetEnv("GITHUB_PASSWORD", ""),
		}
	} else {
		// `gitops` binary reads WEAVE_GITOPS_GIT_HOST_TYPES w/ a GITOPS_ prefix
		// while EE just reads GIT_HOST_TYPES, reconcile them here.
		hostTypes := GetEnv("WEAVE_GITOPS_GIT_HOST_TYPES", "")
		if hostTypes != "" {
			viper.Set("git-host-types", hostTypes)
		}
		return GitProviderEnv{
			Type:           GitProviderGitLab,
			Hostname:       GetEnv("GIT_PROVIDER_HOSTNAME", gitlab.DefaultDomain),
			TokenType:      TokenTypeOauth,
			Token:          GetEnv("GITLAB_TOKEN", ""),
			Org:            GetEnv("GITLAB_ORG", ""),
			Repo:           GetEnv("CLUSTER_REPOSITORY", ""),
			Username:       GetEnv("GITLAB_USER", ""),
			Password:       GetEnv("GITLAB_PASSWORD", ""),
			ClientId:       GetEnv("GITLAB_CLIENT_ID", ""),
			ClientSecret:   GetEnv("GITLAB_CLIENT_SECRET", ""),
			HostTypes:      GetEnv("WEAVE_GITOPS_GIT_HOST_TYPES", ""),
			GitlabHostname: GetEnv("GITLAB_HOSTNAME", ""),
		}
	}
}

// WaitUntil runs checkDone until a timeout is reached
func waitUntil(poll, timeout time.Duration, checkDone func() error, expectError ...bool) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		err := checkDone()

		if len(expectError) > 0 && expectError[0] {
			if err != nil {
				return nil
			}
		} else {
			if err == nil {
				return nil
			}
		}
		logger.Tracef("error occurred %s, retrying in %s\n", err, poll.String())
	}
	return fmt.Errorf("timeout reached %s", timeout.String())
}

func getWaitTimeFromErr(errOutput string) (time.Duration, error) {
	var re = regexp.MustCompile(`(?m)\[rate reset in (.*)\]`)
	match := re.FindAllStringSubmatch(errOutput, -1)

	if len(match) >= 1 && len(match[1][0]) > 0 {
		duration, err := time.ParseDuration(match[1][0])
		if err != nil {
			return 0, fmt.Errorf("error pasing rate reset time %w", err)
		}

		return duration, nil
	}

	return 0, fmt.Errorf("could not found a rate reset on string: %s", errOutput)
}

func extractOrgAndRepo(url string) (string, string) {
	normalized, normErr := gitproviders.NewRepoURL(url)
	gomega.Expect(normErr).ShouldNot(gomega.HaveOccurred())

	re := regexp.MustCompile("^[^/]+//[^/]+/([^/]+)/([^/]+).*$")
	matches := re.FindStringSubmatch(strings.TrimSuffix(normalized.String(), ".git"))

	return matches[1], matches[2]
}

func configRepoAbsolutePath(gp GitProviderEnv) string {
	return path.Join(TmpRepoPath, gp.Repo)
}

func initAndCreateEmptyRepo(gp GitProviderEnv, isPrivateRepo bool) {
	repoAbsolutePath := configRepoAbsolutePath(gp)

	deleteRepo(gp)
	err := deleteDirectory([]string{repoAbsolutePath})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = createGitRepository(gp, "main", isPrivateRepo)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = waitUntil(POLL_INTERVAL_5SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		err = os.MkdirAll(TmpRepoPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %w", TmpRepoPath, err)
		}
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`git clone git@%s:%s/%s.git %s`, gp.Hostname, gp.Org, gp.Repo, repoAbsolutePath))
		if err != nil {
			os.RemoveAll(repoAbsolutePath)
			return err
		}
		return nil
	})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func addSchemeToDomain(domain string) string {
	if domain != github.DefaultDomain && domain != gitlab.DefaultDomain && !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		return "https://" + domain
	}
	return domain
}

func createGitRepository(gp GitProviderEnv, branch string, private bool) error {
	visibility := gitprovider.RepositoryVisibilityPublic
	if private {
		visibility = gitprovider.RepositoryVisibilityPrivate
	}

	description := "Weave GitOps enterprise test repository"
	defaultBranch := branch
	repoInfo := gitprovider.RepositoryInfo{
		Description:   &description,
		Visibility:    &visibility,
		DefaultBranch: &defaultBranch,
	}

	repoCreateOpts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	gitProvider, orgRef, err := getGitProvider(gp.Type, gp.Org, gp.Repo, gp.Token, gp.TokenType, gp.Hostname)
	if err != nil {
		return err
	}

	ctx := context.Background()

	logger.Infof("creating repo %s ...", gp.Repo)

	if err := waitUntil(time.Second, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Create(ctx, orgRef, repoInfo, repoCreateOpts)
		if err != nil && strings.Contains(err.Error(), "rate limit exceeded") {
			waitForRateQuota, err := getWaitTimeFromErr(err.Error())
			if err != nil {
				return err
			}
			logger.Infof("Waiting for rate quota %s \n", waitForRateQuota.String())
			time.Sleep(waitForRateQuota)
			return fmt.Errorf("retry after waiting for rate quota")
		}
		return err
	}); err != nil {
		return fmt.Errorf("error creating repo %s", err)
	}

	logger.Infof("repo %s created ...", gp.Repo)
	logger.Infof("validating access to the repo %s ...", gp.Repo)

	err = waitUntil(POLL_INTERVAL_1SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Get(ctx, orgRef)
		return err
	})
	if err != nil {
		return fmt.Errorf("error validating access to the repository %w", err)
	}
	logger.Infof("repo %s is accessible through the api ...", gp.Repo)

	return nil
}

func getGitProvider(provider string, org string, repo string, token string, tokenType string, hostName string) (gitprovider.Client, gitprovider.OrgRepositoryRef, error) {
	var gitProvider gitprovider.Client

	var orgRef gitprovider.OrgRepositoryRef

	var err error

	switch provider {
	case GitProviderGitHub:
		orgRef = gitproviders.NewOrgRepositoryRef(github.DefaultDomain, org, repo)

		gitProvider, err = github.NewClient(
			gitprovider.WithOAuth2Token(token),
			gitprovider.WithDestructiveAPICalls(true),
		)
	case GitProviderGitLab:

		if hostName == gitlab.DefaultDomain {
			orgRef = gitproviders.NewOrgRepositoryRef(gitlab.DefaultDomain, org, repo)
			gitProvider, err = gitlab.NewClient(
				token,
				tokenType,
				gitprovider.WithOAuth2Token(token),
				gitprovider.WithDestructiveAPICalls(true),
			)
		} else {
			hostName = addSchemeToDomain(hostName)
			orgRef = gitproviders.NewOrgRepositoryRef(hostName, org, repo)
			gitProvider, err = gitlab.NewClient(
				token,
				tokenType,
				gitprovider.WithDomain(hostName),
				gitprovider.WithOAuth2Token(token),
				gitprovider.WithDestructiveAPICalls(true),
			)
		}

	default:
		err = fmt.Errorf("invalid git provider name: %s", provider)
	}

	return gitProvider, orgRef, err
}

func deleteRepo(gp GitProviderEnv) {
	logger.Infof("Delete application repo: %s", path.Join(gp.Org, gp.Repo))

	gitProvider, orgRef, providerErr := getGitProvider(gp.Type, gp.Org, gp.Repo, gp.Token, gp.TokenType, gp.Hostname)
	gomega.Expect(providerErr).ShouldNot(gomega.HaveOccurred())

	ctx := context.Background()
	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)

	// allow repo to be absent (as tests assume this)
	if repoErr == nil {
		deleteErr := or.Delete(ctx)
		gomega.Expect(deleteErr).ShouldNot(gomega.HaveOccurred())
	}
	repoErr = waitUntil(POLL_INTERVAL_1SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Get(ctx, orgRef)
		return err
	}, true)
	gomega.Expect(repoErr).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("repo %s is accessible through the api ...\n", gp.Repo))
}

func verifyPRCreated(gp GitProviderEnv, repoAbsolutePath string) string {
	ctx := context.Background()

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteURL(repoAbsolutePath, "origin")
	gomega.Expect(repoUrlErr).ShouldNot(gomega.HaveOccurred())

	org, _ := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(gp.Type, org, filepath.Base(repoAbsolutePath), gp.Token, gp.TokenType, gp.Hostname)
	gomega.Expect(providerErr).ShouldNot(gomega.HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	gomega.Expect(repoErr).ShouldNot(gomega.HaveOccurred())

	var prs []gitprovider.PullRequest

	gomega.Eventually(func(g gomega.Gomega) {
		var err error
		prs, err = or.PullRequests().List(ctx)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		g.Expect(len(prs)).To(gomega.BeNumerically(">=", 1))
		g.Expect(prs[0].Get().Merged).To(gomega.BeFalse())
	}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), "Failed to verify created PR in the repository")

	return prs[0].Get().WebURL
}

func mergePullRequest(gp GitProviderEnv, repoAbsolutePath string, prLink string) {
	ctx := context.Background()
	prNumberStr := filepath.Base(prLink)
	prNumber, numErr := strconv.Atoi(prNumberStr)
	gomega.Expect(numErr).ShouldNot(gomega.HaveOccurred())

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteURL(repoAbsolutePath, "origin")
	gomega.Expect(repoUrlErr).ShouldNot(gomega.HaveOccurred())

	org, repo := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(gp.Type, org, repo, gp.Token, gp.TokenType, gp.Hostname)
	gomega.Expect(providerErr).ShouldNot(gomega.HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	gomega.Expect(repoErr).ShouldNot(gomega.HaveOccurred())

	gomega.Eventually(func(g gomega.Gomega) {
		err := or.PullRequests().Merge(ctx, prNumber, gitprovider.MergeMethodMerge, "merge for test")
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

	}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed())

}

func gitUpdateCommitPush(repoAbsolutePath string, commitMessage string) {
	logger.Infof("Pushing changes made to file(s) in repo: %s", repoAbsolutePath)
	if commitMessage == "" {
		commitMessage = "edit repo file"
	}

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("cd %s && git add -u && git add -A && git commit -m '%s' && git push", repoAbsolutePath, commitMessage))
}

func getGitRepositoryURL(repoAbsolutePath string) string {
	repoURL, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git config --get remote.origin.url`, repoAbsolutePath), ASSERTION_30SECONDS_TIME_OUT)
	return repoURL
}

func pullGitRepo(repoAbsolutePath string) {
	_, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git pull --rebase`, repoAbsolutePath), ASSERTION_30SECONDS_TIME_OUT)
}

func cleanGitRepository(subDirName string) {
	repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
	logger.Infof("Deleting directory %s from repo: %s", subDirName, repoAbsolutePath)

	absDirPath := path.Join(repoAbsolutePath, subDirName)
	if absDirPath != repoAbsolutePath {
		pullGitRepo(repoAbsolutePath)
		_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("rm -rf %s", path.Join(repoAbsolutePath, subDirName)))
		gitUpdateCommitPush(repoAbsolutePath, "")
	} else {
		logger.Warnf("Deleting management cluster config repository is not allowed: %s", absDirPath)
	}
}
