package handler

import (
	"fmt"
	"io"
	"os"

	humanize "github.com/dustin/go-humanize"
	"golang.org/x/xerrors"
	"mkdeb.sh/archive"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// Tar is an upstream source tar handler.
func Tar(p *deb.Package, recipe *recipe.Recipe, path, typ string) error {
	var compress int

	switch typ {
	case "gzip":
		compress = archive.CompressGzip

	case "x-bzip2":
		compress = archive.CompressBzip2

	case "x-tar":
		compress = archive.CompressNone

	case "x-xz":
		compress = archive.CompressXZ
	}

	// Create a new reader for the source archive
	f, err := os.Open(path)
	if err != nil {
		return xerrors.Errorf("cannot open upstream archive: %w", err)
	}
	defer f.Close()

	src, err := archive.NewReader(f, compress)
	if err != nil {
		return xerrors.Errorf("cannot initialize archive reader: %w", err)
	}
	defer src.Close()

	for {
		h, err := src.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		name := h.Name
		if recipe.Source.Strip > 0 {
			name = stripName(name, recipe.Source.Strip)
		}

		path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream)
		if ok {
			fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(h.Size)))

			if confFile {
				p.RegisterConfFile(path)
			}

			switch {
			case h.Mode&os.ModeDir == os.ModeDir:
				err = p.AddDir(path, h.Mode)
				if err != nil {
					return xerrors.Errorf("cannot add %q dir: %w", name, err)
				}

			case h.Mode&os.ModeSymlink == os.ModeSymlink:
				err = p.AddLink(path, h.LinkName)
				if err != nil {
					return xerrors.Errorf("cannot add %q link: %w", name, err)
				}

			default:
				err = p.AddFile(path, src, h.FileInfo())
				if err != nil {
					return xerrors.Errorf("cannot add %q file: %w", name, err)
				}
			}
		}
	}

	return nil
}
