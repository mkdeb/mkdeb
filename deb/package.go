package deb

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blakesmith/ar"
	"golang.org/x/xerrors"
	"mkdeb.sh/archive"
)

// Package is a Debian package.
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
func NewPackage(name, arch, version string, epoch uint, revision int) (*Package, error) {
	// Initialize archives that will receive internal package data
	control, err := archive.NewWriterBuffer(archive.CompressGzip)
	if err != nil {
		return nil, xerrors.Errorf("cannot create control archive: %w", err)
	}

	data, err := archive.NewWriterBuffer(archive.CompressXZ)
	if err != nil {
		return nil, xerrors.Errorf("cannot create data archive: %w", err)
	}

	return &Package{
		Name:    name,
		Arch:    arch,
		Version: NewVersion(epoch, version, fmt.Sprintf("1~mkdeb%d", revision)),
		Control: NewControl(),

		modTime: time.Now(),
		dirs:    map[string]struct{}{},
		control: control,
		data:    data,
		md5sums: bytes.NewBuffer(nil),
	}, nil
}

// AddControlFile appends a new file to the internal control archive.
func (p *Package) AddControlFile(name string, r io.Reader, fi os.FileInfo) error {
	err := p.control.WriteHeader(&archive.Header{
		Name:    name,
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		User:    "root",
		Group:   "root",
		ModTime: fi.ModTime(),
	})
	if err != nil {
		return err
	}

	_, err = io.Copy(p.control, r)
	return err
}

// AddDir appends a new directory to the internal data archive.
func (p *Package) AddDir(path string, mode os.FileMode) error {
	err := p.ensureParent(path)
	if err != nil {
		return err
	}

	p.dirs[path] = struct{}{}

	return p.data.WriteHeader(&archive.Header{
		Name:    "." + strings.TrimRight(path, "/") + "/",
		Mode:    mode | os.ModeDir,
		User:    "root",
		Group:   "root",
		ModTime: p.modTime,
	})
}

// AddFile appends a new file to the internal data archive.
func (p *Package) AddFile(path string, r io.Reader, fi os.FileInfo) error {
	digest := md5.New()

	err := p.ensureParent(path)
	if err != nil {
		return err
	}

	// Append file and its MD5 sum to the sums listing if regular file
	size := fi.Size()
	p.Control.installedSize += size

	err = p.data.WriteHeader(&archive.Header{
		Name:    "." + path,
		Size:    size,
		Mode:    fi.Mode(),
		User:    "root",
		Group:   "root",
		ModTime: fi.ModTime(),
	})
	if err != nil {
		return err
	}

	_, err = io.Copy(p.data, io.TeeReader(r, digest))
	if err != nil {
		return err
	}

	fmt.Fprintf(p.md5sums, "%x  %s\n", digest.Sum(nil), path[1:])

	return nil
}

// AddLink appends a new symbolic link to the internal data archive.
func (p *Package) AddLink(dst, src string) error {
	err := p.ensureParent(dst)
	if err != nil {
		return err
	}

	return p.data.WriteHeader(&archive.Header{
		Name:     "." + dst,
		LinkName: src,
		Mode:     os.FileMode(0777) | os.ModeSymlink,
		User:     "root",
		Group:    "root",
		ModTime:  p.modTime,
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
	err := p.Control.Set("Name", p.Name)
	if err != nil {
		return xerrors.Errorf("cannot set name: %w", err)
	}

	err = p.Control.Set("Version", p.Version.String())
	if err != nil {
		return xerrors.Errorf("cannot set version: %w", err)
	}

	err = p.Control.Set("Architecture", p.Arch)
	if err != nil {
		return xerrors.Errorf("cannot set architecture: %w", err)
	}

	if len(p.confFiles) > 0 {
		src = bytes.NewBuffer([]byte(strings.Join(p.confFiles, "\n") + "\n"))
		err = p.AddControlFile("conffiles", src, newFileInfo("conffiles", int64(src.Len()), 0644, now, false))
		if err != nil {
			return xerrors.Errorf("cannot add \"conffiles\" file: %w", err)
		}
	}

	src = bytes.NewBuffer([]byte(p.Control.String()))
	err = p.AddControlFile("control", src, newFileInfo("control", int64(src.Len()), 0644, now, false))
	if err != nil {
		return xerrors.Errorf("cannot add \"control\" file: %w", err)
	}

	src = bytes.NewBuffer(p.md5sums.Bytes())
	err = p.AddControlFile("md5sums", src, newFileInfo("md5sums", int64(src.Len()), 0644, now, false))
	if err != nil {
		return xerrors.Errorf("cannot add \"md5sums\" file: %w", err)
	}

	// Close internal archives prior to write their content to prevent incomplete data
	err = p.control.Close()
	if err != nil {
		return xerrors.Errorf("cannot close control archive: %w", err)
	}

	err = p.data.Close()
	if err != nil {
		return xerrors.Errorf("cannot close data archive: %w", err)
	}

	// Initialize archive file and append content
	p.writer = ar.NewWriter(w)
	p.writer.WriteGlobalHeader()

	err = p.append("debian-binary", []byte("2.0\n"), now)
	if err != nil {
		return xerrors.Errorf("cannot append debian-binary: %w", err)
	}

	err = p.append("control.tar.gz", p.control.Bytes(), now)
	if err != nil {
		return xerrors.Errorf("cannot append control.tar.gz: %w", err)
	}

	err = p.append("data.tar.xz", p.data.Bytes(), now)
	if err != nil {
		return xerrors.Errorf("cannot append data.tar.xz: %w", err)
	}

	return nil
}

func (p *Package) ensureParent(path string) error {
	// Check for parent directory
	dirPath := filepath.Dir(path)
	if dirPath == "/" {
		return nil
	}

	_, ok := p.dirs[dirPath]
	if !ok {
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
