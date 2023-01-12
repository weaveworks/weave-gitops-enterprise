package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/mkmik/multierror"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type GetAutomations struct {
	KustomizationFiles   []*capiv1_proto.CommitFile
	HelmReleaseFiles     []*capiv1_proto.CommitFile
	Clusters             []string
	ExternalSecretsFiles []*capiv1_proto.CommitFile
}

func toGitCommitFile(file *capiv1_proto.CommitFile) gitprovider.CommitFile {
	return gitprovider.CommitFile{
		Path:    &file.Path,
		Content: &file.Content,
	}
}

// CreateAutomationsPullRequest receives a list of {kustomization, helmrelease, cluster, externalsecret}
// generates a kustomization file and/or a helm release file for each provided cluster in the list
// and creates a pull request for the generated files
func (s *server) CreateAutomationsPullRequest(ctx context.Context, msg *capiv1_proto.CreateAutomationsPullRequestRequest) (*capiv1_proto.CreateAutomationsPullRequestResponse, error) {
	client, err := s.clientGetter.Client(ctx)

	if err != nil {
		return nil, err
	}

	automations, err := getAutomations(ctx, client, msg.ClusterAutomations)
	if err != nil {
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

	var files []gitprovider.CommitFile

	if len(automations.KustomizationFiles) > 0 {
		for _, f := range automations.KustomizationFiles {
			files = append(files, toGitCommitFile(f))
		}
	}

	if len(automations.HelmReleaseFiles) > 0 {
		for _, f := range automations.HelmReleaseFiles {
			files = append(files, toGitCommitFile(f))
		}
	}

	if len(automations.ExternalSecretsFiles) > 0 {
		for _, f := range automations.ExternalSecretsFiles {
			files = append(files, toGitCommitFile(f))
		}
	}

	if msg.HeadBranch == "" {
		clusters := strings.Join(automations.Clusters, "")
		msg.HeadBranch = getHash(msg.RepositoryUrl, clusters, msg.BaseBranch)
	}
	if msg.Title == "" {
		msg.Title = "Gitops add cluster workloads"
	}
	if msg.Description == "" {
		msg.Description = "Pull request to create cluster workloads"
	}
	if msg.CommitMessage == "" {
		msg.CommitMessage = "Add Kustomization Manifests"
	}

	gp, err := getGitProvider(ctx, msg.RepositoryUrl)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
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
		return nil, fmt.Errorf("unable to create pull request: %w", err)
	}

	return &capiv1_proto.CreateAutomationsPullRequestResponse{
		WebUrl: res.WebURL,
	}, nil
}

// RenderAutomation receives a list of {kustomization, helmrelease, cluster}
// generates a kustomization file and/or a helm release file for each provided cluster in the list
// and returns the generated files
func (s *server) RenderAutomation(ctx context.Context, msg *capiv1_proto.RenderAutomationRequest) (*capiv1_proto.RenderAutomationResponse, error) {
	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	automations, err := getAutomations(ctx, client, msg.ClusterAutomations)

	if err != nil {
		return nil, err
	}

	return &capiv1_proto.RenderAutomationResponse{KustomizationFiles: automations.KustomizationFiles, HelmReleaseFiles: automations.HelmReleaseFiles, ExternalSecretsFiles: automations.ExternalSecretsFiles}, err
}

