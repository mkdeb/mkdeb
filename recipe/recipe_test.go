package recipe

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecipeValid(t *testing.T) {
	r, err := LoadRecipe("testdata/valid")
	assert.Nil(t, err)

	assert.Equal(t, 1, r.Version)
	assert.Equal(t, "foo", r.Name)
	assert.Equal(t, "a great description", r.Description)
	assert.Equal(t, "Foo Bar <foo@example.org>", r.Maintainer)
	assert.Equal(t, "https://example.org/", r.Homepage)

	// Check for "source" section
	assert.Equal(t, "https://example.org/path/to/foo-{{ .Version }}.{{ .Arch }}.tar.gz", r.Source.URL)
	assert.Equal(t, 1, r.Source.Strip)
	assert.Equal(t, map[string]string{"amd64": "amd64"}, r.Source.ArchMapping)

	// Check for "control" section
	assert.Equal(t, "admin", r.Control.Section)
	assert.Equal(t, "optional", r.Control.Priority)
	assert.Equal(t, []string{"bar"}, r.Control.Depends)
	assert.Equal(t, []string{"baz"}, r.Control.PreDepends)
	assert.Equal(t, []string{"barbar"}, r.Control.Recommends)
	assert.Equal(t, []string{"barbaz"}, r.Control.Suggests)
	assert.Equal(t, []string{"foobar"}, r.Control.Enhances)
	assert.Equal(t, []string{"foobaz"}, r.Control.Breaks)
	assert.Equal(t, []string{"foobarbaz"}, r.Control.Conflicts)
	assert.Equal(t, "A long package description providing us with information on the upstream software.",
		r.Control.Description)

	// Check for "install" section
	assert.Equal(t, InstallMap{
		"/etc/init.d": []InstallRule{{Pattern: "init", Rename: "foo", ConfFile: true}},
	}, r.Install.Recipe)
	assert.Equal(t, InstallMap{
		"/usr/bin": []InstallRule{{Pattern: "foo", Rename: "", ConfFile: false}},
	}, r.Install.Upstream)

	// Check for "dirs" section
	assert.Equal(t, []string{"/path/to/dir"}, r.Dirs)

	// Check for "links" section
	assert.Equal(t, map[string]string{"/path/to/link": "/path/to/target"}, r.Links)

	// Check for control and recipe files
	controlFiles := []File{}
	fi, err := os.Stat("testdata/valid/control/postinst")
	assert.Nil(t, err)
	controlFiles = append(controlFiles, File{"testdata/valid/control/postinst", fi})
	fi, err = os.Stat("testdata/valid/control/postrm")
	assert.Nil(t, err)
	controlFiles = append(controlFiles, File{"testdata/valid/control/postrm", fi})
	assert.Equal(t, controlFiles, r.ControlFiles)

	recipeFiles := []File{}
	fi, err = os.Stat("testdata/valid/files/init")
	assert.Nil(t, err)
	recipeFiles = append(recipeFiles, File{"testdata/valid/files/init", fi})
	assert.Equal(t, recipeFiles, r.RecipeFiles)

	// Check for path matching
	path, confFile, ok := r.InstallPath("init", r.Install.Recipe)
	assert.Equal(t, "/etc/init.d/foo", path)
	assert.True(t, confFile, ok)
	path, confFile, ok = r.InstallPath("foo", r.Install.Upstream)
	assert.Equal(t, "/usr/bin/foo", path)
	assert.False(t, confFile)
	assert.True(t, ok)
	path, confFile, ok = r.InstallPath("bar", r.Install.Upstream)
	assert.Equal(t, "", path)
	assert.False(t, confFile, ok)
}

func TestRecipeUnsupportedVersion(t *testing.T) {
	r, err := LoadRecipe("testdata/unsupported-version")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrUnsupportedVersion, r.Validate())
}

func TestRecipeMissingName(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-name")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingName, r.Validate())
}

func TestRecipeMissingDescription(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-description")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingDescription, r.Validate())
}

func TestRecipeMissingMaintainer(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-maintainer")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingMaintainer, r.Validate())
}

func TestRecipeMissingSourceURL(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-source-url")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingSourceURL, r.Validate())
}

func TestRecipeMissingControl(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-control")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingControl, r.Validate())
}

func TestRecipeMissingControlDescription(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-control-description")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingControlDescription, r.Validate())
}

func TestRecipeMissingInstall(t *testing.T) {
	r, err := LoadRecipe("testdata/missing-install")
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, ErrMissingInstall, r.Validate())
}
