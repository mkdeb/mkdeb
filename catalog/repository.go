package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"mkdeb.sh/recipe"
)

// Repository is a recipes repository.
type Repository struct {
	Name   string
	URL    string
	Branch string
	Path   string
}

// NewRepository creates a new recipes repository instance.
func NewRepository(path, name, url, branch string) *Repository {
	return &Repository{
		Name:   name,
		URL:    url,
		Branch: branch,
		Path:   path,
	}
}

// NewRepositoryFromPath creates a new recipes repository instance from an existing path.
func NewRepositoryFromPath(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, fmt.Errorf("cannot get remote: %w", err)
	}
	remoteCfg := remote.Config()

	return &Repository{
		Name:   path[strings.Index(path, "repositories/")+13:],
		URL:    remoteCfg.URLs[0],
		Branch: filepath.Base(remoteCfg.Fetch[0].Src()),
		Path:   path,
	}, nil
}

// Recipe loads a recipe from the repository.
func (r *Repository) Recipe(name string) (*recipe.Recipe, error) {
	return recipe.LoadRecipe(filepath.Join(r.Path, string(name[0]), name))
}

// Update updates the repository from the remote.
func (r *Repository) Update(force bool) error {
	// Clone if repository doesn't exist
	_, err := os.Stat(r.Path)
	if os.IsNotExist(err) {
		_, err := git.PlainClone(r.Path, false, &git.CloneOptions{
			URL:           r.URL,
			ReferenceName: plumbing.ReferenceName("refs/heads/" + r.Branch),
			SingleBranch:  true,
			Depth:         1,
			Progress:      os.Stdout,
		})
		if err != nil {
			return fmt.Errorf("cannot clone repository: %w", err)
		}

		return nil
	}

	// Pull changes from remote
	repo, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("cannot open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("cannot get worktree: %w", err)
	}

	// TODO: implement force (git checkout -f && git clean -d -f)

	err = wt.Pull(&git.PullOptions{
		Depth:    1,
		Progress: os.Stdout,
	})
	if err == git.NoErrAlreadyUpToDate {
		return ErrAlreadyUpToDate
	} else if err != nil {
		return fmt.Errorf("cannot pull repository: %w", err)
	}

	return nil
}

// Walk walks the repository calling a function for each recipe found.
func (r *Repository) Walk(f func(recipe *recipe.Recipe, err error) error) error {
	repo, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("cannot open repository: %w", err)
	}

	idx, err := repo.Storer.Index()
	if err != nil {
		return fmt.Errorf("cannot get index: %w", err)
	}

	for _, entry := range idx.Entries {
		if strings.HasSuffix(entry.Name, "/recipe.yaml") {
			recipe, err := r.Recipe(filepath.Base(filepath.Dir(entry.Name)))
			if err != nil {
				return fmt.Errorf("cannot load %q recipe: %w", filepath.Join(r.Path, entry.Name), err)
			}

			err = f(recipe, err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
