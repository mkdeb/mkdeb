package archive

import (
	"archive/tar"
	"os"
	"path"
	"time"
)

// Header represents an archive file header.
type Header struct {
	Name     string
	LinkName string
	Size     int64
	Mode     os.FileMode
	User     string
	Group    string
	ModTime  time.Time
}

// FileInfo returns an os.FileInfo for the archive header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// TarHeader returns an tar.Header for the archive header.
func (h *Header) TarHeader() *tar.Header {
	var tf byte

	switch {
	case h.Mode&os.ModeDir == os.ModeDir:
		tf = tar.TypeDir

	case h.Mode&os.ModeSymlink == os.ModeSymlink:
		tf = tar.TypeSymlink

	default:
		tf = tar.TypeReg
	}

	return &tar.Header{
		Typeflag: tf,
		Name:     h.Name,
		Linkname: h.LinkName,
		Size:     h.Size,
		Mode:     int64(h.Mode),
		Uname:    h.User,
		Gname:    h.Group,
		ModTime:  h.ModTime,
		Format:   tar.FormatGNU,
	}
}

type headerFileInfo struct {
	h *Header
}

func (fi headerFileInfo) Name() string {
	return path.Base(fi.h.Name)
}

func (fi headerFileInfo) Size() int64 {
	return fi.h.Size
}

func (fi headerFileInfo) Mode() os.FileMode {
	return os.FileMode(fi.h.Mode)
}

func (fi headerFileInfo) ModTime() time.Time {
	return fi.h.ModTime
}

func (fi headerFileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi headerFileInfo) Sys() interface{} {
	return fi.h
}
