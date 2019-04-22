package recipe

import "os"

// File is a recipe file instance.
type File struct {
	Path     string
	FileInfo os.FileInfo
}
