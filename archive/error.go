package archive

import "golang.org/x/xerrors"

var (
	// ErrUnsupportedCompress is an unsupported compression format error.
	ErrUnsupportedCompress = xerrors.New("unsupported compression")
)
