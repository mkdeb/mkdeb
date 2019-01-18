package handler

import (
	"fmt"
	"io"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"mkdeb.sh/archive"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

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
		return errors.Wrap(err, "cannot open upstream archive")
	}
	defer f.Close()

	src, err := archive.NewReader(f, compress)
	if err != nil {
		return errors.Wrap(err, "cannot initialize archive reader")
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

		if path, confFile, ok := recipe.InstallPath(name, recipe.Install.Upstream); ok {
			fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(h.Size)))

			if confFile {
				p.RegisterConfFile(path)
			}

			if h.Mode&os.ModeDir == os.ModeDir {
				if err := p.AddDir(path, h.Mode); err != nil {
					return errors.Wrapf(err, "cannot add %q dir", name)
				}
			} else if h.Mode&os.ModeSymlink == os.ModeSymlink {
				if err := p.AddLink(path, h.LinkName); err != nil {
					return errors.Wrapf(err, "cannot add %q link", name)
				}
			} else {
				if err := p.AddFile(path, src, h.FileInfo()); err != nil {
					return errors.Wrapf(err, "cannot add %q file", name)
				}
			}
		}
	}

	return nil
}
