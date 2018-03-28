package deb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionUpstream(t *testing.T) {
	assert.Equal(t, "1.2.3", NewVersion(0, "1.2.3", "").String())
}

func TestVersionEpoch(t *testing.T) {
	assert.Equal(t, "1:1.2.3", NewVersion(1, "1.2.3", "").String())
}

func TestVersionRevision(t *testing.T) {
	assert.Equal(t, "1.2.3-1", NewVersion(0, "1.2.3", "1").String())
}

func TestVersionEpochRevision(t *testing.T) {
	assert.Equal(t, "1:1.2.3-1", NewVersion(1, "1.2.3", "1").String())
}
