package templates

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"sigs.k8s.io/yaml"
)

func ProfileAnnotations(tmpl templatesv1.Template) map[string]string {
	profileAnnotations := make(map[string]string)
	for k, v := range tmpl.GetAnnotations() {
		if strings.HasPrefix(k, ProfilesAnnotation) {
			profileAnnotations[k] = v
		}
	}
	return profileAnnotations
}

// profileAnnotation is a struct to unmarshal the profile annotations
// "required" and should be a pointer so that we can tell if it was set or not
// as its default value is true even if it was not set
type profileAnnotation struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
	Required  *bool  `json:"required"`
	Editable  bool   `json:"editable"`
	Values    string `json:"values"`
}

func GetProfilesFromTemplate(tl templatesv1.Template) ([]*capiv1_proto.TemplateProfile, error) {
	profilesIndex := map[string]*capiv1_proto.TemplateProfile{}
	for _, v := range ProfileAnnotations(tl) {
		profile := profileAnnotation{}
		err := json.Unmarshal([]byte(v), &profile)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal profiles: %w", err)
		}
		if profile.Name == "" {
			return nil, fmt.Errorf("profile name is required")
		}

		required := true
		if profile.Required != nil {
			required = *profile.Required
		}

		profilesIndex[profile.Name] = &capiv1_proto.TemplateProfile{
			Name:      profile.Name,
			Version:   profile.Version,
			Namespace: profile.Namespace,
			Required:  required,
			Editable:  profile.Editable,
			Values:    profile.Values,
		}
	}

	// Override anything that was still in the index with the profiles from the spec
	for _, v := range tl.GetSpec().Charts.Items {
		profile := capiv1_proto.TemplateProfile{
			Name:      v.Chart,
			Version:   v.Version,
			Namespace: v.TargetNamespace,
			Layer:     v.Layer,
			Required:  v.Required,
			Editable:  v.Editable,
		}

		if v.Values != nil {
			valuesBytes, err := yaml.Marshal(v.Values)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal profile.values for %s: %w", v.Chart, err)
			}
			profile.Values = string(valuesBytes)
		}

		if v.HelmReleaseTemplate.Content != nil {
			profileTemplateBytes, err := yaml.Marshal(v.HelmReleaseTemplate.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal spec for %s: %w", v.Chart, err)
			}
			profile.ProfileTemplate = string(profileTemplateBytes)
		}

		profilesIndex[profile.Name] = &profile
	}

	profiles := []*capiv1_proto.TemplateProfile{}
	for _, v := range profilesIndex {
		profiles = append(profiles, v)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })

	return profiles, nil
}

func TemplateHasRequiredProfiles(tl templatesv1.Template) (bool, error) {
	profiles, err := GetProfilesFromTemplate(tl)
	if err != nil {
		return false, err
	}
	for _, p := range profiles {
		if p.Required {
			return true, nil
		}
	}
	return false, nil
}
