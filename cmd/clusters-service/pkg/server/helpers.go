package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/mkmik/multierror"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func renderTemplateWithValues(t *capiv1.CAPITemplate, name string, values map[string]string) ([][]byte, error) {
	opts := []capi.RenderOptFunc{
		capi.InNamespace(os.Getenv("CAPI_CLUSTERS_NAMESPACE")),
	}
	if os.Getenv("INJECT_PRUNE_ANNOTATION") != "disabled" {
		opts = append(opts, capi.InjectPruneAnnotation)
	}

	templateBits, err := capi.Render(t.Spec, values, opts...)
	if err != nil {
		if missing, ok := isMissingVariableError(err); ok {
			return nil, fmt.Errorf("error rendering template %v due to missing variables: %s", name, missing)
		}
		return nil, fmt.Errorf("error rendering template %v, %v", name, err)
	}

	return templateBits, nil
}

func getHash(inputs ...string) string {
	final := []byte(strings.Join(inputs, ""))
	return fmt.Sprintf("wego-%x", md5.Sum(final))
}

func getToken(ctx context.Context) (string, string, error) {
	token := os.Getenv("GIT_PROVIDER_TOKEN")

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
		Type:      os.Getenv("GIT_PROVIDER_TYPE"),
		TokenType: tokenType,
		Token:     token,
		Hostname:  os.Getenv("GIT_PROVIDER_HOSTNAME"),
	}, nil
}

func (s *server) GetEnterpriseVersion(ctx context.Context, msg *capiv1_proto.GetEnterpriseVersionRequest) (*capiv1_proto.GetEnterpriseVersionResponse, error) {
	return &capiv1_proto.GetEnterpriseVersionResponse{
		Version: version.Version,
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

func isProviderRecognised(provider string) bool {
	for _, p := range providers {
		if strings.EqualFold(provider, p) {
			return true
		}
	}
	return false
}

func getProvider(t *capiv1.CAPITemplate) string {
	meta, err := capi.ParseTemplateMeta(t)

	if err != nil {
		return ""
	}

	for _, obj := range meta.Objects {
		if p, ok := providers[obj.Kind]; ok {
			return p
		}
	}

	return ""
}

func filterTemplatesByProvider(tl []*capiv1_proto.Template, provider string) []*capiv1_proto.Template {
	templates := []*capiv1_proto.Template{}

	for _, t := range tl {
		if strings.EqualFold(t.Provider, provider) {
			templates = append(templates, t)
		}
	}

	return templates
}

func validateDeleteClustersPR(msg *capiv1_proto.DeleteClustersPullRequestRequest) error {
	var err error

	if msg.ClusterNames == nil {
		err = multierror.Append(err, fmt.Errorf("at least one cluster name must be specified"))
	}

	return err
}

func getClusterPathInRepo(clusterName string) string {
	repositoryPath := os.Getenv("CAPI_REPOSITORY_PATH")
	if repositoryPath == "" {
		repositoryPath = DefaultRepositoryPath
	}
	return filepath.Join(repositoryPath, fmt.Sprintf("%s.yaml", clusterName))
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

func getProfilesFromTemplate(annotations map[string]string) ([]*capiv1_proto.TemplateProfile, error) {
	profiles := []*capiv1_proto.TemplateProfile{}
	profile := capiv1_proto.TemplateProfile{}

	for k, v := range annotations {
		if strings.Contains(k, "capi.weave.works/profile-") {
			err := json.Unmarshal([]byte(v), &profile)
			if err != nil {
				return profiles, fmt.Errorf("failed to unmarshal profiles: %w", err)
			}
			profiles = append(profiles, &profile)
		}
	}

	return profiles, nil
}

// getProfileLatestVersion returns the default profile values if not given
func getDefaultValues(ctx context.Context, kubeClient client.Client, name, version, helmRepositoryCacheDir string, sourceRef helmv2beta1.CrossNamespaceObjectReference, helmRepo *sourcev1beta1.HelmRepository) (string, error) {
	ref := &charts.ChartReference{Chart: name, Version: version, SourceRef: sourceRef}
	cc := charts.NewHelmChartClient(kubeClient, os.Getenv("RUNTIME_NAMESPACE"), helmRepo, charts.WithCacheDir(helmRepositoryCacheDir))
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
