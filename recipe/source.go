package recipe

// Source is a recipe source.
type Source struct {
	URL         string            `yaml:"url"`
	Type        string            `yaml:"type"`
	Strip       int               `yaml:"strip"`
	ArchMapping map[string]string `yaml:"arch-mapping"`
}
