package archive

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterGzip(t *testing.T) {
	testWriter(t, CompressGzip)
}

func TestWriterXZ(t *testing.T) {
	testWriter(t, CompressXZ)
}

func TestWriterUnsupported(t *testing.T) {
	_, err := NewWriterBuffer(CompressBzip2)
	assert.Equal(t, ErrUnsupportedCompress, err)
}

func testWriter(t *testing.T, compress int) {
	w, err := NewWriterBuffer(compress)
	assert.Nil(t, err)
	defer w.Close()

	err = w.WriteHeader(&Header{
		Name:    "dir",
		Size:    0,
		Mode:    os.FileMode(0755) | os.ModeDir,
		ModTime: testTime,
	})
	assert.Nil(t, err)

	err = w.WriteHeader(&Header{
		Name:    "file1",
		Size:    4,
		Mode:    os.FileMode(0644),
		ModTime: testTime,
	})
	assert.Nil(t, err)
	n, err := w.Write([]byte("foo\n"))
	assert.Equal(t, 4, n)
	assert.Nil(t, err)

	err = w.WriteHeader(&Header{
		Name:    "file2",
		Size:    0,
		Mode:    os.FileMode(0644),
		ModTime: testTime,
	})
	assert.Nil(t, err)

	err = w.WriteHeader(&Header{
		Name:     "link",
		LinkName: "file1",
		Size:     0,
		Mode:     os.FileMode(0755) | os.ModeSymlink,
		ModTime:  testTime,
	})
	assert.Nil(t, err)
}
