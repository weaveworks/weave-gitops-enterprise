package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/mkmik/multierror"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	common_utils "github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

	repositoryURL := viper.GetString("capi-templates-repository-url")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := viper.GetString("capi-templates-repository-base-branch")
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
			viper.GetString("runtime-namespace"),
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

	repositoryURL := viper.GetString("capi-templates-repository-url")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := viper.GetString("capi-templates-repository-base-branch")
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

	ns := viper.GetString("capi-clusters-namespace")
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

func getHash(inputs ...string) string {
	final := []byte(strings.Join(inputs, ""))
	return fmt.Sprintf("wego-%x", md5.Sum(final))
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

func getGitProvider(ctx context.Context) (*git.GitProvider, error) {
	token, tokenType, err := getToken(ctx)
	if err != nil {
		return nil, err
	}

	return &git.GitProvider{
		Type:      viper.GetString("git-provider-type"),
		TokenType: tokenType,
		Token:     token,
		Hostname:  viper.GetString("git-provider-hostname"),
	}, nil
}

func createProfileYAML(helmRepo *sourcev1beta1.HelmRepository, helmReleases []*helmv2beta1.HelmRelease) ([]byte, error) {
	out := [][]byte{}
	// Add HelmRepository object
	b, err := yaml.Marshal(helmRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HelmRepository object to YAML: %w", err)
	}
	out = append(out, b)
	// Add HelmRelease objects
	for _, v := range helmReleases {
		b, err := yaml.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HelmRelease object to YAML: %w", err)
		}
		out = append(out, b)
	}

	return bytes.Join(out, []byte("---\n")), nil
}

// generateProfileFiles to create a HelmRelease object with the profile and values.
// profileValues is what the client will provide to the API.
// It may have > 1 and its values parameter may be empty.
// Assumption: each profile should have a values.yaml that we can treat as the default.
func generateProfileFiles(ctx context.Context, helmRepoName, helmRepoNamespace, helmRepositoryCacheDir, clusterName string, kubeClient client.Client, profileValues []*capiv1_proto.ProfileValues) (*gitprovider.CommitFile, error) {
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := kubeClient.Get(ctx, client.ObjectKey{
		Name:      helmRepoName,
		Namespace: helmRepoNamespace,
	}, helmRepo)
	if err != nil {
		return nil, fmt.Errorf("cannot find Helm repository: %w", err)
	}
	helmRepoTemplate := &sourcev1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta1.HelmRepositoryKind,
			APIVersion: sourcev1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmRepoName,
			Namespace: helmRepoNamespace,
		},
		Spec: helmRepo.Spec,
	}

	sourceRef := helmv2beta1.CrossNamespaceObjectReference{
		APIVersion: helmRepo.TypeMeta.APIVersion,
		Kind:       helmRepo.TypeMeta.Kind,
		Name:       helmRepo.ObjectMeta.Name,
		Namespace:  helmRepo.ObjectMeta.Namespace,
	}

	var profileName string
	var installs []charts.ChartInstall
	for _, v := range profileValues {
		// Check the values and if empty use profile defaults. This should happen before parsing.
		if v.Values == "" {
			v.Values, err = getDefaultValues(ctx, kubeClient, v.Name, v.Version, helmRepositoryCacheDir, sourceRef, helmRepo)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve default values of profile: %w", err)
			}
		}

		// Check the version and if empty use thr latest version in profile defaults.
		if v.Version == "" {
			v.Version, err = getProfileLatestVersion(ctx, v.Name, helmRepo)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve latest version of profile: %w", err)
			}
		}

		parsed, err := parseValues(v.Values)
		if err != nil {
			return nil, fmt.Errorf("failed to parse values for profile %s/%s: %w", v.Name, v.Version, err)
		}
		installs = append(installs, charts.ChartInstall{
			Ref: charts.ChartReference{
				Chart:   v.Name,
				Version: v.Version,
				SourceRef: helmv2beta1.CrossNamespaceObjectReference{
					Name:      helmRepo.GetName(),
					Namespace: helmRepo.GetNamespace(),
					Kind:      "HelmRepository",
				},
			},
			Layer:  v.Layer,
			Values: parsed,
		})

	}

	helmReleases, err := charts.MakeHelmReleasesInLayers(clusterName, "wego-system", installs)
	if err != nil {
		return nil, fmt.Errorf("making helm releases for cluster %s: %w", clusterName, err)
	}
	c, err := createProfileYAML(helmRepoTemplate, helmReleases)
	if err != nil {
		return nil, err
	}
	profilePath := fmt.Sprintf(".weave-gitops/clusters/%s/system/%s.yaml", clusterName, profileName)
	profileContent := string(c)
	file := &gitprovider.CommitFile{
		Path:    &profilePath,
		Content: &profileContent,
	}

	return file, nil
}

func validateCreateClusterPR(msg *capiv1_proto.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	return err
}

func validateDeleteClustersPR(msg *capiv1_proto.DeleteClustersPullRequestRequest) error {
	var err error

	if msg.ClusterNames == nil {
		err = multierror.Append(err, fmt.Errorf("at least one cluster name must be specified"))
	}

	return err
}

func getClusterPathInRepo(clusterName string) string {
	repositoryPath := viper.GetString("capi-repository-path")
	if repositoryPath == "" {
		repositoryPath = DefaultRepositoryPath
	}
	return filepath.Join(repositoryPath, fmt.Sprintf("%s.yaml", clusterName))
}

// getProfileLatestVersion returns the default profile values if not given
func getDefaultValues(ctx context.Context, kubeClient client.Client, name, version, helmRepositoryCacheDir string, sourceRef helmv2beta1.CrossNamespaceObjectReference, helmRepo *sourcev1beta1.HelmRepository) (string, error) {
	ref := &charts.ChartReference{Chart: name, Version: version, SourceRef: sourceRef}
	cc := charts.NewHelmChartClient(kubeClient, viper.GetString("runtime-namespace"), helmRepo, charts.WithCacheDir(helmRepositoryCacheDir))
	if err := cc.UpdateCache(ctx); err != nil {
		return "", fmt.Errorf("failed to update Helm cache: %w", err)
	}
	bs, err := cc.FileFromChart(ctx, ref, chartutil.ValuesfileName)
	if err != nil {
		return "", fmt.Errorf("cannot retrieve values file from Helm chart %q: %w", ref, err)
	}
	// Base64 encode the content of values.yaml and assign it
	values := base64.StdEncoding.EncodeToString(bs)

	return values, nil
}

// getProfileLatestVersion returns the latest profile version if not given
func getProfileLatestVersion(ctx context.Context, name string, helmRepo *sourcev1beta1.HelmRepository) (string, error) {
	ps, err := charts.ScanCharts(ctx, helmRepo, charts.Profiles)
	version := ""
	if err != nil {
		return "", fmt.Errorf("cannot scan for profiles: %w", err)
	}

	for _, p := range ps {
		if p.Name == name {
			version = p.AvailableVersions[len(p.AvailableVersions)-1]
		}
	}

	return version, nil
}

func parseValues(s string) (map[string]interface{}, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode values: %w", err)
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(decoded, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values from JSON: %w", err)
	}
	return vals, nil
}
