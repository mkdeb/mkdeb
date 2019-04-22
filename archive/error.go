package archive

import "errors"

var (
	// ErrUnsupportedCompress is an unsupported compression format error.
	ErrUnsupportedCompress = errors.New("unsupported compression")
)
