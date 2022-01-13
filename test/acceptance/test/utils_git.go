package acceptance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

const (
	GitProviderGitHub = "github"
	GitProviderGitLab = "gitlab"
	tokenTypeOauth    = "oauth2"
)

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

func getRepoVisibility(org string, repo string, providerName string) string {
	gitProvider, orgRef, err := getGitProvider(org, repo, providerName)
	Expect(err).ShouldNot(HaveOccurred())

	orgInfo, err := gitProvider.OrgRepositories().Get(context.Background(), orgRef)
	Expect(err).ShouldNot(HaveOccurred())

	visibility := string(*orgInfo.Get().Visibility)

	return visibility
}

func initAndCreateEmptyRepo(repoName string, providerName string, isPrivateRepo bool, org string) string {
	repoAbsolutePath := path.Join("/tmp/", repoName)

	deleteRepo(CLUSTER_REPOSITORY, GIT_PROVIDER, GITHUB_ORG)
	err := deleteDirectory([]string{repoAbsolutePath})
	Expect(err).ShouldNot(HaveOccurred())

	err = createGitRepository(repoName, "main", isPrivateRepo, providerName, org)
	Expect(err).ShouldNot(HaveOccurred())

	err = waitUntil(os.Stdout, POLL_INTERVAL_5SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		command := exec.Command("sh", "-c", fmt.Sprintf(`
            git clone git@%s.com:%s/%s.git %s`, providerName, org, repoName, repoAbsolutePath))
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err := command.Run()
		if err != nil {
			os.RemoveAll(repoAbsolutePath)
			return err
		}
		return nil
	})
	Expect(err).ShouldNot(HaveOccurred())

	return repoAbsolutePath
}

func createGitRepository(repoName, branch string, private bool, providerName string, org string) error {
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

	gitProvider, orgRef, err := getGitProvider(org, repoName, providerName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Printf("creating repo %s ...\n", repoName)

	if err := waitUntil(os.Stdout, time.Second, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Create(ctx, orgRef, repoInfo, repoCreateOpts)
		if err != nil && strings.Contains(err.Error(), "rate limit exceeded") {
			waitForRateQuota, err := getWaitTimeFromErr(err.Error())
			if err != nil {
				return err
			}
			fmt.Printf("Waiting for rate quota %s \n", waitForRateQuota.String())
			time.Sleep(waitForRateQuota)
			return fmt.Errorf("retry after waiting for rate quota")
		}
		return err
	}); err != nil {
		return fmt.Errorf("error creating repo %s", err)
	}

	fmt.Printf("repo %s created ...\n", repoName)
	fmt.Printf("validating access to the repo %s ...\n", repoName)

	err = waitUntil(os.Stdout, POLL_INTERVAL_1SECONDS, ASSERTION_30SECONDS_TIME_OUT, func() error {
		_, err := gitProvider.OrgRepositories().Get(ctx, orgRef)
		return err
	})
	if err != nil {
		return fmt.Errorf("error validating access to the repository %w", err)
	}
	fmt.Printf("repo %s is accessible through the api ...\n", repoName)

	return nil
}

func getGitProvider(org string, repo string, providerName string) (gitprovider.Client, gitprovider.OrgRepositoryRef, error) {
	var gitProvider gitprovider.Client

	var orgRef gitprovider.OrgRepositoryRef

	var err error

	switch providerName {
	case GitProviderGitHub:
		orgRef = gitproviders.NewOrgRepositoryRef(github.DefaultDomain, org, repo)

		gitProvider, err = github.NewClient(
			gitprovider.WithOAuth2Token(GITHUB_TOKEN),
			gitprovider.WithDestructiveAPICalls(true),
		)
	case GitProviderGitLab:
		orgRef = gitproviders.NewOrgRepositoryRef(gitlab.DefaultDomain, org, repo)

		gitProvider, err = gitlab.NewClient(
			GITLAB_TOKEN,
			tokenTypeOauth,
			gitprovider.WithOAuth2Token(GITLAB_TOKEN),
			gitprovider.WithDestructiveAPICalls(true),
		)
	default:
		err = fmt.Errorf("invalid git provider name: %s", providerName)
	}

	return gitProvider, orgRef, err
}

func deleteRepo(repoName string, providerName string, org string) {
	log.Printf("Delete application repo: %s", path.Join(GITHUB_ORG, repoName))

	gitProvider, orgRef, providerErr := getGitProvider(org, repoName, providerName)
	Expect(providerErr).ShouldNot(HaveOccurred())

	ctx := context.Background()
	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)

	// allow repo to be absent (as tests assume this)
	if repoErr == nil {
		deleteErr := or.Delete(ctx)
		Expect(deleteErr).ShouldNot(HaveOccurred())
	}
}

func verifyPRCreated(repoAbsolutePath, providerName string) string {
	ctx := context.Background()

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteUrl(repoAbsolutePath, "origin")
	Expect(repoUrlErr).ShouldNot(HaveOccurred())

	org, _ := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(org, filepath.Base(repoAbsolutePath), providerName)
	Expect(providerErr).ShouldNot(HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	Expect(repoErr).ShouldNot(HaveOccurred())

	prs, err := or.PullRequests().List(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(len(prs)).To(Equal(1))
	return prs[0].Get().WebURL
}

func mergePullRequest(repoAbsolutePath string, prLink string, providerName string) {
	ctx := context.Background()
	prNumberStr := filepath.Base(prLink)
	prNumber, numErr := strconv.Atoi(prNumberStr)
	Expect(numErr).ShouldNot(HaveOccurred())

	repoUrlString, repoUrlErr := git.New(nil, wrapper.NewGoGit()).GetRemoteUrl(repoAbsolutePath, "origin")
	Expect(repoUrlErr).ShouldNot(HaveOccurred())

	org, repo := extractOrgAndRepo(repoUrlString)
	gitProvider, orgRef, providerErr := getGitProvider(org, repo, providerName)
	Expect(providerErr).ShouldNot(HaveOccurred())

	or, repoErr := gitProvider.OrgRepositories().Get(ctx, orgRef)
	Expect(repoErr).ShouldNot(HaveOccurred())

	err := or.PullRequests().Merge(ctx, prNumber, gitprovider.MergeMethodMerge, "merge for test")
	Expect(err).ShouldNot(HaveOccurred())
}

func gitUpdateCommitPush(repoAbsolutePath string, commitMessage string) {
	log.Infof("Pushing changes made to file(s) in repo: %s", repoAbsolutePath)
	if commitMessage == "" {
		commitMessage = "edit repo file"
	}

	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cd %s && git add -u && git add -A && git commit -m '%s' && git pull --rebase && git push origin HEAD", repoAbsolutePath, commitMessage))
}

func getGitRepositoryURL(repoAbsolutePath string) string {
	repoURL, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git config --get remote.origin.url`, repoAbsolutePath))
	return repoURL
}

func createGitRepoBranch(repoAbsolutePath string, branchName string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && git checkout -b %s && git push --set-upstream origin %s", repoAbsolutePath, branchName, branchName))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return string(session.Wait().Out.Contents())
}

func pullGitRepo(repoAbsolutePath string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && git pull", repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}
