package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	common_utils "github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1_proto.CreatePullRequestRequest) (*capiv1_proto.CreatePullRequestResponse, error) {
	gp, err := getGitProvider(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	if err := validateCreateClusterPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}

	tmplWithValues, err := renderTemplateWithValues(tmpl, msg.TemplateName, msg.ParameterValues)
	if err != nil {
		return nil, fmt.Errorf("failed to render template with parameter values: %w", err)
	}

	err = capi.ValidateRenderedTemplates(tmplWithValues)
	if err != nil {
		return nil, fmt.Errorf("validation error rendering template %v, %v", msg.TemplateName, err)
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.log, client, tmplWithValues, msg.Credentials, msg.TemplateName)
	if err != nil {
		return nil, err
	}

	// FIXME: parse and read from Cluster in yaml template
	clusterName, ok := msg.ParameterValues["CLUSTER_NAME"]
	if !ok {
		return nil, fmt.Errorf("unable to find 'CLUSTER_NAME' parameter in supplied values")
	}
	// FIXME: parse and read from Cluster in yaml template
	clusterNamespace, ok := msg.ParameterValues["NAMESPACE"]
	if !ok {
		s.log.Info("Couldn't find NAMESPACE param in request, using 'default'.")
		// TODO: https://weaveworks.atlassian.net/browse/WKP-2205
		clusterNamespace = "default"
	}

	path := getClusterPathInRepo(clusterName)
	content := string(tmplWithValuesAndCredentials[:])
	files := []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}

	repositoryURL := os.Getenv("CAPI_TEMPLATES_REPOSITORY_URL")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := os.Getenv("CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH")
	if msg.BaseBranch != "" {
		baseBranch = msg.BaseBranch
	}
	if msg.HeadBranch == "" {
		msg.HeadBranch = getHash(msg.RepositoryUrl, msg.ParameterValues["CLUSTER_NAME"], msg.BaseBranch)
	}
	if msg.Title == "" {
		msg.Title = fmt.Sprintf("Gitops add cluster %s", msg.ParameterValues["CLUSTER_NAME"])
	}
	if msg.Description == "" {
		msg.Description = fmt.Sprintf("Pull request to create cluster %s", msg.ParameterValues["CLUSTER_NAME"])
	}
	if msg.CommitMessage == "" {
		msg.CommitMessage = "Add Cluster Manifests"
	}
	_, err = s.provider.GetRepository(ctx, *gp, repositoryURL)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "failed to access repo %s: %s", repositoryURL, err)
	}

	if len(msg.Values) > 0 {
		profilesFile, err := generateProfileFiles(
			ctx,
			s.profileHelmRepositoryName,
			os.Getenv("RUNTIME_NAMESPACE"),
			s.helmRepositoryCacheDir,
			clusterName,
			client,
			msg.Values,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, *profilesFile)
	}

	var pullRequestURL string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		t, err := common_utils.Generate()
		if err != nil {
			return fmt.Errorf("error generating token for new cluster: %v", err)
		}

		c := &models.Cluster{
			Name:          clusterName,
			CAPIName:      clusterName,
			CAPINamespace: clusterNamespace,
			Token:         t,
		}
		if err := tx.Create(c).Error; err != nil {
			return err
		}

		// FIXME: maybe this should reconcile rather than just try to create in case of other errors, e.g. database row creation
		res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
			GitProvider:       *gp,
			RepositoryURL:     repositoryURL,
			ReposistoryAPIURL: msg.RepositoryApiUrl,
			HeadBranch:        msg.HeadBranch,
			BaseBranch:        baseBranch,
			Title:             msg.Title,
			Description:       msg.Description,
			CommitMessage:     msg.CommitMessage,
			Files:             files,
		})
		if err != nil {
			s.log.Error(err, "Failed to create pull request")
			return err
		}

		// Create the PR, this shouldn't fail, but if it does it will rollback the Cluster but not the delete the PR
		pullRequestURL = res.WebURL
		pr := &models.PullRequest{
			URL:  pullRequestURL,
			Type: "create",
		}
		if err := tx.Create(pr).Error; err != nil {
			return err
		}

		c.PullRequests = append(c.PullRequests, pr)
		if err := tx.Save(c).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create pull request and cluster rows for %q: %w", msg.TemplateName, err)
	}

	return &capiv1_proto.CreatePullRequestResponse{
		WebUrl: pullRequestURL,
	}, nil
}

