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
	template "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

type GetFiles struct {
    RenderedTemplate []gitprovider.CommitFile
	ProfileFiles []gitprovider.CommitFile
    KustomizationFiles   []gitprovider.CommitFile
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1_proto.ListTemplatesRequest) (*capiv1_proto.ListTemplatesResponse, error) {
	templates := []*capiv1_proto.Template{}
	includeGitopsTemplates := msg.TemplateKind == "" || msg.TemplateKind == gapiv1.Kind
	includeCAPITemplates := msg.TemplateKind == "" || msg.TemplateKind == capiv1.Kind

	if includeGitopsTemplates {
		tl, err := s.templatesLibrary.List(ctx, gapiv1.Kind)
		if err != nil {
			return nil, fmt.Errorf("error listing gitops templates: %w", err)
		}
		for _, t := range tl {
			templates = append(templates, ToTemplateResponse(t))
		}
	}

	if includeCAPITemplates {
		tl, err := s.templatesLibrary.List(ctx, capiv1.Kind)
		if err != nil {
			return nil, fmt.Errorf("error listing capi templates: %w", err)
		}
		for _, t := range tl {
			templates = append(templates, ToTemplateResponse(t))
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
	}, nil
}

func (s *server) GetTemplate(ctx context.Context, msg *capiv1_proto.GetTemplateRequest) (*capiv1_proto.GetTemplateResponse, error) {
	// Default to CAPI kind to ease transition
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}
	tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName, msg.TemplateKind)
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
	tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName, msg.TemplateKind)
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
	tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName, msg.TemplateKind)
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

	tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}

	git_files, err := s.getFiles(ctx, tm, msg.ClusterNamespace, msg.TemplateName, msg.TemplateKind, msg.Values, msg.Credentials, msg.Profiles, msg.Kustomizations)
	
	if err != nil {
		return nil, err
	}

	var profileFiles []*capiv1_proto.CommitFile
	var kustomizationFiles []*capiv1_proto.CommitFile
	var renderedTemplate []*capiv1_proto.CommitFile

	if len(git_files.ProfileFiles) > 0 {
		for _, f := range git_files.ProfileFiles {
			profileFiles = append(profileFiles, toCommitFile(f))
		}
	}

	if len(git_files.KustomizationFiles) > 0 {
		for _, f := range git_files.KustomizationFiles {
			kustomizationFiles = append(kustomizationFiles, toCommitFile(f))
		}
	}

	if len(git_files.KustomizationFiles) > 0 {
		for _, f := range git_files.KustomizationFiles {
			kustomizationFiles = append(kustomizationFiles, toCommitFile(f))
		}
	}

	if len(git_files.RenderedTemplate) > 0 {
		for _, f := range git_files.RenderedTemplate {
			renderedTemplate = append(renderedTemplate, toCommitFile(f))
		}
	}

	return &capiv1_proto.RenderTemplateResponse{RenderedTemplate: renderedTemplate, ProfileFiles: profileFiles, KustomizationFiles: kustomizationFiles}, err
}


func (s *server) getFiles(ctx context.Context,tm template.Template, cluster_namespace string, template_name string, template_kind string, parameter_values map[string]string, template_credentials *capiv1_proto.Credential, profiles []*capiv1_proto.ProfileValues, kustomizations []*capiv1_proto.Kustomization) (*GetFiles, error) {	
	if template_kind == "" {
		template_kind = capiv1.Kind
	}	

	s.log.WithValues("request_values", parameter_values, "request_credentials", template_credentials).Info("Received message")

	tmplWithValues, err := renderTemplateWithValues(tm, template_name, getClusterNamespace(cluster_namespace), parameter_values)
	if err != nil {
		return nil, err
	}

	if err = templates.ValidateRenderedTemplates(tmplWithValues); err != nil {
		return nil, fmt.Errorf("validation error rendering template %v, %v", template_name, err)
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.log, client, tmplWithValues, template_credentials, template_name)
	if err != nil {
		return nil, err
	}


	clusterNamespace := getClusterNamespace(parameter_values["NAMESPACE"])
	clusterName, ok := parameter_values["CLUSTER_NAME"]
	if !ok {
		return nil, errors.New("unable to find 'CLUSTER_NAME' parameter in supplied values")
	}

	cluster := createNamespacedName(clusterName, clusterNamespace)

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


	var profileFiles []gitprovider.CommitFile
	var kustomizationFiles []gitprovider.CommitFile

	if len(profiles) > 0 {
		profilesFile, err := generateProfileFiles(
			ctx,
			tm,
			cluster,
			client,
			generateProfileFilesParams{
				helmRepository:         createNamespacedName(s.profileHelmRepositoryName, viper.GetString("runtime-namespace")),
				helmRepositoryCacheDir: s.helmRepositoryCacheDir,
				profileValues:          profiles,
				parameterValues:        parameter_values,
			},
		)
		if err != nil {
			return nil, err
		}
		profileFiles = append(profileFiles, *profilesFile)
	}

	if len(kustomizations) > 0 {
		for _, k := range kustomizations {
			kustomization, err := generateKustomizationFile(ctx, false, cluster, client, k, "")
			if err != nil {
				return nil, err
			}

			kustomizationFiles = append(kustomizationFiles, kustomization)
		}
	}

	return &GetFiles{RenderedTemplate: files, ProfileFiles: profileFiles, KustomizationFiles: kustomizationFiles}, err
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
