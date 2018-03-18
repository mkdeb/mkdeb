package deb

import "errors"

var (
	// ErrInvalidField represents an invalid field error.
	ErrInvalidField = errors.New("invalid field")
	// ErrInvalidValue represents an invalid value error.
	ErrInvalidValue = errors.New("invalid value")
)
