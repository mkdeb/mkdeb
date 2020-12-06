package recipe

import "errors"

var (
	// ErrMissingControl is a missing control error.
	ErrMissingControl = errors.New("missing control")
	// ErrMissingControlDescription is a missing control description error.
	ErrMissingControlDescription = errors.New("missing control description")
	// ErrMissingDescription is a missing description error.
	ErrMissingDescription = errors.New("missing description")
	// ErrMissingInstall is a missing install error.
	ErrMissingInstall = errors.New("missing install")
	// ErrMissingMaintainer is a missing maintainer error.
	ErrMissingMaintainer = errors.New("missing maintainer")
	// ErrMissingName is a missing name error.
	ErrMissingName = errors.New("missing name")
	// ErrMissingSource is a missing source error.
	ErrMissingSource = errors.New("missing source")
	// ErrMissingSourceURL is a missing source URL error.
	ErrMissingSourceURL = errors.New("missing source URL")
	// ErrUnsupportedVersion is an unsupported version error.
	ErrUnsupportedVersion = errors.New("unsupported version")
)
