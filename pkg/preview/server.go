package preview

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/preview"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/server/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/ptr"
)

type ServerOpts struct {
	logr.Logger
	git.ProviderCreator
}

type server struct {
	pb.UnimplementedPreviewServiceServer

	log             logr.Logger
	providerCreator git.ProviderCreator
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewPreviewServiceServer(opts)

	return pb.RegisterPreviewServiceHandlerServer(ctx, mux, s)
}

func NewPreviewServiceServer(opts ServerOpts) pb.PreviewServiceServer {
	return &server{
		log:             opts.Logger,
		providerCreator: opts.ProviderCreator,
	}
}

func (s *server) GetYAML(ctx context.Context, msg *pb.GetYAMLRequest) (*pb.GetYAMLResponse, error) {
	yamlObj, err := generateYAML(msg.GetResource())
	if err != nil {
		return nil, fmt.Errorf("failed to generate YAML for %q: %w", msg.GetResource().GetType(), err)
	}

	path := msg.GetPath()
	if path == "" {
		path = getRepositoryFilePath(yamlObj.name, yamlObj.namespace)
	}

	return &pb.GetYAMLResponse{
		File: &pb.PathContent{
			Path:    path,
			Content: yamlObj.yaml,
		},
	}, nil
}

func (s *server) CreatePullRequest(ctx context.Context, msg *pb.CreatePullRequestRequest) (*pb.CreatePullRequestResponse, error) {
	if msg.GetRepositoryUrl() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("repository URL is required"))
	}

	if msg.GetHeadBranch() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("head branch is required"))
	}

	if msg.GetBaseBranch() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("base branch is required"))
	}

	if msg.GetTitle() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("title is required"))
	}

	if msg.GetDescription() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("description is required"))
	}

	if msg.GetCommitMessage() == "" {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("commit message is required"))
	}

	if msg.GetResource() == nil {
		return nil, fmt.Errorf("failed to create pull request: %w", errors.New("resource is required"))
	}

	yamlObj, err := generateYAML(msg.GetResource())
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	path := msg.GetPath()
	if path == "" {
		path = getRepositoryFilePath(yamlObj.name, yamlObj.namespace)
	}

	var commits []git.Commit
	commits = append(commits, git.Commit{
		CommitMessage: msg.GetCommitMessage(),
		Files: []git.CommitFile{
			{
				Path:    path,
				Content: ptr.To(yamlObj.yaml),
			},
		},
	})

	providerType, providerHostname, err := getProviderTypeAndHostname(msg.GetRepositoryUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	providerToken, providerTokenType, err := getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	providerOptions := []git.ProviderWithFn{git.WithDomain(providerHostname)}
	if providerType == git.AzureDevOpsProviderName {
		providerOptions = append(providerOptions, git.WithToken(providerTokenType, providerToken))
	} else if providerType == git.BitBucketServerProviderName {
		providerOptions = append(providerOptions, git.WithUsername(""))
		providerOptions = append(providerOptions, git.WithToken(providerTokenType, providerToken))
	} else if providerType == git.GitHubProviderName {
		providerOptions = append(providerOptions, git.WithOAuth2Token(providerToken))
	} else if providerType == git.GitLabProviderName {
		providerOptions = append(providerOptions, git.WithToken(providerTokenType, providerToken))
	}

	provider, err := s.providerCreator.Create(providerType, providerOptions...)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "error creating pull request: %s", err.Error())
	}

	res, err := provider.CreatePullRequest(ctx, git.PullRequestInput{
		RepositoryURL: msg.GetRepositoryUrl(),
		Title:         msg.GetTitle(),
		Body:          msg.GetDescription(),
		Head:          msg.GetHeadBranch(),
		Base:          msg.GetBaseBranch(),
		Commits:       commits,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request: %w", err)
	}

	return &pb.CreatePullRequestResponse{
		WebUrl: res.Link,
	}, nil
}

func getRepositoryFilePath(name, namespace string) string {
	cluster := os.Getenv("CLUSTER_NAME")
	return fmt.Sprintf("clusters/%s/namespaces/%s/%s.yaml", cluster, namespace, name)
}

func getToken(ctx context.Context) (string, string, error) {
	token := viper.GetString("git-provider-token")

	providerToken, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		// fallback to env token
		return token, "", nil
	}

	return providerToken.AccessToken, "oauth2", nil
}

func getProviderTypeAndHostname(repositoryURL string) (string, string, error) {
	// read defaults from config
	providerType := viper.GetString("git-provider-type")
	providerHostname := viper.GetString("git-provider-hostname")

	// if user supplies a different gitrepo, derive the provider type and host from the URL
	if repositoryURL != "" {
		repoURL, err := gitproviders.NewRepoURL(repositoryURL)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse repository URL: %w", err)
		}

		// override defaults
		providerType = string(repoURL.Provider())
		providerHostname = repoURL.URL().Host
	}

	return providerType, providerHostname, nil
}