func getAutomations(ctx context.Context, client client.Client, ca []*capiv1_proto.ClusterAutomation) (*GetAutomations, error) {
	applyCreateAutomationDefaults(ca)

	if err := validateAutomations(ca); err != nil {
		return nil, err
	}

	var clusters []string
	var kustomizationFiles []*capiv1_proto.CommitFile
	var helmReleaseFiles []*capiv1_proto.CommitFile
	var externalSecretsFiles []*capiv1_proto.CommitFile

	if len(ca) > 0 {
		for _, c := range ca {
			cluster := createNamespacedName(c.Cluster.Name, c.Cluster.Namespace)

			if c.Kustomization != nil {
				if c.Kustomization.Spec.CreateNamespace {
					namespace, err := generateNamespaceFile(ctx, c.IsControlPlane, cluster, c.Kustomization.Spec.TargetNamespace, c.FilePath)
					if err != nil {
						return nil, err
					}

					kustomizationFiles = append(kustomizationFiles, &capiv1_proto.CommitFile{
						Path:    *namespace.Path,
						Content: *namespace.Content,
					})
				}

				kustomization, err := generateKustomizationFile(ctx, c.IsControlPlane, cluster, client, c.Kustomization, c.FilePath)

				if err != nil {
					return nil, err
				}

				kustomizationFiles = append(kustomizationFiles, &capiv1_proto.CommitFile{
					Path:    *kustomization.Path,
					Content: *kustomization.Content,
				})
			}

			if c.HelmRelease != nil {
				helmRelease, err := generateHelmReleaseFile(ctx, c.IsControlPlane, cluster, client, c.HelmRelease, c.FilePath)

				if err != nil {
					return nil, err
				}

				helmReleaseFiles = append(helmReleaseFiles, &capiv1_proto.CommitFile{
					Path:    *helmRelease.Path,
					Content: *helmRelease.Content,
				})
			}

			if c.ExternalSecret != nil {
				externalSecret, err := generateExternalSecretFile(ctx, c.IsControlPlane, cluster, client, c.ExternalSecret, c.FilePath)

				if err != nil {
					return nil, err
				}

				externalSecretsFiles = append(externalSecretsFiles, &capiv1_proto.CommitFile{
					Path:    *externalSecret.Path,
					Content: *externalSecret.Content,
				})
			}
			clusters = append(clusters, c.Cluster.Name)
		}
	}

	return &GetAutomations{KustomizationFiles: kustomizationFiles, HelmReleaseFiles: helmReleaseFiles, Clusters: clusters, ExternalSecretsFiles: externalSecretsFiles}, nil
}

func generateHelmReleaseFile(
	ctx context.Context,
	isControlPlane bool,
	cluster types.NamespacedName,
	kubeClient client.Client,
	helmRelease *capiv1_proto.HelmRelease,
	filePath string) (gitprovider.CommitFile, error) {
	kustomizationYAML, err := createHelmReleaseObject(helmRelease)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("failed to create Helm Release object: %s/%s: %w", helmRelease.Metadata.Namespace, helmRelease.Metadata.Name, err)
	}

	b, err := yaml.Marshal(kustomizationYAML)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("error marshalling %s helmrelease, %w", helmRelease.Metadata.Name, err)
	}

	hr := createNamespacedName(helmRelease.Metadata.Name, helmRelease.Metadata.Namespace)

	helmReleasePath := getClusterResourcePath(isControlPlane, "helmrelease", cluster, hr)
	if filePath != "" {
		helmReleasePath = filePath
	}

	helmReleaseContent := string(b)

	file := &gitprovider.CommitFile{
		Path:    &helmReleasePath,
		Content: &helmReleaseContent,
	}

	return *file, nil
}

func createHelmReleaseObject(hr *capiv1_proto.HelmRelease) (*helmv2.HelmRelease, error) {
	var jsonValues []byte

	if hr.Spec.Values != "" {
		valuesData, err := ParseValues([]byte(hr.Spec.Values))
		if err != nil {
			return nil, fmt.Errorf("failed to yaml-unmarshal values: %w", err)
		}

		jsonValues, err = json.Marshal(valuesData)
		if err != nil {
			return nil, fmt.Errorf("failed to json-marshal values: %w", err)
		}
	}

	generatedHelmRelease := helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2.GroupVersion.Identifier(),
			Kind:       helmv2.HelmReleaseKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hr.Metadata.Name,
			Namespace: HelmReleaseNamespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			TargetNamespace: hr.Metadata.Namespace,
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: hr.Spec.Chart.Spec.Chart,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: sourcev1.GroupVersion.Identifier(),
						Kind:       sourcev1.HelmRepositoryKind,
						Name:       hr.Spec.Chart.Spec.SourceRef.Name,
						Namespace:  hr.Spec.Chart.Spec.SourceRef.Namespace,
					},
					Version: hr.Spec.Chart.Spec.Version,
				},
			},
			Interval: metav1.Duration{Duration: time.Minute * 10},
			Values:   &apiextensionsv1.JSON{Raw: []byte(jsonValues)},
		},
	}

	return &generatedHelmRelease, nil
}

