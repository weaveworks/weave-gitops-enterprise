package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/validation"
)

const (
	capiClusterRef                      string = "CAPICluster"
	secretRef                           string = "Secret"
	HelmReleaseNamespace                       = "flux-system"
	deleteClustersRequiredErr                  = "at least one cluster must be specified"
	createClusterAutomationsRequiredErr        = "at least one cluster automation must be specified"
	kustomizationKind                          = "GitRepository"
)

var (
	labels = []string{}
)

type generateProfileFilesParams struct {
	helmRepositoryCluster types.NamespacedName
	helmRepository        types.NamespacedName
	chartsCache           helm.ChartsCacheReader
	profileValues         []*capiv1_proto.ProfileValues
	parameterValues       map[string]string
}

func (s *server) ListGitopsClusters(ctx context.Context, msg *capiv1_proto.ListGitopsClustersRequest) (*capiv1_proto.ListGitopsClustersResponse, error) {
	namespacedLists, err := s.managementFetcher.Fetch(ctx, "GitopsCluster", func() client.ObjectList {
		return &gitopsv1alpha1.GitopsClusterList{}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query clusters: %w", err)
	}

	clusters := []*capiv1_proto.GitopsCluster{}
	errors := []*capiv1_proto.ListError{}

	for _, namespacedList := range namespacedLists {
		if namespacedList.Error != nil {
			errors = append(errors, &capiv1_proto.ListError{
				Namespace: namespacedList.Namespace,
				Message:   namespacedList.Error.Error(),
			})
		}
		clustersList := namespacedList.List.(*gitopsv1alpha1.GitopsClusterList)
		for _, c := range clustersList.Items {
			clusters = append(clusters, ToClusterResponse(&c))
		}
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	if s.capiEnabled {
		clusters, err = AddCAPIClusters(ctx, client, clusters)
		if err != nil {
			return nil, err
		}
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
	mgmtCluster, err := getManagementCluster(s.cluster)
	if err != nil {
		return nil, err
	}

	clusters = append(clusters, mgmtCluster)

	sort.Slice(clusters, func(i, j int) bool { return clusters[i].Name < clusters[j].Name })
	return &capiv1_proto.ListGitopsClustersResponse{
		GitopsClusters: clusters,
		Total:          int32(len(clusters)),
		Errors:         errors,
	}, err
}

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1_proto.CreatePullRequestRequest) (*capiv1_proto.CreatePullRequestResponse, error) {
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}

	gp, err := getGitProvider(ctx, msg.RepositoryUrl)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	applyCreateClusterDefaults(msg)

	if err := validateCreateClusterPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, err
	}
	tmpl, err := s.getTemplate(ctx, msg.TemplateName, msg.TemplateNamespace, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}

	clusterNamespace := getClusterNamespace(msg.ParameterValues["NAMESPACE"])

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	// Get list of previous files to be added as deleted files in the commit,
	// Update the  previous values to be nil to skip including it in the updated create-request annotation
	prevFiles := &GetFilesReturn{}
	if msg.PreviousValues != nil {
		prevFiles, err = GetFiles(
			ctx,
			client,
			s.log,
			s.estimator,
			s.chartsCache,
			types.NamespacedName{Name: s.cluster},
			s.profileHelmRepository,
			tmpl,
			GetFilesRequest{clusterNamespace, msg.TemplateName, "CAPITemplate", msg.PreviousValues.ParameterValues, msg.PreviousValues.Credentials, msg.PreviousValues.Values, msg.PreviousValues.Kustomizations, msg.PreviousValues.ExternalSecrets},
			msg,
		)
		if err != nil {
			return nil, err
		}
		msg.PreviousValues = nil
	}

	git_files, err := GetFiles(
		ctx,
		client,
		s.log,
		s.estimator,
		s.chartsCache,
		types.NamespacedName{Name: s.cluster},
		s.profileHelmRepository,
		tmpl,
		GetFilesRequest{clusterNamespace, msg.TemplateName, "CAPITemplate", msg.ParameterValues, msg.Credentials, msg.Values, msg.Kustomizations, msg.ExternalSecrets},
		msg,
	)
	if err != nil {
		return nil, err
	}

	files := []gitprovider.CommitFile{}
	files = append(files, git_files.RenderedTemplate...)
	files = append(files, git_files.ProfileFiles...)
	files = append(files, git_files.KustomizationFiles...)
	files = append(files, git_files.ExternalSecretsFiles...)

	deletedFiles := getDeletedFiles(prevFiles, git_files)
	files = append(files, deletedFiles...)

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
	gp, err := getGitProvider(ctx, msg.RepositoryUrl)
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
			// Files in manifest path
			path := getClusterManifestPath(
				createNamespacedName(
					clusterNamespacedName.Name,
					getClusterNamespace(clusterNamespacedName.Namespace)),
			)
			filesList = append(filesList, gitprovider.CommitFile{
				Path:    &path,
				Content: nil,
			})

			// Files in cluster path
			clusterDirPath := getClusterDirPath(types.NamespacedName{
				Name:      clusterNamespacedName.Name,
				Namespace: getClusterNamespace(clusterNamespacedName.Namespace),
			})

			treeEntries, err := s.provider.GetTreeList(ctx, *gp, repositoryURL, baseBranch, clusterDirPath, true)
			if err != nil {
				return nil, fmt.Errorf("error getting list of trees in repo: %s@%s: %w", repositoryURL, baseBranch, err)
			}

			for _, treeEntry := range treeEntries {
				filesList = append(filesList, gitprovider.CommitFile{
					Path:    &treeEntry.Path,
					Content: nil,
				})
			}

		}
	} else {
		for _, clusterName := range msg.ClusterNames {
			//Files in manifest path
			path := getClusterManifestPath(
				createNamespacedName(clusterName, getClusterNamespace("")),
			)
			filesList = append(filesList, gitprovider.CommitFile{
				Path:    &path,
				Content: nil,
			})

			// Files in cluster path
			clusterDirPath := getClusterDirPath(types.NamespacedName{
				Name:      clusterName,
				Namespace: getClusterNamespace(""),
			})

			treeEntries, err := s.provider.GetTreeList(ctx, *gp, repositoryURL, baseBranch, clusterDirPath, true)
			if err != nil {
				return nil, fmt.Errorf("error getting list of trees in repo: %s@%s: %w", repositoryURL, baseBranch, err)
			}

			for _, treeEntry := range treeEntries {
				filesList = append(filesList, gitprovider.CommitFile{
					Path:    &treeEntry.Path,
					Content: nil,
				})
			}
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

func (s *server) kubeConfigForCluster(ctx context.Context, cluster types.NamespacedName) ([]byte, error) {
	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	gc := &gitopsv1alpha1.GitopsCluster{}
	err = cl.Get(ctx, cluster, gc)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitopsCluster %s: %w", cluster, err)
	}
	if gc.Spec.SecretRef != nil {
		secretRefName := client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      gc.Spec.SecretRef.Name,
		}
		sec, err := secretByName(ctx, cl, secretRefName)
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get secret for cluster %s: %w", cluster, err)
		}
		if sec == nil {
			return nil, fmt.Errorf("failed to load referenced secret %s for cluster %s", secretRefName, cluster)
		}
		val, ok := kubeConfigFromSecret(sec)
		if !ok {
			return nil, fmt.Errorf("secret %q was found but is missing key %q", secretRefName, "value")
		}
		return val, nil
	}

	userSecretName := client.ObjectKey{
		Namespace: getClusterNamespace(cluster.Namespace),
		Name:      fmt.Sprintf("%s-user-kubeconfig", cluster.Name),
	}
	sec, err := secretByName(ctx, cl, userSecretName)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get secret for cluster %s: %w", cluster, err)
	}
	if sec != nil {
		val, ok := kubeConfigFromSecret(sec)
		if !ok {
			return nil, fmt.Errorf("secret %q was found but is missing key %q", userSecretName, "value")
		}
		return val, nil
	}

	clusterSecretName := client.ObjectKey{
		Namespace: getClusterNamespace(cluster.Namespace),
		Name:      fmt.Sprintf("%s-kubeconfig", cluster.Name),
	}
	sec, err = secretByName(ctx, cl, clusterSecretName)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get secret for cluster %s: %w", cluster, err)
	}
	if sec != nil {
		val, ok := kubeConfigFromSecret(sec)
		if !ok {
			return nil, fmt.Errorf("secret %q was found but is missing key %q", clusterSecretName, "value")
		}
		return val, nil
	}
	return nil, fmt.Errorf("unable to get kubeconfig secret for cluster %s", cluster)
}

