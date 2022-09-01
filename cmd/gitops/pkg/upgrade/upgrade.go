package upgrade

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/names"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"helm.sh/helm/v3/pkg/strvals"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpgradeValues struct {
	ConfigRepo    string
	ClusterPath   string
	Version       string
	BaseBranch    string
	HeadBranch    string
	CommitMessage string
	Namespace     string
	Values        []string
	DryRun        bool
}

const EnterpriseChartURL string = "https://charts.dev.wkp.weave.works/releases/charts-v3"
const CredentialsSecretName string = "weave-gitops-enterprise-credentials"
const WegoEnterpriseName string = "weave-gitops-enterprise.yaml"

func Upgrade(ctx context.Context, kubeClient *kube.KubeHTTP, gitClient git.Git, gitProvider gitproviders.GitProvider, upgradeValues UpgradeValues, logger logger.Logger, w io.Writer) error {
	return upgrade(ctx, upgradeValues, kubeClient, gitClient, gitProvider, logger, w)
}

func upgrade(ctx context.Context, uv UpgradeValues, kubeClient *kube.KubeHTTP, gitClient git.Git, gitProvider gitproviders.GitProvider, logger logger.Logger, w io.Writer) error {
	cname := kubeClient.ClusterName

	resources, err := makeHelmResources(uv.Namespace, uv.Version, cname, uv.ConfigRepo, uv.Values)
	if err != nil {
		return fmt.Errorf("error creating helm resources: %w", err)
	}

	appResources, err := makeAppsCapiKustomization(uv.Namespace, uv.ConfigRepo)
	if err != nil {
		return fmt.Errorf("error creating app resources: %w", err)
	}

	resources = append(resources, appResources...)

	out, err := marshalToYamlStream(resources)
	if err != nil {
		return fmt.Errorf("error marshalling helm resources: %w", err)
	}

	stringOut := string(out)

	if uv.DryRun {
		_, _ = w.Write([]byte(stringOut + "\n"))
		return nil
	}

	err = getBasicAuth(ctx, kubeClient, uv.Namespace)
	if err != nil {
		return fmt.Errorf("failed to load credentials for profiles repo from cluster: %v", err)
	}

	normalizedURL, err := gitproviders.NewRepoURL(uv.ConfigRepo)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %q: %w", uv.ConfigRepo, err)
	}

	configBranch := uv.BaseBranch
	if configBranch == "" {
		configBranch, err = gitProvider.GetDefaultBranch(ctx, normalizedURL)
		if err != nil {
			return fmt.Errorf("could not determine default branch for config repository: %q %w", uv.ConfigRepo, err)
		}
	}

	remover, repoDir, err := gitrepo.CloneRepo(ctx, gitClient, normalizedURL, configBranch)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	err = gitClient.Checkout(uv.HeadBranch)
	if err != nil {
		return fmt.Errorf("failed to create new branch %s: %w", uv.HeadBranch, err)
	}

	err = upgradeGitManifests(gitClient, repoDir, uv.ClusterPath, cname, stringOut, logger)
	if err != nil {
		return fmt.Errorf("failed to write update manifest in clone repo: %w", err)
	}

	err = gitrepo.CommitAndPush(ctx, gitClient, uv.CommitMessage, logger)
	if err != nil {
		return fmt.Errorf("failed to commit and push: %w", err)
	}

	pri := gitproviders.PullRequestInfo{
		Title:                     uv.CommitMessage,
		Description:               "Pull request to upgrade to Weave GitOps Enterprise",
		SkipAddingFilesOnCreation: true,
		TargetBranch:              configBranch,
		NewBranch:                 uv.HeadBranch,
	}

	pr, err := gitProvider.CreatePullRequest(ctx, normalizedURL, pri)
	if err != nil {
		return err
	}

	logger.Successf("Pull Request created: %s", pr.Get().WebURL)

	return nil
}

