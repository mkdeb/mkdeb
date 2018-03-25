package recipe

import "errors"

var (
	// ErrMissingControl represents a missing control error.
	ErrMissingControl = errors.New("missing control")
	// ErrMissingControlURL represents a missing control description error.
	ErrMissingControlDescription = errors.New("missing control description")
	// ErrMissingDescription represents a missing description error.
	ErrMissingDescription = errors.New("missing description")
	// ErrMissingInstall represents a missing install error.
	ErrMissingInstall = errors.New("missing install")
	// ErrMissingMaintainer represents a missing maintainer error.
	ErrMissingMaintainer = errors.New("missing maintainer")
	// ErrMissingName represents a missing name error.
	ErrMissingName = errors.New("missing name")
	// ErrMissingSource represents a missing source error.
	ErrMissingSource = errors.New("missing source")
	// ErrMissingSourceURL represents a missing source URL error.
	ErrMissingSourceURL = errors.New("missing source URL")
	// ErrUnsupportedVersion represents an unsupported version error.
	ErrUnsupportedVersion = errors.New("unsupported version")
)
