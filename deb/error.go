package deb

import "golang.org/x/xerrors"

var (
	// ErrInvalidField is an invalid field error.
	ErrInvalidField = xerrors.New("invalid field")
	// ErrInvalidValue is an invalid value error.
	ErrInvalidValue = xerrors.New("invalid value")
)
