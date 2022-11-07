package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/spf13/viper"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	apiTemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetFilesRequest struct {
	ClusterNamespace string
	TemplateName     string
	TemplateKind     string
	ParameterValues  map[string]string
	Credentials      *capiv1_proto.Credential
	Profiles         []*capiv1_proto.ProfileValues
	Kustomizations   []*capiv1_proto.Kustomization
}

type GetFilesReturn struct {
	RenderedTemplate   string
	ProfileFiles       []gitprovider.CommitFile
	KustomizationFiles []gitprovider.CommitFile
	Cluster            types.NamespacedName
	CostEstimate       *capiv1_proto.CostEstimate
}

func (s *server) getTemplate(ctx context.Context, name, namespace, templateKind string) (apiTemplates.Template, error) {
	if namespace == "" {
		return nil, errors.New("need to specify template namespace")
	}
	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}
	switch templateKind {
	case capiv1.Kind:
		var t capiv1.CAPITemplate
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, &t)
		if err != nil {
			return nil, fmt.Errorf("error getting capitemplate %s/%s: %w", namespace, name, err)
		}
		// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
		t.SetGroupVersionKind(capiv1.GroupVersion.WithKind(capiv1.Kind))
		return &t, nil

	case gapiv1.Kind:
		var t gapiv1.GitOpsTemplate
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, &t)
		if err != nil {
			return nil, fmt.Errorf("error getting gitops template %s/%s: %w", namespace, name, err)
		}
		// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
		t.SetGroupVersionKind(gapiv1.GroupVersion.WithKind(gapiv1.Kind))
		return &t, nil
	}

	return nil, nil
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1_proto.ListTemplatesRequest) (*capiv1_proto.ListTemplatesResponse, error) {
	templates := []*capiv1_proto.Template{}
	errors := []*capiv1_proto.ListError{}
	includeGitopsTemplates := msg.TemplateKind == "" || msg.TemplateKind == gapiv1.Kind
	includeCAPITemplates := msg.TemplateKind == "" || msg.TemplateKind == capiv1.Kind

	if includeGitopsTemplates {
		namespacedLists, err := s.managementFetcher.Fetch(ctx, gapiv1.Kind, func() client.ObjectList {
			return &gapiv1.GitOpsTemplateList{}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query gitops templates: %w", err)
		}

		for _, namespacedList := range namespacedLists {
			if namespacedList.Error != nil {
				errors = append(errors, &capiv1_proto.ListError{
					Namespace: namespacedList.Namespace,
					Message:   namespacedList.Error.Error(),
				})
			}
			templatesList := namespacedList.List.(*gapiv1.GitOpsTemplateList)
			for _, t := range templatesList.Items {
				templates = append(templates, ToTemplateResponse(&t))
			}
		}
	}

	if includeCAPITemplates {
		namespacedLists, err := s.managementFetcher.Fetch(ctx, capiv1.Kind, func() client.ObjectList {
			return &capiv1.CAPITemplateList{}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query capi templates: %w", err)
		}

		for _, namespacedList := range namespacedLists {
			if namespacedList.Error != nil {
				errors = append(errors, &capiv1_proto.ListError{
					Namespace: namespacedList.Namespace,
					Message:   namespacedList.Error.Error(),
				})
			}
			templatesList := namespacedList.List.(*capiv1.CAPITemplateList)
			for _, t := range templatesList.Items {
				templates = append(templates, ToTemplateResponse(&t))
			}
		}
	}

	total := int32(len(templates))
	if msg.Provider != "" {
		if !isProviderRecognised(msg.Provider) {
			return nil, fmt.Errorf("provider %q is not recognised", msg.Provider)
		}

		templates = filterTemplatesByProvider(templates, msg.Provider)
	}

	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	return &capiv1_proto.ListTemplatesResponse{
		Templates: templates,
		Total:     total,
		Errors:    errors,
	}, nil
}

func (s *server) GetTemplate(ctx context.Context, msg *capiv1_proto.GetTemplateRequest) (*capiv1_proto.GetTemplateResponse, error) {
	// Default to CAPI kind to ease transition
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}
	tm, err := s.getTemplate(ctx, msg.TemplateName, msg.TemplateNamespace, msg.TemplateKind)
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
	// Default to CAPI kind to ease transition
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}
	tm, err := s.getTemplate(ctx, msg.TemplateName, msg.TemplateNamespace, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	t := ToTemplateResponse(tm)
	if t.Error != "" {
		return nil, fmt.Errorf("error looking up template params for %v, %v", msg.TemplateName, t.Error)
	}

	return &capiv1_proto.ListTemplateParamsResponse{Parameters: t.Parameters, Objects: t.Objects}, err
}

func (s *server) ListTemplateProfiles(ctx context.Context, msg *capiv1_proto.ListTemplateProfilesRequest) (*capiv1_proto.ListTemplateProfilesResponse, error) {
	// Default to CAPI kind to ease transition
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}
	tm, err := s.getTemplate(ctx, msg.TemplateName, msg.TemplateNamespace, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	t := ToTemplateResponse(tm)
	if t.Error != "" {
		return nil, fmt.Errorf("error looking up template annotations for %v, %v", msg.TemplateName, t.Error)
	}

	profiles, err := getProfilesFromTemplate(t.Annotations)
	if err != nil {
		return nil, fmt.Errorf("error getting profiles from template %v, %v", msg.TemplateName, err)
	}

	return &capiv1_proto.ListTemplateProfilesResponse{Profiles: profiles, Objects: t.Objects}, err
}

func toCommitFile(file gitprovider.CommitFile) *capiv1_proto.CommitFile {
	return &capiv1_proto.CommitFile{
		Path:    *file.Path,
		Content: *file.Content,
	}
}

// Similar the others list and get will right now only work with CAPI templates.
// tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName) -> this get is the key.
func (s *server) RenderTemplate(ctx context.Context, msg *capiv1_proto.RenderTemplateRequest) (*capiv1_proto.RenderTemplateResponse, error) {
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}

	s.log.WithValues("request_values", msg.Values, "request_credentials", msg.Credentials).Info("Received message")
	tm, err := s.getTemplate(ctx, msg.TemplateName, msg.TemplateNamespace, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}

	files, err := s.getFiles(
		ctx,
		tm,
		GetFilesRequest{msg.ClusterNamespace, msg.TemplateName, msg.TemplateKind, msg.Values, msg.Credentials, msg.Profiles, msg.Kustomizations},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var profileFiles []*capiv1_proto.CommitFile
	var kustomizationFiles []*capiv1_proto.CommitFile

	if len(files.ProfileFiles) > 0 {
		for _, f := range files.ProfileFiles {
			profileFiles = append(profileFiles, toCommitFile(f))
		}
	}

	if len(files.KustomizationFiles) > 0 {
		for _, f := range files.KustomizationFiles {
			kustomizationFiles = append(kustomizationFiles, toCommitFile(f))
		}
	}

	return &capiv1_proto.RenderTemplateResponse{RenderedTemplate: files.RenderedTemplate, ProfileFiles: profileFiles, KustomizationFiles: kustomizationFiles, CostEstimate: files.CostEstimate}, err
}

func (s *server) getFiles(ctx context.Context, tmpl apiTemplates.Template, msg GetFilesRequest, createRequestMessage *capiv1_proto.CreatePullRequestRequest) (*GetFilesReturn, error) {
	clusterName := msg.ParameterValues["CLUSTER_NAME"]
	clusterNamespace := getClusterNamespace(msg.ParameterValues["NAMESPACE"])
	resourceName := msg.ParameterValues["RESOURCE_NAME"]

	tmplWithValues, err := renderTemplateWithValues(tmpl, msg.TemplateName, getClusterNamespace(msg.ClusterNamespace), msg.ParameterValues)
	if err != nil {
		return nil, err
	}

	if createRequestMessage != nil {
		tmplWithValues, err = templates.InjectJSONAnnotation(tmplWithValues, "templates.weave.works/create-request", createRequestMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to annotate template with parameter values: %w", err)
		}
	}

	if err = templates.ValidateRenderedTemplates(tmplWithValues); err != nil {
		return nil, fmt.Errorf("validation error rendering template %v, %v", msg.TemplateName, err)
	}

	// if this feature is not enabled the Nil estimator will be invoked returning a nil estimate
	costEstimate := getCostEstimate(ctx, s.estimator, tmplWithValues)

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.log, client, tmplWithValues, msg.Credentials, msg.TemplateName)
	if err != nil {
		return nil, err
	}
	content := string(tmplWithValuesAndCredentials)

	if clusterName == "" && resourceName == "" {
		return nil, errors.New("unable to find 'CLUSTER_NAME' or 'RESOURCE_NAME' parameter in supplied values")
	}

	if clusterName != "" {
		resourceName = clusterName
	}

	cluster := createNamespacedName(resourceName, clusterNamespace)

	var profileFiles []gitprovider.CommitFile
	var kustomizationFiles []gitprovider.CommitFile

	if len(msg.Profiles) > 0 {
		profilesFile, err := generateProfileFiles(
			ctx,
			tmpl,
			cluster,
			client,
			generateProfileFilesParams{
				helmRepository:         createNamespacedName(s.profileHelmRepositoryName, viper.GetString("runtime-namespace")),
				helmRepositoryCacheDir: s.helmRepositoryCacheDir,
				profileValues:          msg.Profiles,
				parameterValues:        msg.ParameterValues,
			},
		)
		if err != nil {
			return nil, err
		}
		profileFiles = append(profileFiles, *profilesFile)
	}

	if len(msg.Kustomizations) > 0 {
		for _, k := range msg.Kustomizations {
			// FIXME: dedup this with the automations
			if k.Spec.CreateNamespace {
				namespace, err := generateNamespaceFile(ctx, false, cluster, k.Spec.TargetNamespace, "")
				if err != nil {
					return nil, err
				}
				kustomizationFiles = append(kustomizationFiles, gitprovider.CommitFile{
					Path:    namespace.Path,
					Content: namespace.Content,
				})
			}

			kustomization, err := generateKustomizationFile(ctx, false, cluster, client, k, "")
			if err != nil {
				return nil, err
			}

			kustomizationFiles = append(kustomizationFiles, kustomization)
		}
	}

	if shouldAddCommonBases(tmpl) {
		commonKustomization, err := getCommonKustomization(cluster)
		if err != nil {
			return nil, fmt.Errorf("failed to get common kustomization for %s: %s", msg.ParameterValues, err)
		}
		kustomizationFiles = append(kustomizationFiles, *commonKustomization)
	}

	return &GetFilesReturn{RenderedTemplate: content, ProfileFiles: profileFiles, KustomizationFiles: kustomizationFiles, Cluster: cluster, CostEstimate: costEstimate}, err
}

