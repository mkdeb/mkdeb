package deb

import "errors"

var (
	// ErrInvalidField is an invalid field error.
	ErrInvalidField = errors.New("invalid field")
	// ErrInvalidValue is an invalid value error.
	ErrInvalidValue = errors.New("invalid value")
)