func secretByName(ctx context.Context, cl client.Client, name types.NamespacedName) (*corev1.Secret, error) {
	sec := &corev1.Secret{}
	err := cl.Get(ctx, name, sec)
	if err != nil {
		return nil, err
	}
	return sec, nil
}

// GetKubeconfig returns the Kubeconfig for the given workload cluster
func (s *server) GetKubeconfig(ctx context.Context, msg *capiv1_proto.GetKubeconfigRequest) (*httpbody.HttpBody, error) {
	val, err := s.kubeConfigForCluster(ctx, types.NamespacedName{Name: msg.ClusterName, Namespace: getClusterNamespace(msg.ClusterNamespace)})
	if err != nil {
		return nil, err
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
	commonKustomization := createKustomizationObject(&capiv1_proto.Kustomization{
		Metadata: &capiv1_proto.Metadata{
			Name:      "clusters-bases-kustomization",
			Namespace: "flux-system",
		},
		Spec: &capiv1_proto.KustomizationSpec{
			Path: filepath.Join(
				viper.GetString("capi-repository-clusters-path"),
				"bases",
			),
			SourceRef: &capiv1_proto.SourceRef{
				Name: "flux-system",
			},
		},
	})

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

func getGitProvider(ctx context.Context, repositoryURL string) (*git.GitProvider, error) {
	token, tokenType, err := getToken(ctx)
	if err != nil {
		return nil, err
	}

	// defaults from config
	repoType := viper.GetString("git-provider-type")
	repoHostname := viper.GetString("git-provider-hostname")

	// if user supplies a different gitrepo, derive the provider and the host etc from
	if repositoryURL != "" {
		repoURL, err := gitproviders.NewRepoURL(repositoryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse repository URL: %w", err)
		}

		// override defaults
		repoType = string(repoURL.Provider())
		repoHostname = repoURL.URL().Host
	}

	return &git.GitProvider{
		Type:      repoType,
		TokenType: tokenType,
		Token:     token,
		Hostname:  repoHostname,
	}, nil
}

func createProfileYAML(helmRepo *sourcev1.HelmRepository, helmReleases []*helmv2.HelmRelease, tmpl templatesv1.Template, defaultPath string) ([]capiv1_proto.CommitFile, error) {

	// FIXME: if there is a tmpl.GetSpec().helmRepositoryTemplate.path then use that for the HelmRepository

	// FIXME: for each of the helmReleases if there is a tmpl.GetSpect().charts[].template.path then use that for the HelmRelease
	// we can just match on the name of the chart
	// we should concat the helmreleases with common paths together somehow
	// this functions should return a list of unique filenames and content

	// OLD LOGIC that needs an update:

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
//
// FIXME: we should return multiple files from here
func generateProfileFiles(ctx context.Context, tmpl templatesv1.Template, cluster types.NamespacedName, kubeClient client.Client, args generateProfileFilesParams) ([]gitprovider.CommitFile, error) {
	helmRepo := &sourcev1.HelmRepository{}
	err := kubeClient.Get(ctx, args.helmRepository, helmRepo)
	if err != nil {
		return nil, fmt.Errorf("cannot find Helm repository %s/%s: %w", args.helmRepository.Namespace, args.helmRepository.Name, err)
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

	tmplProcessor, err := templates.NewProcessorForTemplate(tmpl)
	if err != nil {
		return nil, err
	}

	var installs []charts.ChartInstall

	requiredProfiles, err := getProfilesFromTemplate(tmpl)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve default profiles: %w", err)
	}

	profilesIndex := map[string]*capiv1_proto.ProfileValues{}
	for _, v := range args.profileValues {
		profilesIndex[v.Name] = v
	}

	requiredProfilesIndex := map[string]*capiv1_proto.TemplateProfile{}
	for _, v := range requiredProfiles {
		requiredProfilesIndex[v.Name] = v
	}

	// add and overwrite the required values of profilesIndex where necessary.
	for _, requiredProfile := range requiredProfiles {
		p := profilesIndex[requiredProfile.Name]
		// Required profile has been added by the user
		if p != nil {
			if !requiredProfile.Editable {
				p.Values = base64.StdEncoding.EncodeToString([]byte(requiredProfile.Values))
			}
			if p.Namespace == "" {
				p.Namespace = requiredProfile.Namespace
			}
			if p.Version == "" {
				p.Version = requiredProfile.Version
			}
			if p.Layer == "" {
				p.Layer = requiredProfile.Layer
			}
		} else {
			profilesIndex[requiredProfile.Name] = &capiv1_proto.ProfileValues{
				Name:      requiredProfile.Name,
				Version:   requiredProfile.Version,
				Values:    base64.StdEncoding.EncodeToString([]byte(requiredProfile.Values)),
				Namespace: requiredProfile.Namespace,
				Layer:     requiredProfile.Layer,
			}
		}
	}

	for _, v := range profilesIndex {
		// Check the version and if empty read the latest version from cache.
		if v.Version == "" {
			v.Version, err = args.chartsCache.GetLatestVersion(ctx, args.helmRepositoryCluster, args.helmRepository, v.Name)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve latest version of profile: %w", err)
			}
		}

		// Check the version and if empty read the layer from cache.
		if v.Layer == "" {
			v.Layer, err = args.chartsCache.GetLayer(ctx, args.helmRepositoryCluster, args.helmRepository, v.Name, v.Version)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve layer of profile: %w", err)
			}
		}

		values, err := renderValues(v, *tmplProcessor, args.parameterValues)
		if err != nil {
			return nil, fmt.Errorf("cannot get values for profile %s: %w", v.Name, err)
		}

		profileTemplate := []byte{}
		requiredProfile := requiredProfilesIndex[v.Name]
		if requiredProfile != nil {
			profileTemplate, err = tmplProcessor.Render([]byte(requiredProfile.ProfileTemplate), args.parameterValues)
			if err != nil {
				return nil, fmt.Errorf("cannot render spec of profile %s: %w", v.Name, err)
			}
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
			ProfileTemplate: string(profileTemplate),
			Layer:           v.Layer,
			Values:          values,
			Namespace:       v.Namespace,
		})
	}

	helmReleases, err := charts.MakeHelmReleasesInLayers(HelmReleaseNamespace, installs)
	if err != nil {
		return nil, fmt.Errorf("making helm releases for cluster %w", err)
	}
	defaultPath := getClusterProfilesPath(cluster)
	commitFiles, err := createProfileYAML(helmRepoTemplate, helmReleases, tmpl, defaultPath)
	if err != nil {
		return nil, err
	}

	return commitFiles, nil
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

func applyCreateClusterDefaults(msg *capiv1_proto.CreatePullRequestRequest) {
	for _, k := range msg.Kustomizations {
		if k != nil && k.Metadata != nil && k.Metadata.Namespace == "" {
			k.Metadata.Namespace = defaultAutomationNamespace
		}
	}
}

func validateCreateClusterPR(msg *capiv1_proto.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, errors.New("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, errors.New("parameter values must be specified"))
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

	for _, k := range msg.Kustomizations {
		err = multierror.Append(err, validateKustomization(k))
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

func validateKustomization(kustomization *capiv1_proto.Kustomization) error {
	var err error

	if kustomization.Metadata == nil {
		err = multierror.Append(err, errors.New("kustomization metadata must be specified"))
	} else {
		if kustomization.Metadata.Name == "" {
			err = multierror.Append(err, fmt.Errorf("kustomization name must be specified"))
		}

		invalidNamespaceErr := validateNamespace(kustomization.Metadata.Namespace)
		if invalidNamespaceErr != nil {
			err = multierror.Append(err, invalidNamespaceErr)
		}
	}

	if kustomization.Spec.SourceRef != nil {
		if kustomization.Spec.SourceRef.Name == "" {
			err = multierror.Append(
				err,
				fmt.Errorf("sourceRef name must be specified in Kustomization %s",
					kustomization.Metadata.Name))
		}

		invalidNamespaceErr := validateNamespace(kustomization.Spec.SourceRef.Namespace)
		if invalidNamespaceErr != nil {
			err = multierror.Append(err, invalidNamespaceErr)
		}
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

func getClusterDirPath(cluster types.NamespacedName) string {
	return filepath.Join(
		viper.GetString("capi-repository-clusters-path"),
		cluster.Namespace,
		cluster.Name,
	)
}

func getCommonKustomizationPath(cluster types.NamespacedName) string {
	return filepath.Join(
		getClusterDirPath(cluster),
		"clusters-bases-kustomization.yaml",
	)
}

func getClusterProfilesPath(cluster types.NamespacedName) string {
	return filepath.Join(
		getClusterDirPath(cluster),
		profiles.ManifestFileName,
	)
}

// renderValues renders the "values.yaml" section of a HelmRelease, as it can also contain template parameters.
func renderValues(v *capiv1_proto.ProfileValues, tmplProcessor templates.TemplateProcessor, parameterValues map[string]string) (map[string]interface{}, error) {
	// FIXME: look into decoding the base64 in the proto API rather than here.
	decoded, err := base64.StdEncoding.DecodeString(v.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode values: %w", err)
	}

	data, err := tmplProcessor.Render(decoded, parameterValues)
	if err != nil {
		return nil, fmt.Errorf("failed to render values for profile %s/%s: %w", v.Name, v.Version, err)
	}

	parsed, err := ParseValues(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse values for profile %s/%s: %w", v.Name, v.Version, err)
	}

	return parsed, nil
}

// ParseValues takes a YAML encoded values string and returns a struct
func ParseValues(v []byte) (map[string]interface{}, error) {
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
func getManagementCluster(name string) (*capiv1_proto.GitopsCluster, error) {
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

func generateKustomizationFile(
	ctx context.Context,
	isControlPlane bool,
	cluster types.NamespacedName,
	kubeClient client.Client,
	kustomization *capiv1_proto.Kustomization,
	filePath string) (gitprovider.CommitFile, error) {
	kustomizationYAML := createKustomizationObject(kustomization)

	b, err := yaml.Marshal(kustomizationYAML)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("error marshalling %s kustomization, %w", kustomization.Metadata.Name, err)
	}

	k := createNamespacedName(kustomization.Metadata.Name, kustomization.Metadata.Namespace)

	kustomizationPath := getClusterResourcePath(isControlPlane, "kustomization", cluster, k)
	if filePath != "" {
		kustomizationPath = filePath
	}

	kustomizationContent := string(b)

	file := &gitprovider.CommitFile{
		Path:    &kustomizationPath,
		Content: &kustomizationContent,
	}

	return *file, nil
}

func getClusterResourcePath(isControlPlane bool, resourceType string, cluster, resource types.NamespacedName) string {
	var clusterNamespace string
	if !isControlPlane {
		clusterNamespace = cluster.Namespace
	}

	fileName := fmt.Sprintf("%s-%s-%s.yaml", resource.Name, resource.Namespace, resourceType)

	if resourceType == "namespace" {
		fileName = fmt.Sprintf("%s-%s.yaml", resource.Name, resourceType)
	}

	if resourceType == "externalsecret" {
		return filepath.Join(
			viper.GetString("capi-repository-clusters-path"),
			clusterNamespace,
			cluster.Name,
			"secrets",
			fileName,
		)
	}

	return filepath.Join(
		viper.GetString("capi-repository-clusters-path"),
		clusterNamespace,
		cluster.Name,
		fileName,
	)
}

func createKustomizationObject(kustomization *capiv1_proto.Kustomization) *kustomizev1.Kustomization {
	generatedKustomization := &kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kustomization.Metadata.Name,
			Namespace: kustomization.Metadata.Namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:      kustomizationKind,
				Name:      kustomization.Spec.SourceRef.Name,
				Namespace: kustomization.Spec.SourceRef.Namespace,
			},
			Interval:        metav1.Duration{Duration: time.Minute * 10},
			Prune:           true,
			Path:            kustomization.Spec.Path,
			TargetNamespace: kustomization.Spec.TargetNamespace,
		},
	}

	return generatedKustomization
}

func kubeConfigFromSecret(s *corev1.Secret) ([]byte, bool) {
	val, ok := s.Data["value.yaml"]
	if ok {
		return val, true
	}
	val, ok = s.Data["value"]
	if ok {
		return val, true
	}
	return nil, false
}

func createNamespacedName(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

// Get list of gitprovider.CommitFile objects of files that should be deleted with empty content
// Kustomizations and Profiles removed during an edit are added to the deleted list
// Old files with changed paths are added to the deleted list
func getDeletedFiles(prevFiles *GetFilesReturn, newFiles *GetFilesReturn) []gitprovider.CommitFile {
	deletedFiles := []gitprovider.CommitFile{}

	removedKustomizations := getMissingFiles(prevFiles.KustomizationFiles, newFiles.KustomizationFiles)
	removedProfiles := getMissingFiles(prevFiles.ProfileFiles, newFiles.ProfileFiles)
	removedRenderedTemplates := getMissingFiles(prevFiles.RenderedTemplate, newFiles.RenderedTemplate)

	deletedFiles = append(deletedFiles, removedKustomizations...)
	deletedFiles = append(deletedFiles, removedProfiles...)
	deletedFiles = append(deletedFiles, removedRenderedTemplates...)

	return deletedFiles
}
