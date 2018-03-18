package archive

import "errors"

var (
	// ErrUnsupportedCompress represents an unsupported compression format error.
	ErrUnsupportedCompress = errors.New("unsupported compression")
)
