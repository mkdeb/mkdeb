package deb

// Version represents a Debian package version instance.
type Version struct {
	Upstream string
	Revision string
}

// NewVersion creates a Debian package version instance.
func NewVersion(upstream, revision string) *Version {
	return &Version{upstream, revision}
}

func (v *Version) String() string {
	if v.Revision != "" {
		return v.Upstream + v.Revision
	}

	return v.Upstream
}
