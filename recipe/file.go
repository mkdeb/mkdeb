package recipe

import "os"

// File is a recipe file.
type File struct {
	Path     string
	FileInfo os.FileInfo
}
