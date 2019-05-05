package archive

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReaderNone(t *testing.T) {
	f, err := os.Open("testdata/data.tar")
	assert.Nil(t, err)

	testReader(t, f, CompressNone)
}

func TestReaderBzip2(t *testing.T) {
	f, err := os.Open("testdata/data.tar.bz2")
	assert.Nil(t, err)

	testReader(t, f, CompressBzip2)
}

func TestReaderGzip(t *testing.T) {
	f, err := os.Open("testdata/data.tar.gz")
	assert.Nil(t, err)

	testReader(t, f, CompressGzip)
}

func TestReaderXZ(t *testing.T) {
	f, err := os.Open("testdata/data.tar.xz")
	assert.Nil(t, err)

	testReader(t, f, CompressXZ)
}

func TestReaderUnsupported(t *testing.T) {
	r, err := NewReader(nil, -1)
	assert.Nil(t, r)
	assert.Equal(t, ErrUnsupportedCompress, err)
}

func testReader(t *testing.T, f *os.File, compress int) {
	var b []byte

	r, err := NewReader(f, compress)
	assert.Nil(t, err)
	defer r.Close()

	h, err := r.Next()
	assert.Nil(t, err)
	assert.Equal(t, "dir/", h.Name)
	assert.Equal(t, int64(0), h.Size)
	assert.Equal(t, os.FileMode(0755)|os.ModeDir, h.Mode)
	assert.Equal(t, testTime, h.ModTime.UTC())
	n, err := r.Read(b)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)

	h, err = r.Next()
	assert.Nil(t, err)
	assert.Equal(t, "file1", h.Name)
	assert.Equal(t, int64(4), h.Size)
	assert.Equal(t, os.FileMode(0644), h.Mode)
	assert.Equal(t, testTime, h.ModTime.UTC())
	n, err = r.Read(b)
	assert.Equal(t, 0, n)
	assert.Nil(t, err)

	h, err = r.Next()
	assert.Nil(t, err)
	assert.Equal(t, "file2", h.Name)
	assert.Equal(t, int64(0), h.Size)
	assert.Equal(t, os.FileMode(0644), h.Mode)
	assert.Equal(t, testTime, h.ModTime.UTC())
	n, err = r.Read(b)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)

	h, err = r.Next()
	assert.Nil(t, err)
	assert.Equal(t, "link", h.Name)
	assert.Equal(t, int64(0), h.Size)
	assert.Equal(t, os.FileMode(0755)|os.ModeSymlink, h.Mode)
	assert.Equal(t, testTime, h.ModTime.UTC())
	n, err = r.Read(b)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)

	_, err = r.Next()
	assert.Equal(t, io.EOF, err)
}
