package repository

import git "gopkg.in/src-d/go-git.v4"

var (
	// ErrAlreadyUpToDate represents an already up-to-date repository error.
	ErrAlreadyUpToDate = git.NoErrAlreadyUpToDate
)
