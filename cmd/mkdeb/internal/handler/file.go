package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// File is an upstream source file handler.
func File(p *deb.Package, recipe *recipe.Recipe, filePath, typ string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat upstream file: %w", err)
	}

	if fi.IsDir() {
		return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() {
				return nil
			}
			return file(p, recipe, strings.TrimPrefix(path, filePath+"/"), path, info)
		})
	}

	return file(p, recipe, filepath.Base(filePath), filePath, fi)
}

func file(p *deb.Package, recipe *recipe.Recipe, name, filePath string, fi os.FileInfo) error {
	path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream)
	if ok {
		fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(fi.Size())))

		if confFile {
			p.RegisterConfFile(path)
		}

		// Create a new reader for the source file
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("cannot open upstream file: %w", err)
		}
		defer f.Close()

		err = p.AddFile(path, f, fi)
		if err != nil {
			return fmt.Errorf("cannot add %q file: %w", name, err)
		}
	}

	return nil
}
