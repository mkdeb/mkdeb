package deb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionUpstream(t *testing.T) {
	assert.Equal(t, "1.2.3", NewVersion("1.2.3", "").String())
}

func TestVersionRevision(t *testing.T) {
	assert.Equal(t, "1.2.3~mkdeb1", NewVersion("1.2.3", "~mkdeb1").String())
}
