package deb

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileInfo(t *testing.T) {
	now := time.Now()

	fi := newFileInfo("foo", 123, 0644, now, false)
	assert.Equal(t, "foo", fi.Name())
	assert.Equal(t, int64(123), fi.Size())
	assert.Equal(t, os.FileMode(0644), fi.Mode())
	assert.Equal(t, now, fi.ModTime())
	assert.Equal(t, false, fi.IsDir())
}

func TestFileDir(t *testing.T) {
	now := time.Now()

	fi := newFileInfo("foo", 0, 0755, now, true)
	assert.Equal(t, "foo", fi.Name())
	assert.Equal(t, int64(0), fi.Size())
	assert.Equal(t, os.FileMode(0755), fi.Mode())
	assert.Equal(t, now, fi.ModTime())
	assert.Equal(t, true, fi.IsDir())
}
