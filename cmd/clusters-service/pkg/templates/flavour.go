package templates

// Flavour represents a specific template.
type Flavour struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Params      []Param `json:"params"`
	Version     string  `json:"version"`
}

// Param is a parameter that can be templated into a struct.
type Param struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Options     []string `json:"options"`
}
