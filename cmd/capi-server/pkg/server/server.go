package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/mkmik/multierror"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/credentials"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/git"
	capiv1_proto "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/wks/common/database/models"
	common_utils "github.com/weaveworks/wks/common/database/utils"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type server struct {
	library         templates.Library
	provider        git.Provider
	client          client.Client
	discoveryClient discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	db *gorm.DB
	ns string // The namespace where cluster objects reside
}

func NewClusterServer(library templates.Library, provider git.Provider, client client.Client, discoveryClient discovery.DiscoveryInterface, db *gorm.DB, ns string) capiv1_proto.ClustersServiceServer {
	return &server{library: library, provider: provider, client: client, discoveryClient: discoveryClient, db: db, ns: ns}
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1_proto.ListTemplatesRequest) (*capiv1_proto.ListTemplatesResponse, error) {
	tl, err := s.library.List(ctx)
	if err != nil {
		return nil, err
	}
	templates := []*capiv1_proto.Template{}

	for _, t := range tl {
		templates = append(templates, ToTemplateResponse(t))
	}

	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	return &capiv1_proto.ListTemplatesResponse{Templates: templates, Total: int32(len(tl))}, err
}

func (s *server) GetTemplate(ctx context.Context, msg *capiv1_proto.GetTemplateRequest) (*capiv1_proto.GetTemplateResponse, error) {
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	t := ToTemplateResponse(tm)
	if t.Error != "" {
		return nil, fmt.Errorf("error reading template %v, %v", msg.TemplateName, t.Error)
	}
	return &capiv1_proto.GetTemplateResponse{Template: t}, err
}

func (s *server) ListTemplateParams(ctx context.Context, msg *capiv1_proto.ListTemplateParamsRequest) (*capiv1_proto.ListTemplateParamsResponse, error) {
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	t := ToTemplateResponse(tm)
	if t.Error != "" {
		return nil, fmt.Errorf("error looking up template params for %v, %v", msg.TemplateName, t.Error)
	}

	return &capiv1_proto.ListTemplateParamsResponse{Parameters: t.Parameters, Objects: t.Objects}, err
}

