package deb

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blakesmith/ar"
	"github.com/pkg/errors"
	"mkdeb.sh/archive"
)

// Package represents a Debian package instance.
type Package struct {
	Name    string
	Arch    string
	Version *Version
	Control *Control

	modTime   time.Time
	dirs      map[string]struct{}
	control   *archive.WriterBuffer
	data      *archive.WriterBuffer
	md5sums   *bytes.Buffer
	confFiles []string
	writer    *ar.Writer
}

// NewPackage creates a new Debian package instance.
func NewPackage(name, arch, version string, revision int) (*Package, error) {
	// Initialize archives that will receive internal package data
	control, err := archive.NewWriterBuffer(archive.CompressGzip)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create control archive")
	}

	data, err := archive.NewWriterBuffer(archive.CompressXZ)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create data archive")
	}

	return &Package{
		Name:    name,
		Arch:    arch,
		Version: NewVersion(version, fmt.Sprintf("~mkdeb%d", revision)),
		Control: NewControl(),

		modTime: time.Now(),
		dirs:    map[string]struct{}{},
		control: control,
		data:    data,
		md5sums: bytes.NewBuffer(nil),
	}, nil
}

// AddControlFile appends a new file to the internal control archive.
func (p *Package) AddControlFile(name string, r io.Reader, info os.FileInfo) error {
	err := p.control.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Size:     info.Size(),
		Mode:     int64(info.Mode()),
		Uname:    "root",
		Gname:    "root",
		ModTime:  info.ModTime(),
		Format:   tar.FormatGNU,
	})
	if err != nil {
		return err
	}

	_, err = io.Copy(p.control, r)
	return err
}

// AddDir appends a new directory to the internal data archive.
func (p *Package) AddDir(path string, mode os.FileMode) error {
	if err := p.ensureParent(path); err != nil {
		return err
	}

	p.dirs[path] = struct{}{}

	return p.data.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "." + path,
		Mode:     int64(mode),
		Uname:    "root",
		Gname:    "root",
		ModTime:  p.modTime,
		Format:   tar.FormatGNU,
	})
}

// AddFile appends a new file to the internal data archive.
func (p *Package) AddFile(path string, r io.Reader, info os.FileInfo) error {
	digest := md5.New()

	if err := p.ensureParent(path); err != nil {
		return err
	}

	// Append file and its MD5 sum to the sums listing if regular file
	size := info.Size()
	p.Control.installedSize += size

	if err := p.data.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "." + path,
		Size:     size,
		Mode:     int64(info.Mode()),
		Uname:    "root",
		Gname:    "root",
		ModTime:  info.ModTime(),
		Format:   tar.FormatGNU,
	}); err != nil {
		return err
	}

	if _, err := io.Copy(p.data, io.TeeReader(r, digest)); err != nil {
		return err
	}

	fmt.Fprintf(p.md5sums, "%x  %s\n", digest.Sum(nil), path[1:])

	return nil
}

// AddLink appends a new symbolic link to the internal data archive.
func (p *Package) AddLink(dst, src string) error {
	if err := p.ensureParent(dst); err != nil {
		return err
	}

	return p.data.WriteHeader(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     "." + dst,
		Linkname: src,
		Mode:     0777,
		Uname:    "root",
		Gname:    "root",
		ModTime:  p.modTime,
		Format:   tar.FormatGNU,
	})
}

// RegisterConfFile registers a new configuration file path.
func (p *Package) RegisterConfFile(path string) {
	p.confFiles = append(p.confFiles, path)
}

// Write write the content of the Debian package to a given io.Writer.
func (p *Package) Write(w io.Writer) error {
	var src *bytes.Buffer

	now := time.Now()

	// Add generated control files
	if err := p.Control.Set("Name", p.Name); err != nil {
		return errors.Wrap(err, "cannot set name")
	} else if err = p.Control.Set("Version", p.Version.String()); err != nil {
		return errors.Wrap(err, "cannot set version")
	} else if err = p.Control.Set("Architecture", p.Arch); err != nil {
		return errors.Wrap(err, "cannot set architecture")
	}

	if len(p.confFiles) > 0 {
		src = bytes.NewBuffer([]byte(strings.Join(p.confFiles, "\n") + "\n"))
		if err := p.AddControlFile(
			"conffiles",
			src,
			newFileInfo("conffiles", int64(src.Len()), 0644, now, false),
		); err != nil {
			return errors.Wrap(err, "cannot add \"conffiles\" file")
		}
	}

	src = bytes.NewBuffer([]byte(p.Control.String()))
	if err := p.AddControlFile(
		"control",
		src,
		newFileInfo("control", int64(src.Len()), 0644, now, false),
	); err != nil {
		return errors.Wrap(err, "cannot add \"control\" file")
	}

	src = bytes.NewBuffer(p.md5sums.Bytes())
	if err := p.AddControlFile(
		"md5sums",
		src,
		newFileInfo("md5sums", int64(src.Len()), 0644, now, false),
	); err != nil {
		return errors.Wrap(err, "cannot add \"md5sums\" file")
	}

	// Close internal archives prior to write their content to prevent incomplete data
	if err := p.control.Close(); err != nil {
		return errors.Wrap(err, "cannot close control archive")
	} else if err := p.data.Close(); err != nil {
		return errors.Wrap(err, "cannot close data archive")
	}

	// Initialize archive file and append content
	p.writer = ar.NewWriter(w)
	p.writer.WriteGlobalHeader()

	if err := p.append("debian-binary", []byte("2.0\n"), now); err != nil {
		return errors.Wrap(err, "cannot append debian-binary")
	} else if err := p.append("control.tar.gz", p.control.Bytes(), now); err != nil {
		return errors.Wrap(err, "cannot append control.tar.gz")
	} else if err := p.append("data.tar.xz", p.data.Bytes(), now); err != nil {
		return errors.Wrap(err, "cannot append data.tar.xz")
	}

	return nil
}

func (p *Package) ensureParent(path string) error {
	// Check for parent directory
	dirPath := filepath.Dir(path)
	if dirPath == "/" {
		return nil
	}

	if _, ok := p.dirs[dirPath]; !ok {
		return p.AddDir(dirPath, 0755)
	}

	return nil
}

func (p *Package) append(name string, b []byte, modTime time.Time) error {
	err := p.writer.WriteHeader(&ar.Header{
		Name:    name,
		ModTime: modTime,
		Mode:    0644,
		Size:    int64(len(b)),
	})
	if err != nil {
		return err
	}

	_, err = p.writer.Write(b)
	return err
}