func marshalToYamlStream(objects []runtime.Object) ([]byte, error) {
	out := [][]byte{}

	for _, obj := range objects {
		b, err := yaml.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HelmRepository object to YAML: %w", err)
		}

		out = append(out, b)
	}

	return bytes.Join(out, []byte("---\n")), nil
}

func makeAppsCapiKustomization(namespace, repoURL string) ([]runtime.Object, error) {
	normalizedURL, err := gitproviders.NewRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL %q: %w", repoURL, err)
	}

	gitRepositoryName := names.CreateClusterSourceName(normalizedURL)

	appsCapiKustomization := &kustomizev2.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apps-capi",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev2.KustomizationKind,
			APIVersion: kustomizev2.GroupVersion.String(),
		},
		Spec: kustomizev2.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Path:     "./.weave-gitops/apps/capi",
			Prune:    true,
			SourceRef: kustomizev2.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: gitRepositoryName,
			},
		},
	}

	return []runtime.Object{appsCapiKustomization}, nil
}

func makeHelmResources(namespace, version, clusterName, repoURL string, values []string) ([]runtime.Object, error) {
	helmRepository := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-charts",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		Spec: sourcev1.HelmRepositorySpec{
			Interval: metav1.Duration{Duration: time.Minute},
			URL:      EnterpriseChartURL,
			SecretRef: &meta.LocalObjectReference{
				Name: CredentialsSecretName,
			},
		},
	}

	// default helmrelease values
	base := map[string]interface{}{
		"config": map[string]interface{}{
			"cluster": map[string]interface{}{
				"name": clusterName,
			},
			"capi": map[string]interface{}{
				"repositoryURL": repoURL,
			},
		},
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %w", err)
		}
	}

	valuesJson, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("error marshalling values object: %w", err)
	}

	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "mccp",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      helmRepository.GetName(),
						Namespace: helmRepository.GetNamespace(),
					},
					Version: version,
				},
			},
			Values: &v1.JSON{Raw: valuesJson},
		},
	}

	return []runtime.Object{helmRepository, helmRelease}, nil
}

func upgradeGitManifests(gitClient git.Git, repoDir, clusterPath, cname, wegoEnterpriseManifests string, logger logger.Logger) error {
	wegoEnterprisePath := filepath.Join(clusterPath, WegoEnterpriseName)

	if err := gitClient.Write(wegoEnterprisePath, []byte(wegoEnterpriseManifests)); err != nil {
		return fmt.Errorf("failed to write out manifest to %s: %w", wegoEnterprisePath, err)
	}

	configRoot := clusterPath
	if !strings.HasPrefix(configRoot, "/") {
		configRoot = filepath.Join(repoDir, configRoot)
	}

	findResult := utils.FindCoreConfig(configRoot)

	switch findResult.Status {
	case utils.Valid:
		relativePath := strings.TrimPrefix(findResult.Path, repoDir)

		err := gitClient.Remove(relativePath)
		if err != nil {
			logger.Warningf("Failed to remove existing core configuration: %q, skipping.\n", relativePath)
		}

	case utils.Embedded:
		logger.Warningf("The core configuration manifests are contained in a file: %q that also contains other manifests, skipping deletion.\n", findResult.Path)

	case utils.Partial:
		logger.Warningf("The core configuration manifests are not contained in a single file; skipping deletion.\n")

	case utils.Missing:
		logger.Warningf("Could not find existing core configuration manifests.\n")
	}

	return nil
}

func getBasicAuth(ctx context.Context, kubeClient *kube.KubeHTTP, ns string) error {
	deployKeySecret := &corev1.Secret{}

	err := kubeClient.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      CredentialsSecretName,
	}, deployKeySecret)
	if err != nil {
		return fmt.Errorf("failed to get entitlement: %v", err)
	}

	username := string(deployKeySecret.Data["username"])
	password := string(deployKeySecret.Data["password"])

	if username == "" || password == "" {
		return errors.New("username or password missing in entitlement secret, may be an old entitlement file")
	}

	return nil
}
