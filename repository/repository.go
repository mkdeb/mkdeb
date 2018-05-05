package repository

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"mkdeb.sh/recipe"
)

const (
	recipesRepositoryURL = "https://github.com/mkdeb/recipes.git"
	recipesRepositoryRef = "refs/heads/master"
)

// Repository represents a recipes repository instance.
type Repository struct {
	Path string
}

// NewRepository creates a new instance of a recipes repository given a root path.
func NewRepository(path string) *Repository {
	return &Repository{
		Path: path,
	}
}

// Exists returns whether or not a repository exists.
func (r *Repository) Exists() bool {
	_, err := os.Stat(r.Path)
	return err == nil
}

// Init initializes the recipes repository.
func (r *Repository) Init(progress io.Writer) error {
	_, err := git.PlainClone(r.Path, false, &git.CloneOptions{
		URL:           recipesRepositoryURL,
		ReferenceName: recipesRepositoryRef,
		SingleBranch:  true,
		Progress:      progress,
	})

	return err
}

// Recipe loads a recipe from the repository.
func (r *Repository) Recipe(name string) (*recipe.Recipe, error) {
	return recipe.LoadRecipe(filepath.Join(r.Path, string(name[0]), name))
}

// Update updates the recipes repository from the remote origin.
func (r *Repository) Update(progress io.Writer, force bool) error {
	if force {
		if err := os.RemoveAll(r.Path); err != nil {
			return errors.Wrap(err, "cannot reset repository")
		}
	}

	repo, err := git.PlainOpen(r.Path)
	if err != nil {
		return errors.Wrap(err, "cannot open repository")
	}

	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "cannot get worktree")
	}

	err = wt.Pull(&git.PullOptions{
		Progress: progress,
	})
	if err == git.NoErrAlreadyUpToDate {
		return ErrAlreadyUpToDate
	} else if err != nil {
		return err
	}

	return nil
}

// Walk walks the repository calling a function for each recipe found.
func (r *Repository) Walk(f func(recipe *recipe.Recipe, err error) error) error {
	repo, err := git.PlainOpen(r.Path)
	if err != nil {
		return errors.Wrap(err, "cannot open repository")
	}

	idx, err := repo.Storer.Index()
	if err != nil {
		return errors.Wrap(err, "cannot get index")
	}

	for _, entry := range idx.Entries {
		if strings.HasSuffix(entry.Name, "/recipe.yaml") {
			recipe, err := r.Recipe(filepath.Base(filepath.Dir(entry.Name)))
			f(recipe, err)
		}
	}

	return nil
}