func applyCreateAutomationDefaults(msg []*capiv1_proto.ClusterAutomation) {
	for _, c := range msg {
		if c.HelmRelease != nil && c.HelmRelease.Metadata != nil && c.HelmRelease.Metadata.Namespace == "" {
			c.HelmRelease.Metadata.Namespace = defaultAutomationNamespace
		}
		if c.Kustomization != nil && c.Kustomization.Metadata != nil && c.Kustomization.Metadata.Namespace == "" {
			c.Kustomization.Metadata.Namespace = defaultAutomationNamespace
		}
	}
}

func validateAutomations(ca []*capiv1_proto.ClusterAutomation) error {
	var err error

	if len(ca) == 0 {
		err = multierror.Append(err, fmt.Errorf(createClusterAutomationsRequiredErr))
	}

	for _, c := range ca {
		if c.Cluster == nil {
			err = multierror.Append(err, fmt.Errorf("cluster object must be specified"))
		} else {
			if c.Cluster.Name == "" {
				err = multierror.Append(err, fmt.Errorf("cluster name must be specified"))
			}

			invalidNamespaceErr := validateNamespace(c.Cluster.Namespace)
			if invalidNamespaceErr != nil {
				err = multierror.Append(err, invalidNamespaceErr)
			}
		}

		if c.Kustomization != nil {
			err = multierror.Append(err, validateKustomization(c.Kustomization))
		} else if c.HelmRelease != nil {
			err = multierror.Append(err, validateHelmRelease(c.HelmRelease))
		} else if c.ExternalSecret != nil {
			err = multierror.Append(err, validateExternalSecret(c.ExternalSecret))
		} else {
			err = multierror.Append(err, fmt.Errorf("cluster automation must contain either kustomization or helm release or external secret"))
		}
	}

	return err
}

func validateHelmRelease(helmRelease *capiv1_proto.HelmRelease) error {
	var err error

	if helmRelease.Metadata == nil {
		err = multierror.Append(err, errors.New("helmrelease metadata must be specified"))
	} else {
		if helmRelease.Metadata.Name == "" {
			err = multierror.Append(err, fmt.Errorf("helmrelease name must be specified"))
		}

		invalidNamespaceErr := validateNamespace(helmRelease.Metadata.Namespace)
		if invalidNamespaceErr != nil {
			err = multierror.Append(err, invalidNamespaceErr)
		}
	}

	if helmRelease.Spec.Chart == nil {
		err = multierror.Append(
			err,
			fmt.Errorf("chart must be specified in HelmRelease %s",
				helmRelease.Metadata.Name))
	} else {
		if helmRelease.Spec.Chart.Spec.Chart == "" {
			err = multierror.Append(
				err,
				fmt.Errorf("chart name must be specified in HelmRelease %s",
					helmRelease.Metadata.Name))
		}

		if helmRelease.Spec.Chart.Spec.SourceRef != nil {
			if helmRelease.Spec.Chart.Spec.SourceRef.Name == "" {
				err = multierror.Append(
					err,
					fmt.Errorf("sourceRef name must be specified in chart %s in HelmRelease %s",
						helmRelease.Spec.Chart.Spec.Chart, helmRelease.Metadata.Name))
			}

			invalidNamespaceErr := validateNamespace(helmRelease.Spec.Chart.Spec.SourceRef.Namespace)
			if invalidNamespaceErr != nil {
				err = multierror.Append(err, invalidNamespaceErr)
			}
		}
	}

	return err
}

func generateNamespaceFile(
	ctx context.Context,
	isControlPlane bool,
	cluster types.NamespacedName,
	name,
	filePath string) (gitprovider.CommitFile, error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	b, err := yaml.Marshal(namespace)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("error marshalling %s namespace, %w", name, err)
	}

	k := createNamespacedName(name, "")

	namespacePath := getClusterResourcePath(isControlPlane, "namespace", cluster, k)

	if filePath != "" {
		namespacePath = filePath
	}

	namespaceContent := string(b)

	file := &gitprovider.CommitFile{
		Path:    &namespacePath,
		Content: &namespaceContent,
	}

	return *file, nil
}

