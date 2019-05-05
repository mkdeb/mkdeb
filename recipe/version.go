package recipe

// VersionSupported returns whether or not a recipe version is supported.
func VersionSupported(version int) bool {
	return version == 1
}
