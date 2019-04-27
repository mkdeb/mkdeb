package main

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"mkdeb.sh/cmd/mkdeb/internal/print"
	"mkdeb.sh/recipe"
	"mkdeb.sh/repository"
)

var updateCommand = cli.Command{
	Name:      "update",
	Usage:     "Update recipes repositories",
	Action:    execUpdate,
	ArgsUsage: " ",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force",
			Usage: "Force actions and repair any dangling state",
		},
		cli.BoolFlag{
			Name:  "index-only",
			Usage: "Only perform recipes indexing",
		},
	},
}

func execUpdate(ctx *cli.Context) error {
	var idx bleve.Index

	r := repository.NewRepository(repositoryDir)

	if !ctx.Bool("index-only") {
		// Update repository
		if !r.Exists() {
			print.Step("Initializing repository...")

			if err := r.Init(os.Stdout); err != nil {
				return nil
			}
		} else {
			print.Step("Updating repository...")

			if err := r.Update(os.Stdout, ctx.Bool("force")); err == repository.ErrAlreadyUpToDate {
				fmt.Println(err)
			} else if err != nil {
				return err
			}
		}
	}

	// Index recipes
	if ctx.Bool("force") {
		if err := os.RemoveAll(indexDir); err != nil {
			return errors.Wrap(err, "cannot reset index")
		}
	}

	print.Step("Indexing repository...")

	if _, err := os.Stat(indexDir); err == nil {
		idx, err = bleve.Open(indexDir)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		idx, err = bleve.New(indexDir, bleve.NewIndexMapping())
		if err != nil {
			return errors.Wrap(err, "cannot create index")
		}
	}
	defer idx.Close()

	if err := r.Walk(func(recipe *recipe.Recipe, err error) error {
		if err != nil {
			return err
		}

		idx.Index(recipe.Name, indexRecord{
			Name:        recipe.Name,
			Description: recipe.Description,
		})

		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk repository")
	}

	count, _ := idx.DocCount()

	message.Set(language.English, "update.result", plural.Selectf(1, "%d",
		plural.One, "indexed %d recipe",
		plural.Other, "indexed %d recipes",
	))

	print.Summary("📋", message.NewPrinter(language.English).Sprintf("update.result", count))

	return nil
}

type indexRecord struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