func (s *server) DeleteClustersPullRequest(ctx context.Context, msg *capiv1_proto.DeleteClustersPullRequestRequest) (*capiv1_proto.DeleteClustersPullRequestResponse, error) {
	gp, err := getGitProvider(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	if err := validateDeleteClustersPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, err
	}

	repositoryURL := os.Getenv("CAPI_TEMPLATES_REPOSITORY_URL")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := os.Getenv("CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH")
	if msg.BaseBranch != "" {
		baseBranch = msg.BaseBranch
	}

	var filesList []gitprovider.CommitFile
	for _, clusterName := range msg.ClusterNames {
		path := getClusterPathInRepo(clusterName)
		filesList = append(filesList, gitprovider.CommitFile{
			Path:    &path,
			Content: nil,
		})
	}

	if msg.HeadBranch == "" {
		clusters := strings.Join(msg.ClusterNames, "")
		msg.HeadBranch = getHash(msg.RepositoryUrl, clusters, msg.BaseBranch)
	}
	if msg.Title == "" {
		msg.Title = fmt.Sprintf("Gitops delete clusters: %s", msg.ClusterNames)
	}
	if msg.Description == "" {
		msg.Description = fmt.Sprintf("Pull request to delete clusters: %s", strings.Join(msg.ClusterNames, ", "))
	}
	if msg.CommitMessage == "" {
		msg.CommitMessage = "Remove Clusters Manifests"
	}
	_, err = s.provider.GetRepository(ctx, *gp, repositoryURL)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "failed to get repo %s: %s", repositoryURL, err)
	}

	var pullRequestURL string

	// FIXME: maybe this should reconcile rather than just try to create in case of other errors, e.g. database row creation
	res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider:       *gp,
		RepositoryURL:     repositoryURL,
		ReposistoryAPIURL: msg.RepositoryApiUrl,
		HeadBranch:        msg.HeadBranch,
		BaseBranch:        baseBranch,
		Title:             msg.Title,
		Description:       msg.Description,
		CommitMessage:     msg.CommitMessage,
		Files:             filesList,
	})
	if err != nil {
		s.log.Error(err, "Failed to create pull request")
		return nil, err
	}

	pullRequestURL = res.WebURL

	err = s.db.Transaction(func(tx *gorm.DB) error {
		pr := &models.PullRequest{
			URL:  pullRequestURL,
			Type: "delete",
		}
		if err := tx.Create(pr).Error; err != nil {
			return err
		}

		for _, clusterName := range msg.ClusterNames {
			var cluster models.Cluster
			if err := tx.Where("name = ?", clusterName).First(&cluster).Error; err != nil {
				return err
			}

			cluster.PullRequests = append(cluster.PullRequests, pr)
			if err := tx.Save(cluster).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &capiv1_proto.DeleteClustersPullRequestResponse{
		WebUrl: pullRequestURL,
	}, nil
}

// GetKubeconfig returns the Kubeconfig for the given workload cluster
func (s *server) GetKubeconfig(ctx context.Context, msg *capiv1_proto.GetKubeconfigRequest) (*httpbody.HttpBody, error) {
	var sec corev1.Secret
	name := fmt.Sprintf("%s-kubeconfig", msg.ClusterName)

	ns := os.Getenv("CAPI_CLUSTERS_NAMESPACE")
	if ns == "" {
		return nil, fmt.Errorf("environment variable %q cannot be empty", "CAPI_CLUSTERS_NAMESPACE")
	}

	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	key := client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}
	err = cl.Get(ctx, key, &sec)
	if err != nil {
		return nil, fmt.Errorf("unable to get secret %q for Kubeconfig: %w", name, err)
	}

	val, ok := sec.Data["value"]
	if !ok {
		return nil, fmt.Errorf("secret %q was found but is missing key %q", key, "value")
	}

	var acceptHeader string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if accept, ok := md["accept"]; ok {
			acceptHeader = strings.Join(accept, ",")
		}
	}

	if strings.Contains(acceptHeader, "application/octet-stream") {
		return &httpbody.HttpBody{
			ContentType: "application/octet-stream",
			Data:        val,
		}, nil
	}

	res, err := json.Marshal(&capiv1_proto.GetKubeconfigResponse{
		Kubeconfig: base64.StdEncoding.EncodeToString(val),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return &httpbody.HttpBody{
		ContentType: "application/json",
		Data:        res,
	}, nil
}
