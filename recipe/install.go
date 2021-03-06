package recipe

// Install is a recipe installation.
type Install struct {
	Recipe   InstallMap `yaml:"recipe"`
	Upstream InstallMap `yaml:"upstream"`
}

// InstallMap is a recipe installation map.
type InstallMap map[string][]InstallRule

// InstallRule is a recipe installation rule.
type InstallRule struct {
	Pattern  string `yaml:"pattern"`
	Exclude  string `yaml:"exclude"`
	Rename   string `yaml:"rename"`
	ConfFile bool   `yaml:"conffile"`
}
