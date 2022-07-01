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
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

const (
	GitProviderGitHub = "github"
	GitProviderGitLab = "gitlab"
	TokenTypeOauth    = "oauth2"
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
	Expect(normErr).ShouldNot(HaveOccurred())

	re := regexp.MustCompile("^[^/]+//[^/]+/([^/]+)/([^/]+).*$")
	matches := re.FindStringSubmatch(strings.TrimSuffix(normalized.String(), ".git"))

	return matches[1], matches[2]
}

func configRepoAbsolutePath(gp GitProviderEnv) string {
	return path.Join(os.Getenv("HOME"), gp.Repo)
}

func initAndCreateEmptyRepo(gp GitProviderEnv, isPrivateRepo bool) {
	repoAbsolutePath := configRepoAbsolutePath(gp)

	deleteRepo(gp)
	err := deleteDirectory([]string{repoAbsolutePath})
	Expect(err).ShouldNot(HaveOccurred())

	err = createGitRepository(gp, "main", isPrivateRepo)
	Expect(err).ShouldNot(HaveOccurred())

	err = waitUntil(POLL_INTERVAL_5SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`git clone git@%s:%s/%s.git %s`, gp.Hostname, gp.Org, gp.Repo, repoAbsolutePath))
		if err != nil {
			os.RemoveAll(repoAbsolutePath)
			return err
		}
		return nil
	})
	Expect(err).ShouldNot(HaveOccurred())
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

	description := "Weave Gitops enterprise test repository"
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
	Expect(providerErr).ShouldNot(HaveOccurred())

	ctx := context.Background()
	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)

	// allow repo to be absent (as tests assume this)
	if repoErr == nil {
		deleteErr := or.Delete(ctx)
		Expect(deleteErr).ShouldNot(HaveOccurred())
	}
	repoErr = waitUntil(POLL_INTERVAL_1SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Get(ctx, orgRef)
		return err
	}, true)
	Expect(repoErr).ShouldNot(HaveOccurred(), fmt.Sprintf("repo %s is accessible through the api ...\n", gp.Repo))
}

func verifyPRCreated(gp GitProviderEnv, repoAbsolutePath string) string {
	ctx := context.Background()

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteUrl(repoAbsolutePath, "origin")
	Expect(repoUrlErr).ShouldNot(HaveOccurred())

	org, _ := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(gp.Type, org, filepath.Base(repoAbsolutePath), gp.Token, gp.TokenType, gp.Hostname)
	Expect(providerErr).ShouldNot(HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	Expect(repoErr).ShouldNot(HaveOccurred())

	var prs []gitprovider.PullRequest

	Eventually(func(g Gomega) {
		prs, err := or.PullRequests().List(ctx)
		g.Expect(err).ShouldNot(HaveOccurred())

		g.Expect(len(prs)).To(BeNumerically(">=", 1))
		g.Expect(prs[0].Get().Merged).To(BeFalse())
	}, ASSERTION_1MINUTE_TIME_OUT).Should(Succeed())

	return prs[0].Get().WebURL
}

func mergePullRequest(gp GitProviderEnv, repoAbsolutePath string, prLink string) {
	ctx := context.Background()
	prNumberStr := filepath.Base(prLink)
	prNumber, numErr := strconv.Atoi(prNumberStr)
	Expect(numErr).ShouldNot(HaveOccurred())

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteUrl(repoAbsolutePath, "origin")
	Expect(repoUrlErr).ShouldNot(HaveOccurred())

	org, repo := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(gp.Type, org, repo, gp.Token, gp.TokenType, gp.Hostname)
	Expect(providerErr).ShouldNot(HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	Expect(repoErr).ShouldNot(HaveOccurred())

	Eventually(func(g Gomega) {
		err := or.PullRequests().Merge(ctx, prNumber, gitprovider.MergeMethodMerge, "merge for test")
		g.Expect(err).ShouldNot(HaveOccurred())

	}, ASSERTION_1MINUTE_TIME_OUT).Should(Succeed())

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

func createGitRepoBranch(repoAbsolutePath string, branchName string) string {
	stdOut, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git checkout -b %s && git push --set-upstream origin %s && git checkout main`, repoAbsolutePath, branchName, branchName), ASSERTION_30SECONDS_TIME_OUT)
	return stdOut
}

func pullGitRepo(repoAbsolutePath string) {
	_, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git pull --rebase`, repoAbsolutePath), ASSERTION_30SECONDS_TIME_OUT)
}

func cleanGitRepository(subDirName string) {
	repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
	logger.Infof("Deleting directory %s from repo: %s", subDirName, repoAbsolutePath)

	pullGitRepo(repoAbsolutePath)
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("rm -rf %s", path.Join(repoAbsolutePath, subDirName)))
	gitUpdateCommitPush(repoAbsolutePath, "")
}
