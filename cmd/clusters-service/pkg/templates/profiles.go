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

// profileAnnotation is a struct to unmarshal the profile annotations.
// "required" should be a pointer so that we can tell if it was set or not
// as its default value is true even if it was not set
type profileAnnotation struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
	Required  *bool  `json:"required"`
	Editable  bool   `json:"editable"`
	Values    string `json:"values"`
}

// GetProfilesFromAnnotations returns a list of profiles defined in the template.
// Both the annotations and the spec are used to determine the profiles.
// spec.Charts takes precedence over annotations if both are defined for the same profile.
func GetProfilesFromTemplate(tl templatesv1.Template) ([]*capiv1_proto.TemplateProfile, error) {
	profilesIndex, err := getProfilesFromAnnotations(tl)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles from annotations: %w", err)
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
			SourceRef: &capiv1_proto.SourceRef{
				Name:      v.SourceRef.Name,
				Namespace: v.SourceRef.Namespace,
			},
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

func getProfilesFromAnnotations(tl templatesv1.Template) (map[string]*capiv1_proto.TemplateProfile, error) {
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

	return profilesIndex, nil
}

// TemplateHasRequiredProfiles returns true if the template has any required profiles.
// Note: Its an implicit system requirement that annotations are valid JSON before being
// rendered, so we can determine button status in the UI etc. This fn will raise an error on
// invalid JSON and thats ok, users need to fix their templates.
func TemplateHasRequiredProfiles(tl templatesv1.Template) (bool, error) {
	profiles, err := GetProfilesFromTemplate(tl)
	if err != nil {
		return false, fmt.Errorf("failed to get profiles from template: %w", err)
	}
	for _, p := range profiles {
		if p.Required {
			return true, nil
		}
	}
	return false, nil
}
