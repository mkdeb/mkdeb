package handler

import (
	"archive/zip"
	"fmt"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// Zip is an upstream source zip handler.
func Zip(p *deb.Package, recipe *recipe.Recipe, path, typ string) error {
	// Create a new reader for the source archive
	r, err := zip.OpenReader(path)
	if err != nil {
		return errors.Wrap(err, "cannot open upstream archive")
	}
	defer r.Close()

	for _, file := range r.File {
		name := file.Name
		if recipe.Source.Strip > 0 {
			name = stripName(name, recipe.Source.Strip)
		}

		if path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream); ok {
			fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(file.UncompressedSize64))

			if confFile {
				p.RegisterConfFile(path)
			}

			f, err := file.Open()
			if err != nil {
				return errors.Wrapf(err, "cannot open %q file", file.Name)
			}
			defer f.Close()

			mode := file.Mode()
			if mode&os.ModeDir == os.ModeDir {
				if err := p.AddDir(path, mode); err != nil {
					return errors.Wrapf(err, "cannot add %q dir", name)
				}
			} else {
				if err := p.AddFile(path, f, file.FileInfo()); err != nil {
					return errors.Wrapf(err, "cannot add %q file", name)
				}
			}
		}
	}

	return nil
}
