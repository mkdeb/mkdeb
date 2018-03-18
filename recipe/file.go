package recipe

import "os"

// File represents a recipe file instance.
type File struct {
	Path     string
	FileInfo os.FileInfo
}
