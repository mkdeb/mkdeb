package deb

import "fmt"

// Version represents a Debian package version instance.
type Version struct {
	Epoch    uint
	Upstream string
	Revision string
}

// NewVersion creates a Debian package version instance.
func NewVersion(epoch uint, upstream, revision string) *Version {
	return &Version{epoch, upstream, revision}
}

func (v *Version) String() string {
	var s string

	if v.Epoch > 0 {
		s += fmt.Sprintf("%d:", v.Epoch)
	}

	s += v.Upstream

	if v.Revision != "" {
		s += fmt.Sprintf("-%s", v.Revision)
	}

	return s
}
