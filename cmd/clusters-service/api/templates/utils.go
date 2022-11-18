package templates

// ConvertV1SpecToSpec converts between the two specs.
func ConvertV1SpecToSpec(s TemplateSpecV1) TemplateSpec {
	return TemplateSpec{
		Description:       s.Description,
		RenderType:        s.RenderType,
		Params:            s.Params,
		ResourceTemplates: s.ResourceTemplates,
		TestField:         "coming-from-v1",
	}
}
