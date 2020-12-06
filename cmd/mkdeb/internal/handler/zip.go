package handler

import (
	"archive/zip"
	"fmt"
	"os"

	humanize "github.com/dustin/go-humanize"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// Zip is an upstream source zip handler.
func Zip(p *deb.Package, recipe *recipe.Recipe, path, typ string) error {
	// Create a new reader for the source archive
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("cannot open upstream archive: %w", err)
	}
	defer r.Close()

	for _, file := range r.File {
		name := file.Name
		if recipe.Source.Strip > 0 {
			name = stripName(name, recipe.Source.Strip)
		}

		path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream)
		if ok {
			fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(file.UncompressedSize64))

			if confFile {
				p.RegisterConfFile(path)
			}

			f, err := file.Open()
			if err != nil {
				return fmt.Errorf("cannot open %q file: %w", file.Name, err)
			}
			defer f.Close()

			mode := file.Mode()
			if mode&os.ModeDir == os.ModeDir {
				err = p.AddDir(path, mode)
				if err != nil {
					return fmt.Errorf("cannot add %q dir: %w", name, err)
				}
			} else {
				err = p.AddFile(path, f, file.FileInfo())
				if err != nil {
					return fmt.Errorf("cannot add %q file: %w", name, err)
				}
			}
		}
	}

	return nil
}
