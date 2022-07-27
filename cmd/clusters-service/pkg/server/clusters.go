package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/mkmik/multierror"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	templatesv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"k8s.io/apimachinery/pkg/api/validation"
)

const (
	capiClusterRef            string = "CAPICluster"
	secretRef                 string = "Secret"
	HelmReleaseNamespace             = "flux-system"
	deleteClustersRequiredErr        = "at least one cluster must be specified"
)

var (
	labels = []string{}
)

type generateProfileFilesParams struct {
	helmRepository         types.NamespacedName
	helmRepositoryCacheDir string
	profileValues          []*capiv1_proto.ProfileValues
	parameterValues        map[string]string
}

func (s *server) ListGitopsClusters(ctx context.Context, msg *capiv1_proto.ListGitopsClustersRequest) (*capiv1_proto.ListGitopsClustersResponse, error) {
	listOptions := client.ListOptions{
		Limit:    msg.GetPageSize(),
		Continue: msg.GetPageToken(),
	}
	cl, nextPageToken, err := s.clustersLibrary.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	clusters := []*capiv1_proto.GitopsCluster{}

	for _, c := range cl {
		clusters = append(clusters, ToClusterResponse(c))
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	clusters, err = AddCAPIClusters(ctx, client, clusters)
	if err != nil {
		return nil, err
	}

	if msg.Label != "" {
		if !isLabelRecognised(msg.Label) {
			return nil, fmt.Errorf("label %q is not recognised", msg.Label)
		}

		clusters = filterClustersByLabel(clusters, msg.Label)
	}

	if msg.RefType != "" {
		clusters, err = filterClustersByType(clusters, msg.RefType)
		if err != nil {
			return nil, err
		}
	}

	// Append the management cluster to the end of clusters list
	mgmtCluster, err := getManagementCluster()
	if err != nil {
		return nil, err
	}

	clusters = append(clusters, mgmtCluster)

	sort.Slice(clusters, func(i, j int) bool { return clusters[i].Name < clusters[j].Name })
	return &capiv1_proto.ListGitopsClustersResponse{
		GitopsClusters: clusters,
		NextPageToken:  nextPageToken,
		Total:          int32(len(clusters))}, err
}

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1_proto.CreatePullRequestRequest) (*capiv1_proto.CreatePullRequestResponse, error) {
	gp, err := getGitProvider(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	if err := validateCreateClusterPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.templatesLibrary.Get(ctx, msg.TemplateName, "CAPITemplate")
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}

	clusterNamespace := getClusterNamespace(msg.ParameterValues["NAMESPACE"])
	tmplWithValues, err := renderTemplateWithValues(tmpl, msg.TemplateName, clusterNamespace, msg.ParameterValues)
	if err != nil {
		return nil, fmt.Errorf("failed to render template with parameter values: %w", err)
	}

	err = templates.ValidateRenderedTemplates(tmplWithValues)
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
	cluster := types.NamespacedName{
		Name:      clusterName,
		Namespace: clusterNamespace,
	}

	content := string(tmplWithValuesAndCredentials[:])
	path := getClusterManifestPath(cluster)
	files := []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}

	if viper.GetString("add-bases-kustomization") == "enabled" {
		commonKustomization, err := getCommonKustomization(cluster)
		if err != nil {
			return nil, fmt.Errorf("failed to get common kustomization for %s: %s", clusterName, err)
		}
		files = append(files, *commonKustomization)
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
			tmpl,
			cluster,
			client,
			generateProfileFilesParams{
				helmRepository: types.NamespacedName{
					Name:      s.profileHelmRepositoryName,
					Namespace: viper.GetString("runtime-namespace"),
				},
				helmRepositoryCacheDir: s.helmRepositoryCacheDir,
				profileValues:          msg.Values,
				parameterValues:        msg.ParameterValues,
			},
		)
		if err != nil {
			return nil, err
		}
		files = append(files, *profilesFile)
	}

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
		return nil, fmt.Errorf("unable to create pull request for %q: %w", msg.TemplateName, err)
	}

	return &capiv1_proto.CreatePullRequestResponse{
		WebUrl: res.WebURL,
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
	if len(msg.ClusterNamespacedNames) > 0 {
		for _, clusterNamespacedName := range msg.ClusterNamespacedNames {
			path := getClusterManifestPath(
				types.NamespacedName{
					Name:      clusterNamespacedName.Name,
					Namespace: getClusterNamespace(clusterNamespacedName.Namespace),
				},
			)
			filesList = append(filesList, gitprovider.CommitFile{
				Path:    &path,
				Content: nil,
			})
		}
	} else {
		for _, clusterName := range msg.ClusterNames {
			path := getClusterManifestPath(
				types.NamespacedName{
					Name:      clusterName,
					Namespace: getClusterNamespace(""),
				},
			)
			filesList = append(filesList, gitprovider.CommitFile{
				Path:    &path,
				Content: nil,
			})
		}
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

	return &capiv1_proto.DeleteClustersPullRequestResponse{
		WebUrl: res.WebURL,
	}, nil
}

// GetKubeconfig returns the Kubeconfig for the given workload cluster
func (s *server) GetKubeconfig(ctx context.Context, msg *capiv1_proto.GetKubeconfigRequest) (*httpbody.HttpBody, error) {
	var sec corev1.Secret
	name := fmt.Sprintf("%s-kubeconfig", msg.ClusterName)

	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	key := client.ObjectKey{
		Namespace: getClusterNamespace(msg.ClusterNamespace),
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

func getCommonKustomization(cluster types.NamespacedName) (*gitprovider.CommitFile, error) {

	commonKustomizationPath := getCommonKustomizationPath(cluster)
	commonKustomization := &kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clusters-bases-kustomization",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: "flux-system",
			},
			Interval: metav1.Duration{Duration: time.Minute * 10},
			Prune:    true,
			Path: filepath.Join(
				viper.GetString("capi-repository-clusters-path"),
				"bases",
			),
		},
	}
	b, err := yaml.Marshal(commonKustomization)
	if err != nil {
		return nil, fmt.Errorf("error marshalling common kustomization, %w", err)
	}
	commonKustomizationString := string(b)
	file := &gitprovider.CommitFile{
		Path:    &commonKustomizationPath,
		Content: &commonKustomizationString,
	}
	return file, nil
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

func createProfileYAML(helmRepo *sourcev1.HelmRepository, helmReleases []*helmv2.HelmRelease) ([]byte, error) {
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
func generateProfileFiles(ctx context.Context, tmpl *templatesv1.Template, cluster types.NamespacedName, kubeClient client.Client, args generateProfileFilesParams) (*gitprovider.CommitFile, error) {
	helmRepo := &sourcev1.HelmRepository{}
	err := kubeClient.Get(ctx, args.helmRepository, helmRepo)
	if err != nil {
		return nil, fmt.Errorf("cannot find Helm repository: %w", err)
	}
	helmRepoTemplate := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      args.helmRepository.Name,
			Namespace: args.helmRepository.Namespace,
		},
		Spec: helmRepo.Spec,
	}

	sourceRef := helmv2.CrossNamespaceObjectReference{
		APIVersion: helmRepo.TypeMeta.APIVersion,
		Kind:       helmRepo.TypeMeta.Kind,
		Name:       helmRepo.ObjectMeta.Name,
		Namespace:  helmRepo.ObjectMeta.Namespace,
	}

	tmplProcessor, err := templates.NewProcessorForTemplate(*tmpl)
	if err != nil {
		return nil, err
	}

	var installs []charts.ChartInstall
	for _, v := range args.profileValues {
		// Check the values and if not editable in the Template Profiles or empty, replace with default values. This should happen before parsing.
		if v.Values != "" {
			requiredProfiles, err := getProfilesFromTemplate(tmpl.Annotations)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve default profiles: %w", err)
			}
			for _, rp := range requiredProfiles {
				if rp.Version == v.Version && rp.Name == v.Name {
					if !rp.Editable {
						if rp.Values != "" {
							v.Values = rp.Values
						} else {
							v.Values, err = getDefaultValues(ctx, kubeClient, v.Name, v.Version, args.helmRepositoryCacheDir, sourceRef, helmRepo)
							if err != nil {
								return nil, fmt.Errorf("cannot retrieve default values of profile: %w", err)
							}
						}
					}
				}
			}
		} else {
			v.Values, err = getDefaultValues(ctx, kubeClient, v.Name, v.Version, args.helmRepositoryCacheDir, sourceRef, helmRepo)
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

		decoded, err := base64.StdEncoding.DecodeString(v.Values)
		if err != nil {
			return nil, fmt.Errorf("failed to base64 decode values: %w", err)
		}

		data, err := tmplProcessor.Render(decoded, args.parameterValues)
		if err != nil {
			return nil, fmt.Errorf("failed to render values for profile %s/%s: %w", v.Name, v.Version, err)
		}

		parsed, err := parseValues(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse values for profile %s/%s: %w", v.Name, v.Version, err)
		}
		installs = append(installs, charts.ChartInstall{
			Ref: charts.ChartReference{
				Chart:   v.Name,
				Version: v.Version,
				SourceRef: helmv2.CrossNamespaceObjectReference{
					Name:      helmRepo.GetName(),
					Namespace: helmRepo.GetNamespace(),
					Kind:      "HelmRepository",
				},
			},
			Layer:     v.Layer,
			Values:    parsed,
			Namespace: v.Namespace,
		})

	}

	helmReleases, err := charts.MakeHelmReleasesInLayers(cluster.Name, HelmReleaseNamespace, installs)
	if err != nil {
		return nil, fmt.Errorf("making helm releases for cluster %s: %w", cluster.Name, err)
	}
	c, err := createProfileYAML(helmRepoTemplate, helmReleases)
	if err != nil {
		return nil, err
	}

	profilePath := getClusterProfilesPath(cluster)
	profileContent := string(c)
	file := &gitprovider.CommitFile{
		Path:    &profilePath,
		Content: &profileContent,
	}

	return file, nil
}

