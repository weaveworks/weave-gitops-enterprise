package gitproviders

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestDetectGitProviderFromURL(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name     string
		url      string
		provider GitProviderName
	}{
		{name: "ssh+github", url: "ssh://git@github.com/weaveworks/weave-gitops.git", provider: GitProviderGitHub},
		{name: "ssh+gitlab", url: "ssh://git@gitlab.com/weaveworks/weave-gitops.git", provider: GitProviderGitLab},
		{name: "https+bitbucket", url: "https://bitbucket.weave.works/scm/wg/config.git", provider: GitProviderBitBucketServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := detectGitProviderFromURL(tt.url, map[string]string{
				"bitbucket.weave.works": "bitbucket-server",
			})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(provider).To(Equal(tt.provider))
		})
	}
}

func TestGetOwnerFromURL(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name     string
		url      string
		provider GitProviderName
		owner    string
	}{
		{name: "github", url: "ssh://git@github.com/weaveworks/weave-gitops.git", provider: GitProviderGitHub, owner: "weaveworks"},
		{name: "gitlab", url: "ssh://git@gitlab.com/weaveworks/weave-gitops.git", provider: GitProviderGitLab, owner: "weaveworks"},
		{name: "gitlab with subgroup", url: "ssh://git@gitlab.com/weaveworks/infra/weave-gitops.git", provider: GitProviderGitLab, owner: "weaveworks/infra"},
		{name: "gitlab with nested subgroup", url: "ssh://git@gitlab.com/weaveworks/infra/dev/weave-gitops.git", provider: GitProviderGitLab, owner: "weaveworks/infra/dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			g.Expect(err).NotTo(HaveOccurred())
			owner, err := getOwnerFromURL(*u, tt.provider)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(owner).To(Equal(tt.owner))
		})
	}

	t.Run("missing owner", func(t *testing.T) {
		u, err := url.Parse("ssh://git@gitlab.com/weave-gitops.git")
		g.Expect(err).NotTo(HaveOccurred())
		_, err = getOwnerFromURL(*u, GitProviderGitLab)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("could not get owner from url ssh://git@gitlab.com/weave-gitops.git"))
	})

	t.Run("empty url", func(t *testing.T) {
		u, err := url.Parse("")
		g.Expect(err).NotTo(HaveOccurred())
		_, err = getOwnerFromURL(*u, GitProviderGitLab)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("could not get owner from url "))
	})
}

type expectedRepoURL struct {
	s        string
	owner    string
	name     string
	provider GitProviderName
	protocol RepositoryURLProtocol
}

func TestNewRepoURL(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name           string
		url            string
		gitProviderEnv string
		result         expectedRepoURL
	}{

		{
			name: "github git clone style",
			url:  "git@github.com:someuser/podinfo.git",
			result: expectedRepoURL{
				s:        "ssh://git@github.com/someuser/podinfo.git",
				owner:    "someuser",
				name:     "podinfo",
				provider: GitProviderGitHub,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "github url style",
			url:  "ssh://git@github.com/someuser/podinfo.git",
			result: expectedRepoURL{
				s:        "ssh://git@github.com/someuser/podinfo.git",
				owner:    "someuser",
				name:     "podinfo",
				provider: GitProviderGitHub,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "github https",
			url:  "https://github.com/someuser/podinfo.git",
			result: expectedRepoURL{
				s:        "ssh://git@github.com/someuser/podinfo.git",
				owner:    "someuser",
				name:     "podinfo",
				provider: GitProviderGitHub,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "gitlab git clone style",
			url:  "git@gitlab.com:someuser/podinfo.git",
			result: expectedRepoURL{
				s:        "ssh://git@gitlab.com/someuser/podinfo.git",
				owner:    "someuser",
				name:     "podinfo",
				provider: GitProviderGitLab,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "gitlab https",
			url:  "https://gitlab.com/someuser/podinfo.git",
			result: expectedRepoURL{
				s:        "ssh://git@gitlab.com/someuser/podinfo.git",
				owner:    "someuser",
				name:     "podinfo",
				provider: GitProviderGitLab,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "trailing slash in url",
			url:  "https://github.com/sympatheticmoose/podinfo-deploy/",
			result: expectedRepoURL{
				s:        "ssh://git@github.com/sympatheticmoose/podinfo-deploy.git",
				owner:    "sympatheticmoose",
				name:     "podinfo-deploy",
				provider: GitProviderGitHub,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "subsubgroup",
			url:  "https://github.com/sympatheticmoose/infra/dev/podinfo-deploy/",
			result: expectedRepoURL{
				s:        "ssh://git@github.com/sympatheticmoose/infra/dev/podinfo-deploy.git",
				owner:    "sympatheticmoose/infra/dev",
				name:     "podinfo-deploy",
				provider: GitProviderGitHub,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name:           "custom domain",
			url:            "git@gitlab.acme.org/sympatheticmoose/podinfo-deploy/",
			gitProviderEnv: "gitlab.acme.org=gitlab",
			result: expectedRepoURL{
				s:        "ssh://git@gitlab.acme.org/sympatheticmoose/podinfo-deploy.git",
				owner:    "sympatheticmoose",
				name:     "podinfo-deploy",
				provider: "gitlab",
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "azure ssh clone",
			url:  "git@ssh.dev.azure.com:v3/weaveworks/weave-gitops-integration/config",
			result: expectedRepoURL{
				s:        "ssh://git@ssh.dev.azure.com/v3/weaveworks/weave-gitops-integration/config.git",
				owner:    "weaveworks/weave-gitops-integration",
				name:     "config",
				provider: GitProviderAzureDevOps,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "azure https clone",
			url:  "https://weaveworks@dev.azure.com/weaveworks/weave-gitops-integration/_git/config",
			result: expectedRepoURL{
				s:        "ssh://git@dev.azure.com/weaveworks/weave-gitops-integration/_git/config.git",
				owner:    "weaveworks/weave-gitops-integration",
				name:     "config",
				provider: GitProviderAzureDevOps,
				protocol: RepositoryURLProtocolSSH,
			},
		},
		{
			name: "azure https",
			url:  "https://dev.azure.com/weaveworks/weave-gitops-integration/_git/config",
			result: expectedRepoURL{
				s:        "ssh://git@dev.azure.com/weaveworks/weave-gitops-integration/_git/config.git",
				owner:    "weaveworks/weave-gitops-integration",
				name:     "config",
				provider: GitProviderAzureDevOps,
				protocol: RepositoryURLProtocolSSH,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitProviderEnv != "" {
				viper.Set("git-host-types", tt.gitProviderEnv)
			}
			result, err := NewRepoURL(tt.url)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(result.String()).To(Equal(tt.result.s))
			u, err := url.Parse(tt.result.s)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.URL()).To(Equal(u))
			g.Expect(result.Owner()).To(Equal(tt.result.owner))
			g.Expect(result.Provider()).To(Equal(tt.result.provider))
			g.Expect(result.Protocol()).To(Equal(tt.result.protocol))
		})
	}
}
