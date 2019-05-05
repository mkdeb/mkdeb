package handler

import (
	"fmt"
	"os"
	"path/filepath"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// File is an upstream source file handler.
func File(p *deb.Package, recipe *recipe.Recipe, filePath, typ string) error {
	name := filepath.Base(filePath)

	path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream)
	if ok {
		fi, err := os.Stat(filePath)
		if err != nil {
			return errors.Wrap(err, "cannot stat upstream file")
		}

		fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(fi.Size())))

		if confFile {
			p.RegisterConfFile(path)
		}

		// Create a new reader for the source file
		f, err := os.Open(filePath)
		if err != nil {
			return errors.Wrap(err, "cannot open upstream file")
		}
		defer f.Close()

		err = p.AddFile(path, f, fi)
		if err != nil {
			return errors.Wrapf(err, "cannot add %q file", name)
		}
	}

	return nil
}
