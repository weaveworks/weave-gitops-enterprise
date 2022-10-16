package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/spf13/viper"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

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
	// Default to CAPI kind to ease transition
	if msg.TemplateKind == "" {
		msg.TemplateKind = capiv1.Kind
	}
	s.log.WithValues("request_values", msg.Values, "request_credentials", msg.Credentials).Info("Received message")
	tm, err := s.templatesLibrary.Get(ctx, msg.TemplateName, msg.TemplateKind)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	renderedTemplates, err := renderTemplateWithValues(tm, msg.TemplateName, getClusterNamespace(msg.ClusterNamespace), msg.Values)
	if err != nil {
		return nil, err
	}

	var templateBits [][]byte
	for _, renderedTemplate := range renderedTemplates {
		templateBits = append(templateBits, renderedTemplate.Data...)
	}

	if err = templates.ValidateRenderedTemplates(templateBits); err != nil {
		return nil, fmt.Errorf("validation error rendering template %v, %v", msg.TemplateName, err)
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.log, client, templateBits, msg.Credentials, msg.TemplateName)
	if err != nil {
		return nil, err
	}

	resultStr := string(tmplWithValuesAndCredentials[:])

	var profileFiles []*capiv1_proto.CommitFile
	var kustomizationFiles []*capiv1_proto.CommitFile

	cluster := createNamespacedName(msg.Values["CLUSTER_NAME"], msg.Values["NAMESPACE"])

	if len(msg.Profiles) > 0 {
		profilesFile, err := generateProfileFiles(
			ctx,
			tm,
			cluster,
			client,
			generateProfileFilesParams{
				helmRepository:         createNamespacedName(s.profileHelmRepositoryName, viper.GetString("runtime-namespace")),
				helmRepositoryCacheDir: s.helmRepositoryCacheDir,
				profileValues:          msg.Profiles,
				parameterValues:        msg.Values,
			},
		)
		if err != nil {
			return nil, err
		}
		profileFiles = append(profileFiles, toCommitFile(*profilesFile))
	}

	if len(msg.Kustomizations) > 0 {
		for _, k := range msg.Kustomizations {
			kustomization, err := generateKustomizationFile(ctx, false, cluster, client, k, "")
			if err != nil {
				return nil, err
			}

			kustomizationFiles = append(kustomizationFiles, toCommitFile(kustomization))
		}
	}
	// Temporary mock data of cost estimate
	costEstimate := &capiv1_proto.CostEstimate{
		Currency: "USD",
		Range: &capiv1_proto.CostEstimate_Range{
			Low:  0,
			High: 1000000,
		},
	}

	return &capiv1_proto.RenderTemplateResponse{RenderedTemplate: resultStr, ProfileFiles: profileFiles, KustomizationFiles: kustomizationFiles, CostEstimate: costEstimate}, err
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
