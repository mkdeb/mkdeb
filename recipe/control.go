package recipe

// Control is a recipe control instance.
type Control struct {
	Section  string `yaml:"section"`
	Priority string `yaml:"priority"`
	Version  struct {
		Epoch uint `yaml:"epoch"`
	} `yaml:"version"`
	Depends     []string `yaml:"depends"`
	PreDepends  []string `yaml:"pre-depends"`
	Recommends  []string `yaml:"recommends"`
	Suggests    []string `yaml:"suggests"`
	Enhances    []string `yaml:"enhances"`
	Breaks      []string `yaml:"breaks"`
	Conflicts   []string `yaml:"conflicts"`
	Description string   `yaml:"description"`
}