func validateNamespace(namespace string) error {
	if namespace == "" {
		return nil
	}
	errs := validation.ValidateNamespaceName(namespace, false)
	if len(errs) != 0 {
		return fmt.Errorf("invalid namespace: %s, %s", namespace, strings.Join(errs, ","))
	}

	return nil
}

func validateCreateClusterPR(msg *capiv1_proto.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	invalidNamespaceErr := validateNamespace(msg.ParameterValues["NAMESPACE"])
	if invalidNamespaceErr != nil {
		err = multierror.Append(err, invalidNamespaceErr)
	}

	for i := range msg.Values {
		invalidNamespaceErr := validateNamespace(msg.Values[i].Namespace)
		if invalidNamespaceErr != nil {
			err = multierror.Append(err, invalidNamespaceErr)
		}
	}

	return err
}

func validateDeleteClustersPR(msg *capiv1_proto.DeleteClustersPullRequestRequest) error {
	var err error

	if len(msg.ClusterNamespacedNames) == 0 && len(msg.ClusterNames) == 0 {
		err = multierror.Append(err, fmt.Errorf(deleteClustersRequiredErr))
	}

	return err
}

func getClusterManifestPath(cluster types.NamespacedName) string {
	return filepath.Join(
		viper.GetString("capi-repository-path"),
		cluster.Namespace,
		fmt.Sprintf("%s.yaml", cluster.Name),
	)
}

