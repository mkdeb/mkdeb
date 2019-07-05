package recipe

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
	yaml "gopkg.in/yaml.v2"
)

const defaultSourceType = "archive"

// Recipe is a packaging recipe.
type Recipe struct {
	Version     int               `yaml:"version"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Maintainer  string            `yaml:"maintainer"`
	Homepage    string            `yaml:"homepage"`
	Source      *Source           `yaml:"source"`
	Control     *Control          `yaml:"control"`
	Install     *Install          `yaml:"install"`
	Dirs        []string          `yaml:"dirs"`
	Links       map[string]string `yaml:"links"`

	ControlFiles []File
	RecipeFiles  []File
}

// LoadRecipe loads a packaging recipe given a file path.
func LoadRecipe(path string) (*Recipe, error) {
	var r *Recipe

	data, err := ioutil.ReadFile(filepath.Join(path, "recipe.yaml"))
	if err != nil {
		return nil, err
	} else if err = yaml.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	// Set defaults
	if r.Source == nil {
		r.Source = &Source{}
	}

	if r.Source.Type == "" {
		r.Source.Type = defaultSourceType
	}

	if len(r.Source.ArchMapping) == 0 {
		r.Source.ArchMapping = map[string]string{"all": ""}
	}

	// Load control and recipe files references from filesystem
	files, err := ioutil.ReadDir(filepath.Join(path, "control"))
	if err != nil && !os.IsNotExist(err) {
		return nil, xerrors.Errorf("cannot read control directory: %w", err)
	}

	for _, fi := range files {
		r.ControlFiles = append(r.ControlFiles, File{
			Path:     filepath.Join(path, "control", fi.Name()),
			FileInfo: fi,
		})
	}

	files, err = ioutil.ReadDir(filepath.Join(path, "files"))
	if err != nil && !os.IsNotExist(err) {
		return nil, xerrors.Errorf("cannot read files directory: %w", err)
	}

	for _, fi := range files {
		r.RecipeFiles = append(r.RecipeFiles, File{
			Path:     filepath.Join(path, "files", fi.Name()),
			FileInfo: fi,
		})
	}

	return r, nil
}

// InstallPath returns the destination installation path, whether it matches a configuration file path.
//
// Last returned boolean will be false if the input path doesn't match the installation rules and true otherwise.
func (r *Recipe) InstallPath(path string, m InstallMap) (string, bool, bool) {
	for base, rules := range m {
		for _, rule := range rules {
			if pathMatch(rule.Pattern, rule.Exclude, path) {
				if rule.Rename != "" {
					path = rule.Rename
				}
				return filepath.Join(base, path), rule.ConfFile, true
			}
		}
	}

	return "", false, false
}

// Validate checks for recipe validity.
func (r *Recipe) Validate() error {
	switch {
	case !VersionSupported(r.Version):
		return ErrUnsupportedVersion

	case r.Name == "":
		return ErrMissingName

	case r.Description == "":
		return ErrMissingDescription

	case r.Maintainer == "":
		return ErrMissingMaintainer

	case r.Source == nil:
		return ErrMissingSource

	case r.Source.URL == "":
		return ErrMissingSourceURL

	case r.Control == nil:
		return ErrMissingControl

	case r.Control.Description == "":
		return ErrMissingControlDescription

	case r.Install == nil:
		return ErrMissingInstall
	}

	return nil
}

func pathMatch(pattern, exclude, value string) bool {
	// Remove slashes from pattern and value as "path.Match" doesn't handle them
	pattern = strings.Replace(pattern, "/", "\x1e", -1)
	exclude = strings.Replace(exclude, "/", "\x1e", -1)
	value = strings.Replace(value, "/", "\x1e", -1)

	ok, _ := path.Match(pattern, value)
	if ok {
		nok, _ := path.Match(exclude, value)
		ok = !nok
	}
	return ok
}
