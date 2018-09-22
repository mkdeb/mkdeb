package recipe

// Install represents a recipe installation instance,
type Install struct {
	Recipe   InstallMap `yaml:"recipe"`
	Upstream InstallMap `yaml:"upstream"`
}

// InstallMap represents a recipe installation map instance.
type InstallMap map[string][]InstallRule

// InstallRule represents a recipe installation rule instance.
type InstallRule struct {
	Pattern  string `yaml:"pattern"`
	Exclude  string `yaml:"exclude"`
	Rename   string `yaml:"rename"`
	ConfFile bool   `yaml:"conffile"`
}
