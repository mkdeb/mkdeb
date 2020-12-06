package catalog

import "errors"

// Errors:
var (
	ErrAlreadyUpToDate    = errors.New("already up-to-date")
	ErrRecipeNotFound     = errors.New("recipe not found")
	ErrRepositoryExist    = errors.New("repository already exists")
	ErrRepositoryNotExist = errors.New("repository does not exist")
)