func shouldAddCommonBases(t apiTemplates.Template) bool {
	anno := t.GetAnnotations()[templates.AddCommonBasesAnnotation]
	if anno != "" {
		return anno == "true"
	}

	return viper.GetString("add-bases-kustomization") != "disabled" && isCAPITemplate(t)
}

func getCostEstimate(ctx context.Context, estimator estimation.Estimator, tmplWithValues [][]byte) *capiv1_proto.CostEstimate {
	unstructureds, err := templates.ConvertToUnstructured(tmplWithValues)
	if err != nil {
		return &capiv1_proto.CostEstimate{Message: fmt.Sprintf("failed to parse rendered templates: %s", err)}
	}

	estimate, err := estimator.Estimate(ctx, unstructureds)
	if err != nil {
		return &capiv1_proto.CostEstimate{Message: fmt.Sprintf("failed to calculate estimate for cluster costs: %s", err)}
	}
	if estimate == nil {
		return &capiv1_proto.CostEstimate{Message: "no estimate returned"}
	}

	return &capiv1_proto.CostEstimate{
		Currency: estimate.Currency,
		Range: &capiv1_proto.CostEstimate_Range{
			Low:  estimate.Low,
			High: estimate.High,
		},
	}
}

func isProviderRecognised(provider string) bool {
	for _, p := range providers {
		if strings.EqualFold(provider, p) {
			return true
		}
	}
	return false
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

func getProfilesFromTemplate(annotations map[string]string) ([]*capiv1_proto.TemplateProfile, error) {
	profiles := []*capiv1_proto.TemplateProfile{}
	for k, v := range annotations {
		if strings.Contains(k, "capi.weave.works/profile-") {
			profile := capiv1_proto.TemplateProfile{}
			err := json.Unmarshal([]byte(v), &profile)
			if err != nil {
				return profiles, fmt.Errorf("failed to unmarshal profiles: %w", err)
			}
			profiles = append(profiles, &profile)
		}
	}

	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })

	return profiles, nil
}
