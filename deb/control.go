package deb

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	wordwrap "github.com/mitchellh/go-wordwrap"
)

// ControlDescriptionWrap is the control description text wrapping value.
const ControlDescriptionWrap = 76

// Control is a Debian control.
type Control struct {
	name          string
	version       string
	section       string
	priority      string
	architecture  string
	depends       []string
	preDepends    []string
	recommends    []string
	suggests      []string
	enhances      []string
	breaks        []string
	conflicts     []string
	installedSize int64
	maintainer    string
	description   string
	homepage      string
}

// NewControl creates a new Debian control instance.
func NewControl() *Control {
	c := &Control{}
	c.Set("Version", "0.0.0")
	c.Set("Priority", "extra")
	c.Set("Architecture", "all")

	return c
}

// Set sets a given control field value.
func (c *Control) Set(key string, value interface{}) error {
	rv := reflect.ValueOf(value)
	f := reflect.ValueOf(c).Elem().FieldByName(keyToFieldName(key))
	if !f.IsValid() {
		return ErrInvalidField
	} else if f.Kind() != rv.Kind() {
		return ErrInvalidValue
	}

	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(rv)

	return nil
}

// String generates a Debian control data string representation.
func (c *Control) String() string {
	var data string

	data += fmt.Sprintf("Package: %s\n", c.name)
	data += fmt.Sprintf("Version: %s\n", c.version)
	if c.section != "" {
		data += fmt.Sprintf("Section: %s\n", c.section)
	}
	data += fmt.Sprintf("Priority: %s\n", c.priority)
	data += fmt.Sprintf("Architecture: %s\n", c.architecture)
	if len(c.depends) > 0 {
		data += fmt.Sprintf("Depends: %s\n", formatDepends(c.depends))
	}
	if len(c.preDepends) > 0 {
		data += fmt.Sprintf("Pre-Depends: %s\n", formatDepends(c.preDepends))
	}
	if len(c.recommends) > 0 {
		data += fmt.Sprintf("Recommends: %s\n", formatDepends(c.recommends))
	}
	if len(c.suggests) > 0 {
		data += fmt.Sprintf("Suggests: %s\n", formatDepends(c.suggests))
	}
	if len(c.enhances) > 0 {
		data += fmt.Sprintf("Enhances: %s\n", formatDepends(c.enhances))
	}
	if len(c.breaks) > 0 {
		data += fmt.Sprintf("Breaks: %s\n", formatDepends(c.breaks))
	}
	if len(c.conflicts) > 0 {
		data += fmt.Sprintf("Conflicts: %s\n", formatDepends(c.conflicts))
	}
	if c.installedSize > 0 {
		data += fmt.Sprintf("Installed-Size: %d\n", c.installedSize/1024)
	}
	if c.maintainer != "" {
		data += fmt.Sprintf("Maintainer: %s\n", c.maintainer)
	}
	data += fmt.Sprintf("Description: %s\n", formatDescription(c.description))
	if c.homepage != "" {
		data += fmt.Sprintf("Homepage: %s\n", c.homepage)
	}

	return data
}

func keyToFieldName(key string) string {
	return strings.ToLower(string(key[0])) + strings.Replace(key[1:], "-", "", -1)
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
