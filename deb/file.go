package deb

import (
	"os"
	"time"
)

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func newFileInfo(name string, size int64, mode os.FileMode, modTime time.Time, isDir bool) *fileInfo {
	return &fileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
	}
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f *fileInfo) IsDir() bool {
	return f.isDir
}

func (f *fileInfo) Sys() interface{} {
	return f
}