func generateExternalSecretFile(
	ctx context.Context,
	isControlPlane bool,
	cluster types.NamespacedName,
	kubeClient client.Client,
	externalSecret *capiv1_proto.ExternalSecret,
	filePath string) (gitprovider.CommitFile, error) {

	externalSecretYAML, err := createExternalSecretObject(externalSecret)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("failed to create External Secret object: %s/%s: %w", externalSecret.Metadata.Namespace, externalSecret.Metadata.Name, err)
	}

	b, err := yaml.Marshal(externalSecretYAML)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("error marshalling %s external secret, %w", externalSecret.Metadata.Name, err)
	}
	es := createNamespacedName(externalSecret.Metadata.Name, externalSecret.Metadata.Namespace)
	externalSecretPath := getClusterResourcePath(isControlPlane, "externalsecret", cluster, es)
	if filePath != "" {
		externalSecretPath = filePath
	}

	externalSecretContent := string(b)

	return gitprovider.CommitFile{
		Path:    &externalSecretPath,
		Content: &externalSecretContent,
	}, nil
}

func createExternalSecretObject(es *capiv1_proto.ExternalSecret) (*esv1beta1.ExternalSecret, error) {

	refreshInterval, err := time.ParseDuration(es.Spec.RefreshInterval)
	if err != nil {
		return &esv1beta1.ExternalSecret{}, err
	}
	generatedExternalSecret := &esv1beta1.ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       esv1beta1.ExtSecretKind,
			APIVersion: esv1beta1.ExtSecretGroupVersionKind.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      es.Metadata.Name,
			Namespace: es.Metadata.Namespace,
		},
		Spec: esv1beta1.ExternalSecretSpec{
			SecretStoreRef: esv1beta1.SecretStoreRef{
				Name: es.Spec.SecretStoreRef.Name,
				Kind: "SecretStore",
			},
			RefreshInterval: &metav1.Duration{
				Duration: refreshInterval,
			},
			Target: esv1beta1.ExternalSecretTarget{
				Name:           es.Spec.Target.Name,
				CreationPolicy: esv1beta1.ExternalSecretCreationPolicy("Owner"),
			},
			Data: []esv1beta1.ExternalSecretData{
				{
					SecretKey: es.Spec.Data.SecretKey,
					RemoteRef: esv1beta1.ExternalSecretDataRemoteRef{
						Key:      es.Spec.Data.RemoteRef.Key,
						Property: es.Spec.Data.RemoteRef.Property,
					},
				},
			},
		},
	}

	return generatedExternalSecret, nil
}

func validateExternalSecret(externalSecret *capiv1_proto.ExternalSecret) error {
	var err error

	if externalSecret.Metadata == nil {
		err = multierror.Append(err, errors.New("external secret metadata must be specified"))
		return err
	} else {
		if externalSecret.Metadata.Name == "" {
			err = multierror.Append(err, fmt.Errorf("external secret name must be specified"))
			return err
		}
		if externalSecret.Metadata.Namespace == "" {
			err = multierror.Append(err, fmt.Errorf("external secret namespace must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		} else {
			invalidNamespaceErr := validateNamespace(externalSecret.Metadata.Namespace)
			if invalidNamespaceErr != nil {
				err = multierror.Append(err, invalidNamespaceErr)
			}
		}
	}

	if externalSecret.Spec.SecretStoreRef == nil {
		err = multierror.Append(err, fmt.Errorf("external secret secretStoreRef must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
	} else {
		if externalSecret.Spec.SecretStoreRef.Name == "" {
			err = multierror.Append(err, fmt.Errorf("secretStoreRef name must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		}
	}

	if externalSecret.Spec.Target == nil {
		err = multierror.Append(err, fmt.Errorf("external secret target must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
	} else {
		if externalSecret.Spec.Target.Name == "" {
			err = multierror.Append(err, fmt.Errorf("target name must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		}
	}

	if externalSecret.Spec.Data == nil {
		err = multierror.Append(err, fmt.Errorf("external secret data must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
	} else {
		if externalSecret.Spec.Data.SecretKey == "" {
			err = multierror.Append(err, fmt.Errorf("secretKey must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		}
		if externalSecret.Spec.Data.RemoteRef.Key == "" {
			err = multierror.Append(err, fmt.Errorf("remoteRef key kind must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		}
		if externalSecret.Spec.Data.RemoteRef.Property == "" {
			err = multierror.Append(err, fmt.Errorf("remoteRef property kind must be specified in ExternalSecret %s", externalSecret.Metadata.Name))
		}
	}

	return err
}
