package archive

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderFileInfo(t *testing.T) {
	h := &Header{
		Name:    "dir",
		Size:    0,
		Mode:    os.FileMode(0755) | os.ModeDir,
		ModTime: testTime,
	}

	fi := h.FileInfo()
	assert.Equal(t, "dir", fi.Name())
	assert.Equal(t, int64(0), fi.Size())
	assert.Equal(t, os.FileMode(0755)|os.ModeDir, fi.Mode())
	assert.Equal(t, testTime, fi.ModTime())
	assert.Equal(t, h, fi.Sys())
	assert.True(t, fi.IsDir())
}
