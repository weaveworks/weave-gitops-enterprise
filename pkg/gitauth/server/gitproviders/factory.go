package gitproviders

// GitProviderName holds a Git provider definition.
type GitProviderName string

const (
	GitProviderGitHub          GitProviderName = "github"
	GitProviderGitLab          GitProviderName = "gitlab"
	GitProviderBitBucketServer GitProviderName = "bitbucket-server"
	GitProviderAzureDevOps     GitProviderName = "azure-devops"
)