func getCommonKustomizationPath(cluster types.NamespacedName) string {
	return filepath.Join(
		viper.GetString("capi-repository-clusters-path"),
		cluster.Namespace,
		cluster.Name,
		"clusters-bases-kustomization.yaml",
	)
}

func getClusterProfilesPath(cluster types.NamespacedName) string {
	return filepath.Join(
		viper.GetString("capi-repository-clusters-path"),
		cluster.Namespace,
		cluster.Name,
		profiles.ManifestFileName,
	)
}

// getProfileLatestVersion returns the default profile values if not given
func getDefaultValues(ctx context.Context, kubeClient client.Client, name, version, helmRepositoryCacheDir string, sourceRef helmv2.CrossNamespaceObjectReference, helmRepo *sourcev1.HelmRepository) (string, error) {
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
func getProfileLatestVersion(ctx context.Context, name string, helmRepo *sourcev1.HelmRepository) (string, error) {
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

func parseValues(v []byte) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(v, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values from JSON: %w", err)
	}
	return vals, nil
}

func isLabelRecognised(label string) bool {
	for _, l := range labels {
		if strings.EqualFold(label, l) {
			return true
		}
	}
	return false
}

func filterClustersByLabel(cl []*capiv1_proto.GitopsCluster, label string) []*capiv1_proto.GitopsCluster {
	clusters := []*capiv1_proto.GitopsCluster{}

	for _, c := range cl {
		for _, l := range c.Labels {
			if strings.EqualFold(l, label) {
				clusters = append(clusters, c)
			}
		}
	}

	return clusters
}

func filterClustersByType(cl []*capiv1_proto.GitopsCluster, refType string) ([]*capiv1_proto.GitopsCluster, error) {
	clusters := []*capiv1_proto.GitopsCluster{}

	for _, c := range cl {
		switch refType {
		case capiClusterRef:
			if c.CapiClusterRef != nil {
				clusters = append(clusters, c)
			}
		case secretRef:
			if c.SecretRef != nil {
				clusters = append(clusters, c)
			}
		default:
			return nil, fmt.Errorf("reference type %q is not recognised", refType)
		}
	}

	return clusters, nil
}

// getManagementCluster returns the management cluster as a gitops cluster
func getManagementCluster() (*capiv1_proto.GitopsCluster, error) {
	name := "management"

	cluster := &capiv1_proto.GitopsCluster{
		Name: name,
		Conditions: []*capiv1_proto.Condition{
			{
				Type:   "Ready",
				Status: "True",
			},
		},
		ControlPlane: true,
	}

	return cluster, nil
}
