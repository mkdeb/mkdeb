package deb

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testPkg *Package

func init() {
	testPkg, _ = NewPackage("foo", "all", "1.2.3", 0, 1)
}

func TestPackage(t *testing.T) {
	assert.Equal(t, "foo", testPkg.Name)
	assert.Equal(t, "all", testPkg.Arch)
	assert.Equal(t, &Version{0, "1.2.3", "1~mkdeb1"}, testPkg.Version)
}

func TestPackageAddControlFile(t *testing.T) {
	err := testPkg.AddControlFile(
		"postinst",
		strings.NewReader("#!/bin/sh\ntrue\n"),
		newFileInfo("postinst", 15, os.FileMode(0755), time.Now(), false),
	)
	assert.Nil(t, err)
	assert.True(t, len(testPkg.control.Bytes()) > 0)
}

func TestPackageAddDir(t *testing.T) {
	err := testPkg.AddDir("/path/to/dir", os.FileMode(0755))
	assert.Nil(t, err)
	assert.Contains(t, testPkg.dirs, "/path/to/dir")
	assert.True(t, len(testPkg.data.Bytes()) > 0)
	assert.Equal(t, int64(0), testPkg.Control.installedSize)
}

func TestPackageAddFile(t *testing.T) {
	err := testPkg.AddFile(
		"/path/to/file",
		strings.NewReader("# noop\n"),
		newFileInfo("/path/to/file", 7, os.FileMode(0644), time.Now(), false),
	)
	assert.Nil(t, err)
	assert.True(t, len(testPkg.data.Bytes()) > 0)
	assert.Equal(t, int64(7), testPkg.Control.installedSize)
}

func TestPackageAddLink(t *testing.T) {
	err := testPkg.AddLink("/path/to/link", "/path/to/target")
	assert.Nil(t, err)
	assert.True(t, len(testPkg.data.Bytes()) > 0)
	assert.Equal(t, int64(7), testPkg.Control.installedSize)
}

func TestPackageRegisterConfFile(t *testing.T) {
	testPkg.RegisterConfFile("/path/to/conffile")
	assert.Contains(t, testPkg.confFiles, "/path/to/conffile")
}

func TestPackageWrite(t *testing.T) {
	err := testPkg.Write(ioutil.Discard)
	assert.Nil(t, err)
}
