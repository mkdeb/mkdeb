package recipe

// Source represents a recipe source instance.
type Source struct {
	URL         string            `yaml:"url"`
	Strip       int               `yaml:"strip"`
	ArchMapping map[string]string `yaml:"arch-mapping"`
}
