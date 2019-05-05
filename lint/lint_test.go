package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"mkdeb.sh/recipe"
)

func TestInfo(t *testing.T) {
	for _, test := range []struct {
		input    string
		expected *RuleInfo
	}{
		{
			input: "name-empty",
			expected: &RuleInfo{
				Tag:         "name-empty",
				Level:       LevelError,
				Description: "\nRecipe name must not be empty.\n",
			},
		},
		{
			input:    "unsupported-rule",
			expected: nil,
		},
	} {
		assert.Equal(t, test.expected, Info(test.input))
	}
}

func TestLint(t *testing.T) {
	for _, test := range []struct {
		recipe   string
		problems []*Problem
		ok       bool
	}{
		{
			recipe: "testdata/valid",
			ok:     true,
		},
		{
			recipe:   "testdata/invalid-name",
			problems: []*Problem{{LevelError, "name-invalid", []interface{}{"Foo"}}},
		},
	} {
		rcp, err := recipe.LoadRecipe(test.recipe)
		assert.Nil(t, err)

		problems, ok := Lint(rcp)
		assert.Equal(t, test.problems, problems)
		assert.Equal(t, test.ok, ok)
	}
}

func TestLintVersion(t *testing.T) {
	for _, test := range []struct {
		input    int
		problems []*Problem
	}{
		{
			input: 1,
		},
		{
			input:    0,
			problems: []*Problem{{LevelError, "version-unsupported", []interface{}{0}}},
		},
	} {
		l := linter{}
		l.lintVersion(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintName(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "valid",
		},
		{
			input: "also-valid",
		},
		{
			input: "a-rather-quite-long-recipe-name-that-should-trigger-a-linting-warning",
			problems: []*Problem{{LevelWarning, "name-too-long",
				[]interface{}{"a-rather-quite-long-recipe-name-that-should-trigger-a-linting-warning"}}},
		},
		{
			input:    "",
			problems: []*Problem{{LevelError, "name-empty", nil}},
		},
		{
			input:    "Invalid",
			problems: []*Problem{{LevelError, "name-invalid", []interface{}{"Invalid"}}},
		},
		{
			input:    "!invalid",
			problems: []*Problem{{LevelError, "name-invalid", []interface{}{"!invalid"}}},
		},
		{
			input:    "-invalid",
			problems: []*Problem{{LevelError, "name-invalid", []interface{}{"-invalid"}}},
		},
		{
			input:    "invalid-",
			problems: []*Problem{{LevelError, "name-invalid", []interface{}{"invalid-"}}},
		},
	} {
		l := linter{}
		l.lintName(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintDescription(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "A valid description",
		},
		{
			input:    "",
			problems: []*Problem{{LevelError, "description-empty", nil}},
		},
		{
			input: "a description without uppercase first character",
			problems: []*Problem{{LevelWarning, "description-missing-uppercase",
				[]interface{}{"a description without uppercase first character"}}},
		},
		{
			input: "A description with an extra dot.",
			problems: []*Problem{{LevelWarning, "description-extra-dot",
				[]interface{}{"A description with an extra dot."}}},
		},
		{
			input: "A rather quite long recipe description that should trigger a linting warning, " +
				"but it stills require some extra text for that",
			problems: []*Problem{{LevelWarning, "description-too-long",
				[]interface{}{"A rather quite long recipe description that should trigger a linting warning, " +
					"but it stills require some extra text for that"}}},
		},
	} {
		l := linter{}
		l.lintDescription(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintMaintainer(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "Foo Bar <foo@example.org>",
		},
		{
			input: "foo@example.org",
		},
		{
			input:    "",
			problems: []*Problem{{LevelError, "maintainer-empty", nil}},
		},
		{
			input:    "Foo Bar",
			problems: []*Problem{{LevelError, "maintainer-invalid", []interface{}{"Foo Bar"}}},
		},
	} {
		l := linter{}
		l.lintMaintainer(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintHomepage(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "https://example.net/",
		},
		{
			input:    "",
			problems: []*Problem{{LevelWarning, "homepage-empty", nil}},
		},
		{
			input:    "example.net",
			problems: []*Problem{{LevelError, "homepage-invalid", []interface{}{"example.net"}}},
		},
		{
			input:    "invalid",
			problems: []*Problem{{LevelError, "homepage-invalid", []interface{}{"invalid"}}},
		},
	} {
		l := linter{}
		l.lintHomepage(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintSource(t *testing.T) {
	l := linter{}
	l.lintSource(nil)
	assert.Equal(t, []*Problem{{LevelError, "source-empty", nil}}, l.problems)
}

func TestLintSourceURL(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "https://example.net/path/to/archive-{{ .Version }}_{{ .Arch }}.tar.gz",
		},
		{
			input:    "",
			problems: []*Problem{{LevelError, "source-url-empty", nil}},
		},
		{
			input: "https://example.net/path/to/archive-{{ .Version }}_{{ .Arch",
			problems: []*Problem{{LevelError, "source-url-invalid",
				[]interface{}{"https://example.net/path/to/archive-{{ .Version }}_{{ .Arch"}}},
		},
		{
			input: "example.net/path/to/archive-{{ .Version }}_{{ .Arch }}.tar.gz",
			problems: []*Problem{{LevelError, "source-url-invalid",
				[]interface{}{"example.net/path/to/archive-{{ .Version }}_{{ .Arch }}.tar.gz"}}},
		},
	} {
		l := linter{}
		l.lintSourceURL(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintSourceType(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "",
		},
		{
			input: "archive",
		},
		{
			input: "file",
		},
		{
			input:    "invalid",
			problems: []*Problem{{LevelError, "source-type-invalid", []interface{}{"invalid"}}},
		},
	} {
		l := linter{}
		l.lintSourceType(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintSourceStrip(t *testing.T) {
	for _, test := range []struct {
		input    int
		problems []*Problem
	}{
		{
			input: 0,
		},
		{
			input: 1,
		},
		{
			input:    -1,
			problems: []*Problem{{LevelError, "source-strip-invalid", []interface{}{-1}}},
		},
	} {
		l := linter{}
		l.lintSourceStrip(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintControl(t *testing.T) {
	l := linter{}
	l.lintControl(nil)
	assert.Equal(t, []*Problem{{LevelError, "control-empty", nil}}, l.problems)
}

func TestLintControlDescription(t *testing.T) {
	for _, test := range []struct {
		input    string
		problems []*Problem
	}{
		{
			input: "A valid description",
		},
		{
			input:    "",
			problems: []*Problem{{LevelError, "control-description-empty", nil}},
		},
	} {
		l := linter{}
		l.lintControlDescription(test.input)
		assert.Equal(t, test.problems, l.problems, "value: %q", test.input)
	}
}

func TestLintInstall(t *testing.T) {
	l := linter{}
	l.lintInstall(nil)
	assert.Equal(t, []*Problem{{LevelError, "install-empty", nil}}, l.problems)
}

func TestLintInstallMap(t *testing.T) {
	for _, test := range []struct {
		subkey   string
		input    recipe.InstallMap
		problems []*Problem
	}{
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"/path/to/folder": []recipe.InstallRule{{Pattern: "*"}},
			},
		},
		{
			subkey:   "upstream",
			problems: []*Problem{{LevelError, "install-upstream-empty", nil}},
		},
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"path/to/folder": []recipe.InstallRule{{Pattern: "*"}},
			},
			problems: []*Problem{{LevelError, "install-destination-relative", []interface{}{"path/to/folder"}}},
		},
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"/path/to/folder": nil,
			},
			problems: []*Problem{{LevelError, "install-rule-empty", []interface{}{"/path/to/folder"}}},
		},
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"/path/to/folder": []recipe.InstallRule{{}},
			},
			problems: []*Problem{{LevelError, "install-rule-pattern-empty", []interface{}{"/path/to/folder", 0}}},
		},
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"/path/to/folder": []recipe.InstallRule{{Pattern: "foo", Rename: "a"}, {Pattern: "bar", Rename: "a"}},
			},
			problems: []*Problem{{LevelError, "install-rule-rename-duplicate", []interface{}{"/path/to/folder", "a"}}},
		},
		{
			subkey: "upstream",
			input: recipe.InstallMap{
				"/path/to/folder": []recipe.InstallRule{{Pattern: "foo", ConfFile: true}},
			},
			problems: []*Problem{{LevelWarning, "install-rule-conffile-outside-etc",
				[]interface{}{"/path/to/folder", 0}}},
		},
	} {
		l := linter{}
		l.lintInstallMap(test.subkey, test.input)
		assert.Equal(t, test.problems, l.problems)
	}
}

func TestDirs(t *testing.T) {
	for _, test := range []struct {
		input    []string
		problems []*Problem
	}{
		{
			input: []string{"/path/to/dir"},
		},
		{
			input:    []string{"/path/to/dir", "path/to/another/dir"},
			problems: []*Problem{{LevelError, "dirs-path-relative", []interface{}{"path/to/another/dir"}}},
		},
	} {
		l := linter{}
		l.lintDirs(test.input)
		assert.Equal(t, test.problems, l.problems)
	}
}

func TestLinks(t *testing.T) {
	for _, test := range []struct {
		input    map[string]string
		problems []*Problem
	}{
		{
			input: map[string]string{"/path/to/link": "/path/to/target"},
		},
		{
			input:    map[string]string{"path/to/link": "/path/to/target"},
			problems: []*Problem{{LevelError, "links-destination-relative", []interface{}{"path/to/link"}}},
		},
		{
			input:    map[string]string{"/path/to/link": "path/to/target"},
			problems: []*Problem{{LevelError, "links-source-relative", []interface{}{"path/to/target"}}},
		},
	} {
		l := linter{}
		l.lintLinks(test.input)
		assert.Equal(t, test.problems, l.problems)
	}
}

func TestLintUnsupportedRule(t *testing.T) {
	l := linter{}
	assert.Panics(t, func() { l.emit("unsupported-rule") })
}
