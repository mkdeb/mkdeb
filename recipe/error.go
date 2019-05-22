package recipe

import "golang.org/x/xerrors"

var (
	// ErrMissingControl is a missing control error.
	ErrMissingControl = xerrors.New("missing control")
	// ErrMissingControlDescription is a missing control description error.
	ErrMissingControlDescription = xerrors.New("missing control description")
	// ErrMissingDescription is a missing description error.
	ErrMissingDescription = xerrors.New("missing description")
	// ErrMissingInstall is a missing install error.
	ErrMissingInstall = xerrors.New("missing install")
	// ErrMissingMaintainer is a missing maintainer error.
	ErrMissingMaintainer = xerrors.New("missing maintainer")
	// ErrMissingName is a missing name error.
	ErrMissingName = xerrors.New("missing name")
	// ErrMissingSource is a missing source error.
	ErrMissingSource = xerrors.New("missing source")
	// ErrMissingSourceURL is a missing source URL error.
	ErrMissingSourceURL = xerrors.New("missing source URL")
	// ErrUnsupportedVersion is an unsupported version error.
	ErrUnsupportedVersion = xerrors.New("unsupported version")
)
