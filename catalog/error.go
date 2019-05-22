package catalog

import "golang.org/x/xerrors"

// Errors:
var (
	ErrAlreadyUpToDate    = xerrors.New("already up-to-date")
	ErrRecipeNotFound     = xerrors.New("recipe not found")
	ErrRepositoryExist    = xerrors.New("repository already exists")
	ErrRepositoryNotExist = xerrors.New("repository does not exist")
)