func (s *server) RenderTemplate(ctx context.Context, msg *capiv1_proto.RenderTemplateRequest) (*capiv1_proto.RenderTemplateResponse, error) {
	log.WithFields(log.Fields{
		"request_values":      msg.Values,
		"request_credentials": msg.Credentials,
	}).Info("Received message")
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	templateBits, err := capi.Render(tm.Spec, msg.Values)
	if err != nil {
		if missing, ok := isMissingVariableError(err); ok {
			return nil, fmt.Errorf("error rendering template %v due to missing variables: %s", msg.TemplateName, missing)
		}
		return nil, fmt.Errorf("error rendering template %v, %v", msg.TemplateName, err)
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.client, templateBits, msg.Credentials, msg.TemplateName)
	if err != nil {
		return nil, err
	}

	resultStr := string(tmplWithValuesAndCredentials[:])

	return &capiv1_proto.RenderTemplateResponse{RenderedTemplate: resultStr}, err
}

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1_proto.CreatePullRequestRequest) (*capiv1_proto.CreatePullRequestResponse, error) {
	if err := validateCreateClusterPR(msg); err != nil {
		log.WithError(err).Errorf("Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}
	tmplWithValues, err := capi.Render(tmpl.Spec, msg.ParameterValues)
	if err != nil {
		return nil, fmt.Errorf("unable to render template %q: %w", msg.TemplateName, err)
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.client, tmplWithValues, msg.Credentials, msg.TemplateName)
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
		log.Warn("Couldn't find NAMESPACE param in request, using 'default'.")
		// TODO: https://weaveworks.atlassian.net/browse/WKP-2205
		clusterNamespace = "default"
	}

	content := string(tmplWithValuesAndCredentials[:])

	repositoryURL := os.Getenv("CAPI_TEMPLATES_REPOSITORY_URL")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := os.Getenv("CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH")
	if msg.BaseBranch != "" {
		baseBranch = msg.BaseBranch
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

		path := getClusterPathInRepo(clusterName)
		// FIXME: maybe this should reconcile rather than just try to create in case of other errors, e.g. database row creation
		res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
			GitProvider: git.GitProvider{
				Type:     os.Getenv("GIT_PROVIDER_TYPE"),
				Token:    os.Getenv("GIT_PROVIDER_TOKEN"),
				Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
			},
			RepositoryURL: repositoryURL,
			HeadBranch:    msg.HeadBranch,
			BaseBranch:    baseBranch,
			Title:         msg.Title,
			Description:   msg.Description,
			CommitMessage: msg.CommitMessage,
			Files: []gitprovider.CommitFile{
				{
					Path:    &path,
					Content: &content,
				},
			},
		})
		if err != nil {
			log.WithError(err).Errorf("Failed to create pull request")
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

		prCluster := &models.PRCluster{
			PRID:      pr.ID,
			ClusterID: c.ID,
		}
		if err := tx.Create(prCluster).Error; err != nil {
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

func validateCreateClusterPR(msg *capiv1_proto.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	if msg.HeadBranch == "" {
		err = multierror.Append(err, fmt.Errorf("head branch must be specified"))
	}

	if msg.Title == "" {
		err = multierror.Append(err, fmt.Errorf("title must be specified"))
	}

	if msg.Description == "" {
		err = multierror.Append(err, fmt.Errorf("description must be specified"))
	}

	if msg.CommitMessage == "" {
		err = multierror.Append(err, fmt.Errorf("commit message must be specified"))
	}

	return err
}

// ListCredentials searches the management cluster and lists any objects that match specific given types
func (s *server) ListCredentials(ctx context.Context, msg *capiv1_proto.ListCredentialsRequest) (*capiv1_proto.ListCredentialsResponse, error) {
	creds := []*capiv1_proto.Credential{}
	foundCredentials, err := credentials.FindCredentials(ctx, s.client, s.discoveryClient)
	if err != nil {
		return nil, err
	}

	for _, identity := range foundCredentials {
		creds = append(creds, &capiv1_proto.Credential{
			Group:     identity.GroupVersionKind().Group,
			Version:   identity.GroupVersionKind().Version,
			Kind:      identity.GetKind(),
			Name:      identity.GetName(),
			Namespace: identity.GetNamespace(),
		})
	}

	return &capiv1_proto.ListCredentialsResponse{Credentials: creds, Total: int32(len(creds))}, nil
}

// GetKubeconfig returns the Kubeconfig for the given workload cluster
func (s *server) GetKubeconfig(ctx context.Context, msg *capiv1_proto.GetKubeconfigRequest) (*httpbody.HttpBody, error) {
	var sec corev1.Secret
	key := client.ObjectKey{
		Namespace: s.ns,
		Name:      fmt.Sprintf("%s-kubeconfig", msg.ClusterName),
	}
	err := s.client.Get(ctx, key, &sec)
	if err != nil {
		return nil, fmt.Errorf("unable to get secret %q for Kubeconfig: %w", key, err)
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

func (s *server) DeleteClustersPullRequest(ctx context.Context, msg *capiv1_proto.DeleteClustersPullRequestRequest) (*capiv1_proto.DeleteClustersPullRequestResponse, error) {
	if err := validateDeleteClustersPR(msg); err != nil {
		log.WithError(err).Errorf("Failed to create pull request, message payload was invalid")
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

	var pullRequestURL string

	// FIXME: maybe this should reconcile rather than just try to create in case of other errors, e.g. database row creation
	res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: git.GitProvider{
			Type:     os.Getenv("GIT_PROVIDER_TYPE"),
			Token:    os.Getenv("GIT_PROVIDER_TOKEN"),
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: repositoryURL,
		HeadBranch:    msg.HeadBranch,
		BaseBranch:    baseBranch,
		Title:         msg.Title,
		Description:   msg.Description,
		CommitMessage: msg.CommitMessage,
		Files:         filesList,
	})
	if err != nil {
		log.WithError(err).Errorf("Failed to create pull request")
		return nil, err
	}

	pullRequestURL = res.WebURL

	pr := &models.PullRequest{
		URL:  pullRequestURL,
		Type: "delete",
	}
	if err := s.db.Create(pr).Error; err != nil {
		return nil, err
	}

	for _, clusterName := range msg.ClusterNames {
		var cluster models.Cluster
		s.db.Where("name = ?", clusterName).Find(&cluster)

		prCluster := &models.PRCluster{
			PRID:      pr.ID,
			ClusterID: cluster.ID,
		}
		if err := s.db.Create(prCluster).Error; err != nil {
			return nil, err
		}
	}

	return &capiv1_proto.DeleteClustersPullRequestResponse{
		WebUrl: pullRequestURL,
	}, nil
}

func validateDeleteClustersPR(msg *capiv1_proto.DeleteClustersPullRequestRequest) error {
	var err error

	if msg.ClusterNames == nil {
		err = multierror.Append(err, fmt.Errorf("at least one cluster name must be specified"))
	}

	if msg.HeadBranch == "" {
		err = multierror.Append(err, fmt.Errorf("head branch must be specified"))
	}

	if msg.Title == "" {
		err = multierror.Append(err, fmt.Errorf("title must be specified"))
	}

	if msg.Description == "" {
		err = multierror.Append(err, fmt.Errorf("description must be specified"))
	}

	if msg.CommitMessage == "" {
		err = multierror.Append(err, fmt.Errorf("commit message must be specified"))
	}

	return err
}

func getClusterPathInRepo(clusterName string) string {
	return fmt.Sprintf("management/%s.yaml", clusterName)
}

func isMissingVariableError(err error) (string, bool) {
	errStr := err.Error()
	prefix := "processing template: value for variables"
	suffix := "is not set. Please set the value using os environment variables or the clusterctl config file"
	if strings.HasPrefix(errStr, prefix) && strings.HasSuffix(errStr, suffix) {
		missing := strings.TrimSpace(errStr[len(prefix):strings.Index(errStr, suffix)])
		return missing, true
	}
	return "", false
}
