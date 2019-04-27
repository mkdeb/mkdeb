package catalog

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/token/lowercase"
	"github.com/blevesearch/bleve/analysis/tokenizer/single"
	"github.com/blevesearch/bleve/search/query"
	"github.com/pkg/errors"
	"mkdeb.sh/recipe"
)

const (
	// DefaultRepository is the default recipes repository short name.
	DefaultRepository = "mkdeb/core"

	reqSize = 100000 // FIXME: find a better way to handle request size
)

// Catalog is a recipes catalog.
type Catalog struct {
	Path  string
	index bleve.Index
}

// New creates a new recipes catalog instance.
func New(path string) (*Catalog, error) {
	var err error

	c := &Catalog{
		Path: path,
	}

	idxDir := filepath.Join(c.Path, "index")

	_, err = os.Stat(idxDir)
	if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		mapping.AddCustomAnalyzer("custom", map[string]interface{}{
			"type":      "custom",
			"tokenizer": single.Name,
			"token_filters": []string{
				lowercase.Name,
			},
		})
		mapping.DefaultAnalyzer = "custom"

		c.index, err = bleve.New(idxDir, mapping)
	} else {
		c.index, err = bleve.Open(idxDir)
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot open index")
	}

	return c, nil
}

// Close closes the catalog index.
func (c *Catalog) Close() error {
	if c.index != nil {
		return c.index.Close()
	}
	return nil
}

// Index indexes repositories recipes into the catalog.
func (c *Catalog) Index() (uint64, error) {
	err := c.Walk(func(recipe *recipe.Recipe, repo *Repository, err error) error {
		if err != nil {
			return err
		}

		return c.index.Index(repo.Name+"/"+recipe.Name, SearchHit{
			Name:        recipe.Name,
			Description: recipe.Description,
			Repository:  repo.Name,
		})
	})
	if err != nil {
		return 0, err
	}

	return c.index.DocCount()
}

// InstallRepository installs a new recipes repository into the catalog.
func (c *Catalog) InstallRepository(name, url, branch string, force bool) (uint64, error) {
	var count uint64

	path := filepath.Join(c.Path, "repositories", name)

	if !force {
		_, err := os.Stat(path)
		if err == nil {
			return 0, ErrRepositoryExist
		}
	}

	repo := NewRepository(path, name, url, branch)

	err := repo.Update(force)
	if err != nil {
		return 0, err
	}

	// Walk repository to perform a batch index
	batch := c.index.NewBatch()

	err = repo.Walk(func(recipe *recipe.Recipe, err error) error {
		if err != nil {
			return err
		}

		count++

		return batch.Index(repo.Name+"/"+recipe.Name, SearchHit{
			Name:        recipe.Name,
			Description: recipe.Description,
			Repository:  repo.Name,
		})
	})
	if err != nil {
		return 0, errors.Wrap(err, "cannot index repository")
	}

	err = c.index.Batch(batch)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Repositories returns a list of installed recipes repositories.
func (c *Catalog) Repositories() ([]*Repository, error) {
	path := filepath.Join(c.Path, "repositories")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	namespaces, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := []*Repository{}
	for _, namespace := range namespaces {
		if namespace.IsDir() {
			files, err := ioutil.ReadDir(filepath.Join(path, namespace.Name()))
			if err != nil {
				return nil, err
			}

			for _, file := range files {
				if file.IsDir() {
					repo, err := NewRepositoryFromPath(filepath.Join(path, namespace.Name(), file.Name()))
					if err != nil {
						return nil, err
					}
					result = append(result, repo)
				}
			}
		}
	}

	return result, nil
}

// Recipe searches the catalog for a recipe.
func (c *Catalog) Recipe(name string) (*recipe.Recipe, error) {
	var (
		q    query.Query
		path string
	)

	if strings.Contains(name, "/") {
		q = bleve.NewDocIDQuery([]string{name})
	} else {
		wq := bleve.NewTermQuery(name)
		wq.SetField("Name")

		q = wq
	}

	req := bleve.NewSearchRequest(q)
	req.Size = reqSize

	result, err := c.index.Search(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot search index")
	}

	if result.Total == 0 {
		return nil, ErrRecipeNotFound
	}

	// Search for default repository on duplicate recipes short names
	if result.Total > 1 {
		for _, hit := range result.Hits {
			if strings.HasPrefix(hit.ID, DefaultRepository+"/") {
				path = hit.ID
				break
			}
		}
	}

	if path == "" {
		path = result.Hits[0].ID
	}

	dir, base := filepath.Dir(path), filepath.Base(path)
	return recipe.LoadRecipe(filepath.Join(c.Path, "repositories", dir, string(base[0]), base))
}

// UninstallRepository uninstalls a recipes repository from the catalog.
func (c *Catalog) UninstallRepository(name string) (uint64, error) {
	var count uint64

	path := filepath.Join(c.Path, "repositories", name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, ErrRepositoryNotExist
	}

	req := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	req.Size = reqSize

	result, err := c.index.Search(req)
	if err != nil {
		return 0, err
	}

	// Walk index to delete matching recipes
	batch := c.index.NewBatch()

	for _, hit := range result.Hits {
		if strings.HasPrefix(hit.ID, name+"/") {
			count++
			batch.Delete(hit.ID)
		}
	}

	err = c.index.Batch(batch)
	if err != nil {
		return 0, err
	}

	err = os.RemoveAll(path)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Walk walks the catalog calling a function for each recipe found.
func (c *Catalog) Walk(f func(recipe *recipe.Recipe, repo *Repository, err error) error) error {
	repos, err := c.Repositories()
	if err != nil {
		return err
	}

	for _, repo := range repos {
		err = repo.Walk(func(recipe *recipe.Recipe, err error) error {
			return f(recipe, repo, err)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Ready returns whether or not the catalog is ready for use.
func Ready(path string) bool {
	_, err := os.Stat(filepath.Join(path, "repositories"))
	return err == nil
}
