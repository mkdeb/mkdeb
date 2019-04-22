package repository

import git "gopkg.in/src-d/go-git.v4"

var (
	// ErrAlreadyUpToDate is an already up-to-date repository error.
	ErrAlreadyUpToDate = git.NoErrAlreadyUpToDate
)
