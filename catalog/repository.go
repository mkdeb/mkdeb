package catalog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
		return nil, errors.Wrap(err, "cannot open repository")
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get remote")
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
			Progress:      os.Stdout,
		})
		if err != nil {
			return errors.Wrap(err, "cannot clone repository")
		}

		return nil
	}

	// Pull changes from remote
	repo, err := git.PlainOpen(r.Path)
	if err != nil {
		return errors.Wrap(err, "cannot open repository")
	}

	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "cannot get worktree")
	}

	// TODO: implement force (git checkout -f && git clean -d -f)

	err = wt.Pull(&git.PullOptions{
		Progress: os.Stdout,
	})
	if err == git.NoErrAlreadyUpToDate {
		return ErrAlreadyUpToDate
	} else if err != nil {
		return errors.Wrap(err, "cannot pull repository")
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
