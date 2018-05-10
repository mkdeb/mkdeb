package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"mkdeb.sh/cmd/mkdeb/internal/columns"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/config"
	"github.com/blevesearch/bleve/search/highlight/format/ansi"
	"github.com/blevesearch/bleve/search/query"
	"github.com/facette/natsort"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"mkdeb.sh/repository"
)

var searchCommand = cli.Command{
	Name:      "search",
	Usage:     "search for recipes",
	Action:    execSearch,
	ArgsUsage: "term",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "desc",
			Usage: "include recipes description when searching",
		},
	},
}

func execSearch(ctx *cli.Context) error {
	var q *query.WildcardQuery

	desc := ctx.Bool("desc")

	if ctx.NArg() > 1 {
		cli.ShowCommandHelpAndExit(ctx, "search", 1)
	}

	if r := repository.NewRepository(repositoryDir); !r.Exists() {
		if err := ctx.App.Run([]string{ctx.App.Name, "update"}); err != nil {
			return errors.Wrap(err, "failed to initialize repository")
		}
	}

	idx, err := bleve.Open(indexDir)
	if err != nil {
		return err
	}

	if ctx.NArg() == 0 {
		q = bleve.NewWildcardQuery("*")
	} else {
		q = bleve.NewWildcardQuery("*" + ctx.Args().First() + "*")
	}
	if !desc {
		q.SetField("name")
	}

	req := bleve.NewSearchRequest(q)
	if desc {
		color := ansi.Underscore
		if ctx.NArg() == 0 {
			color = ansi.Reset
		}

		bleve.Config.Cache.DefineFragmentFormatter("custom", map[string]interface{}{
			"type":  "ansi",
			"color": color,
		})

		bleve.Config.Cache.DefineHighlighter("custom", map[string]interface{}{
			"type":       "simple",
			"fragmenter": "simple",
			"formatter":  "custom",
		})

		req.Highlight = bleve.NewHighlightWithStyle("custom")
		req.Highlight.AddField("description")
	}

	result, err := idx.Search(req)
	if err != nil {
		return errors.Wrap(err, "failed to search in index")
	}

	if result.Total > 0 {
		items := make([]string, result.Total)
		for i, hit := range result.Hits {
			if !desc {
				items[i] = hit.ID
			} else {
				items[i] = hit.ID + "\t" + hit.Fragments["description"][0]
			}
		}
		natsort.Sort(items)

		if !desc {
			columns.Print(items, 2)
		} else {
			tr := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			for _, item := range items {
				tr.Write([]byte(item + "\n"))
			}
			tr.Flush()
		}
	} else {
		fmt.Println("no match found")
	}

	return nil
}
