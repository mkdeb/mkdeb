package catalog

import (
	"fmt"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

// SearchHit is a catalog recipe search hit.
type SearchHit struct {
	Name        string
	Description string
	Repository  string
}

// Search searches the catalog index for recipes matches.
func (c *Catalog) Search(term string, includeDesc bool) ([]*SearchHit, error) {
	var q query.Query

	if term != "" {
		wq := bleve.NewWildcardQuery("*" + strings.ToLower(term) + "*")
		if !includeDesc {
			wq.SetField("Name")
		}

		q = wq
	} else {
		q = bleve.NewMatchAllQuery()
	}

	req := bleve.NewSearchRequest(q)
	req.Fields = []string{"Name", "Description", "Repository"}
	req.Size = reqSize

	result, err := c.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("cannot search index: %w", err)
	}

	recipes := []*SearchHit{}
	for _, hit := range result.Hits {
		recipes = append(recipes, &SearchHit{
			Name:        hitField(hit.Fields, "Name"),
			Description: hitField(hit.Fields, "Description"),
			Repository:  hitField(hit.Fields, "Repository"),
		})
	}

	return recipes, nil
}

func hitField(fields map[string]interface{}, key string) string {
	v, _ := fields[key].(string)
	return v
}
