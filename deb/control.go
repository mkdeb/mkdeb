package deb

import (
	"fmt"
	"strings"

	wordwrap "github.com/mitchellh/go-wordwrap"
)

// ControlDescriptionWrap is the control description text wrapping value.
const ControlDescriptionWrap = 76

// Control is a Debian control.
type Control struct {
	Name          string
	Version       string
	Section       string
	Priority      string
	Architecture  string
	Depends       []string
	PreDepends    []string
	Recommends    []string
	Suggests      []string
	Enhances      []string
	Breaks        []string
	Conflicts     []string
	InstalledSize int64
	Maintainer    string
	Description   string
	Homepage      string
}

// NewControl creates a new Debian control instance.
func NewControl() *Control {
	return &Control{
		Version:      "0.0.0",
		Priority:     "extra",
		Architecture: "all",
	}
}

// String generates a Debian control data string representation.
func (c *Control) String() string {
	var data string

	data += fmt.Sprintf("Package: %s\n", c.Name)
	data += fmt.Sprintf("Version: %s\n", c.Version)
	if c.Section != "" {
		data += fmt.Sprintf("Section: %s\n", c.Section)
	}
	data += fmt.Sprintf("Priority: %s\n", c.Priority)
	data += fmt.Sprintf("Architecture: %s\n", c.Architecture)
	if len(c.Depends) > 0 {
		data += fmt.Sprintf("Depends: %s\n", formatDepends(c.Depends))
	}
	if len(c.PreDepends) > 0 {
		data += fmt.Sprintf("Pre-Depends: %s\n", formatDepends(c.PreDepends))
	}
	if len(c.Recommends) > 0 {
		data += fmt.Sprintf("Recommends: %s\n", formatDepends(c.Recommends))
	}
	if len(c.Suggests) > 0 {
		data += fmt.Sprintf("Suggests: %s\n", formatDepends(c.Suggests))
	}
	if len(c.Enhances) > 0 {
		data += fmt.Sprintf("Enhances: %s\n", formatDepends(c.Enhances))
	}
	if len(c.Breaks) > 0 {
		data += fmt.Sprintf("Breaks: %s\n", formatDepends(c.Breaks))
	}
	if len(c.Conflicts) > 0 {
		data += fmt.Sprintf("Conflicts: %s\n", formatDepends(c.Conflicts))
	}
	if c.InstalledSize > 0 {
		data += fmt.Sprintf("Installed-Size: %d\n", c.InstalledSize/1024)
	}
	if c.Maintainer != "" {
		data += fmt.Sprintf("Maintainer: %s\n", c.Maintainer)
	}
	data += fmt.Sprintf("Description: %s\n", formatDescription(c.Description))
	if c.Homepage != "" {
		data += fmt.Sprintf("Homepage: %s\n", c.Homepage)
	}

	return data
}

func formatDepends(depends []string) string {
	return strings.Join(depends, ", ")
}

func formatDescription(desc string) string {
	desc = wordwrap.WrapString(strings.TrimSpace(desc), ControlDescriptionWrap)
	desc = strings.Replace(desc, "\n\n", "\n.\n", -1)
	desc = strings.Replace(desc, "\n", "\n ", -1)
	return desc
}
