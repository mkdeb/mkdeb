package deb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControlEmpty(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Description: 
`

	c := NewControl()
	assert.Equal(t, expected, c.String())
}

func TestControlName(t *testing.T) {
	expected := `Package: foo
Version: 0.0.0
Priority: extra
Architecture: all
Description: 
`

	c := NewControl()
	c.Name = "foo"
	assert.Equal(t, expected, c.String())
}

func TestControlVersion(t *testing.T) {
	expected := `Package: 
Version: 1.2.3
Priority: extra
Architecture: all
Description: 
`

	c := NewControl()
	c.Version = "1.2.3"
	assert.Equal(t, expected, c.String())
}

func TestControlSection(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Section: admin
Priority: extra
Architecture: all
Description: 
`

	c := NewControl()
	c.Section = "admin"
	assert.Equal(t, expected, c.String())
}

func TestControlPriority(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: optional
Architecture: all
Description: 
`

	c := NewControl()
	c.Priority = "optional"
	assert.Equal(t, expected, c.String())
}

func TestControlArchitecture(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: amd64
Description: 
`

	c := NewControl()
	c.Architecture = "amd64"
	assert.Equal(t, expected, c.String())
}

func TestControlDependsSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Depends: foo
Description: 
`

	c := NewControl()
	c.Depends = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlDependsMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Depends: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Depends = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlPreDependsSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Pre-Depends: foo
Description: 
`

	c := NewControl()
	c.PreDepends = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlPreDependsMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Pre-Depends: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.PreDepends = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlRecommendsSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Recommends: foo
Description: 
`

	c := NewControl()
	c.Recommends = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlRecommendsMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Recommends: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Recommends = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlSuggestsSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Suggests: foo
Description: 
`

	c := NewControl()
	c.Suggests = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlSuggestsMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Suggests: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Suggests = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlEnhancesSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Enhances: foo
Description: 
`

	c := NewControl()
	c.Enhances = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlEnhancesMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Enhances: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Enhances = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlBreaksSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Breaks: foo
Description: 
`

	c := NewControl()
	c.Breaks = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlBreaksMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Breaks: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Breaks = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlConflictsSingle(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Conflicts: foo
Description: 
`

	c := NewControl()
	c.Conflicts = []string{"foo"}
	assert.Equal(t, expected, c.String())
}

func TestControlConflictsMultiple(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Conflicts: foo, bar (>= 1.2.3)
Description: 
`

	c := NewControl()
	c.Conflicts = []string{"foo", "bar (>= 1.2.3)"}
	assert.Equal(t, expected, c.String())
}

func TestControlInstalledSize(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Installed-Size: 1234
Description: 
`

	c := NewControl()
	c.InstalledSize = int64(1263616)
	assert.Equal(t, expected, c.String())
}

func TestControlMaintainer(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Maintainer: Foo Bar <foo@example.org>
Description: 
`

	c := NewControl()
	c.Maintainer = "Foo Bar <foo@example.org>"
	assert.Equal(t, expected, c.String())
}

func TestControlDescription(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Description: a short description on the first line
 A long description that should be wrapped once the line length is more than
 76 characters long.
 .
 And with dots separating paragraphs.
`

	c := NewControl()
	c.Description = `a short description on the first line
A long description that should be wrapped once the line length is more than 76 characters long.

And with dots separating paragraphs.
`
	assert.Equal(t, expected, c.String())
}

func TestControlHomepage(t *testing.T) {
	expected := `Package: 
Version: 0.0.0
Priority: extra
Architecture: all
Description: 
Homepage: https://example.org
`

	c := NewControl()
	c.Homepage = "https://example.org"
	assert.Equal(t, expected, c.String())
}
